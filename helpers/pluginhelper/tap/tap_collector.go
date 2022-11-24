/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tap

import (
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"github.com/apache/incubator-devlake/plugins/helper"
	"gorm.io/gorm"
)

// CollectorArgs args to initialize a Collector
type CollectorArgs[Stream any] struct {
	helper.RawDataSubTaskArgs
	// The function that creates and returns a tap client
	TapClient Tap[Stream]
	// Optional - This function is called for the selected streams at runtime. Use this if any runtime modification is needed.
	TapStreamModifier func(stream *Stream) bool
	// The config the tap needs at runtime in order to execute
	TapConfig any
	// The specific tap stream to invoke at runtime
	StreamName   string
	ConnectionId uint64
	Table        string
	Incremental  bool
}

// Collector the collector that communicates with singer taps
type Collector[Stream any] struct {
	ctx               core.SubTaskContext
	rawSubtask        *helper.RawDataSubTask
	tapClient         Tap[Stream]
	tapStreamModifier func(stream *Stream) bool
	tapConfig         any
	streamVersion     uint64
	streamName        string
	connectionId      uint64
	incremental       bool
}

// NewTapCollector constructor for Collector
func NewTapCollector[Stream any](args *CollectorArgs[Stream]) (*Collector[Stream], errors.Error) {
	rawDataSubTask, err := helper.NewRawDataSubTask(args.RawDataSubTaskArgs)
	if err != nil {
		return nil, err
	}
	collector := &Collector[Stream]{
		ctx:               args.Ctx,
		rawSubtask:        rawDataSubTask,
		tapClient:         args.TapClient,
		tapStreamModifier: args.TapStreamModifier,
		streamName:        args.StreamName,
		tapConfig:         args.TapConfig,
		connectionId:      args.ConnectionId,
		incremental:       args.Incremental,
	}
	if err = collector.prepareTap(); err != nil {
		return nil, err
	}
	return collector, nil
}

func (c *Collector[Stream]) getState() (*State, errors.Error) {
	db := c.ctx.GetDal()
	rawState := RawState{}
	rawState.ID = c.getStateId()
	if err := db.First(&rawState); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound.Wrap(err, "record not found")
		}
		return nil, err
	}
	return ToState(&rawState), nil
}

func (c *Collector[Stream]) pushState(state *State) errors.Error {
	db := c.ctx.GetDal()
	rawState := FromState(state)
	rawState.ID = c.getStateId()
	return db.CreateOrUpdate(rawState)
}

func (c *Collector[Stream]) getStateId() string {
	return fmt.Sprintf("{%s:%d:%d}", fmt.Sprintf("%s::%s", c.tapClient.GetName(), c.streamName), c.connectionId, c.streamVersion)
}

func (c *Collector[Stream]) prepareTap() errors.Error {
	if c.tapConfig == nil {
		return errors.Default.New("no tap config is set")
	}
	err := c.tapClient.SetConfig(c.tapConfig)
	if err != nil {
		return err
	}
	c.streamVersion, err = c.tapClient.SetProperties(c.streamName, c.tapStreamModifier)
	if err != nil {
		return err
	}
	return nil
}

// Execute executes the collector
func (c *Collector[Stream]) Execute() (err errors.Error) {
	initialState, err := c.getState()
	if err != nil && err.GetType() != errors.NotFound {
		return err
	}
	if initialState != nil {
		err = c.tapClient.SetState(initialState.Value)
		if err != nil {
			return err
		}
	}
	resultStream, err := c.tapClient.Run(c.ctx.GetContext())
	if err != nil {
		return err
	}
	err = c.prepareDB()
	if err != nil {
		return err
	}
	c.ctx.SetProgress(0, -1)
	ctx := c.ctx.GetContext()
	var batchedResults []json.RawMessage
	defer func() {
		if err == nil {
			// push whatever is left
			err = c.pushResults(batchedResults)
		}
	}()
	for result := range resultStream {
		if result.Err != nil {
			err = errors.Default.Wrap(result.Err, "error found in streamed tap result")
			return err
		}
		select {
		case <-ctx.Done():
			err = errors.Convert(ctx.Err())
			return err
		default:
		}
		output := result.Out
		if tapRecord, ok := output.AsTapRecord(); ok {
			batchedResults = append(batchedResults, tapRecord.Record)
			c.ctx.IncProgress(1)
			continue
		} else if tapState, ok := output.AsTapState(); ok {
			err = c.pushResults(batchedResults)
			if err != nil {
				return err
			}
			tapState.Type = fmt.Sprintf("TAP_%s", tapState.Type)
			err = c.pushState(tapState)
			if err != nil {
				return errors.Default.Wrap(err, "error saving tap state")
			}
			batchedResults = nil
			continue
		}
	}
	return nil
}

func (c *Collector[Stream]) pushResults(results []json.RawMessage) errors.Error {
	if len(results) == 0 {
		return nil
	}
	c.ctx.GetLogger().Info("%s flushing %d records", c.tapClient.GetName(), len(results))
	rows := make([]*helper.RawData, len(results))
	defaultInput, _ := json.Marshal(nil)
	for i, result := range results {
		rows[i] = &helper.RawData{
			Params: c.rawSubtask.GetParams(),
			Data:   result,
			Url:    "",           // n/a
			Input:  defaultInput, // n/a
		}
	}
	err := c.ctx.GetDal().Create(rows, dal.From(c.rawSubtask.GetTable()))
	if err != nil {
		return errors.Default.Wrap(err, "error pushing records to collector table")
	}
	return nil
}

func (c *Collector[Stream]) prepareDB() errors.Error {
	db := c.ctx.GetDal()
	err := db.AutoMigrate(&helper.RawData{}, dal.From(c.rawSubtask.GetTable()))
	if err != nil {
		return errors.Default.Wrap(err, "error auto-migrating collector")
	}
	if !c.incremental {
		err = c.ctx.GetDal().Delete(&helper.RawData{}, dal.From(c.rawSubtask.GetTable()), dal.Where("params = ?", c.rawSubtask.GetParams()))
		if err != nil {
			return errors.Default.Wrap(err, "error deleting data from collector")
		}
	}
	return nil
}

var _ core.SubTask = (*Collector[any])(nil)
