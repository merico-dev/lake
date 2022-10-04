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

package helper

import (
	goerror "errors"
	"fmt"
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/singer"
	"gorm.io/gorm"
	"reflect"
)

// SingerExtractorArgs add doc
type SingerExtractorArgs[Record any] struct {
	Ctx                  core.SubTaskContext
	SingerConfig         interface{}
	ConnectionId         uint64
	Extract              func(*Record) ([]interface{}, errors.Error)
	BatchSize            int
	TapType              string
	TapClass             string
	TapSchemaSetter      func(stream *singer.Stream) bool
	StreamPropertiesFile string
}

// SingerApiExtractor add doc
type SingerApiExtractor[Record, State any] struct {
	args          *SingerExtractorArgs[Record]
	tap           *singer.Tap
	streamVersion uint64
}

// NewSingerApiExtractor add doc
func NewSingerApiExtractor[Record any](args *SingerExtractorArgs[Record]) *SingerApiExtractor[Record, singer.TapStateValue] {
	env := config.GetConfig()
	cmd := env.GetString(args.TapClass)
	if cmd == "" {
		panic("singer tap command not provided")
	}
	tap := singer.NewTap(&singer.Config{
		Mappings:             args.SingerConfig,
		Cmd:                  cmd,
		TapType:              args.TapType,
		StreamPropertiesFile: args.StreamPropertiesFile,
	})
	extractor := &SingerApiExtractor[Record, singer.TapStateValue]{
		args: args,
		tap:  tap,
	}
	extractor.setupSingerTap()
	return extractor
}

func (e *SingerApiExtractor[Record, State]) setupSingerTap() {
	e.tap.WriteConfig()
	e.streamVersion = e.tap.SetProperties(e.args.TapSchemaSetter)
}

// TODO fix this...
func (e *SingerApiExtractor[Record, State]) getState() (*singer.TapState[State], errors.Error) {
	db := e.args.Ctx.GetDal()
	rawState := singer.RawTapState{
		Id: e.getStateId(),
	}
	if err := db.First(&rawState); err != nil {
		if goerror.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound.Wrap(err, "record not found")
		}
		return nil, err
	}
	return singer.ToState[State](&rawState), nil
}

// TODO fix this...
func (e *SingerApiExtractor[Record, State]) pushState(state *singer.TapState[State]) errors.Error {
	db := e.args.Ctx.GetDal()
	rawState := singer.FromState[State](e.args.ConnectionId, state)
	rawState.Id = e.getStateId()
	return db.CreateOrUpdate(rawState)
}

func (e *SingerApiExtractor[Record, State]) getStateId() string {
	//TODO should this account for the schema state?
	return fmt.Sprintf("{singer:%d:%s:%d}", e.args.ConnectionId, e.args.TapType, e.streamVersion)
}

// Execute add doc
func (e *SingerApiExtractor[Record, State]) Execute() (err errors.Error) {
	initialState, err := e.getState()
	if err != nil && err.GetType() != errors.NotFound {
		return err
	}
	if initialState != nil {
		e.tap.WriteState(initialState.Value)
	}
	divider := NewBatchSaveDivider2(e.args.Ctx, e.args.BatchSize)
	var state singer.TapState[State]
	recordsProcessed := 0
	defer func() {
		e.args.Ctx.GetLogger().Info("%s processed %d records", e.args.TapClass, recordsProcessed)
		err = divider.Close()
		if err == nil && state.Value != nil {
			// save the last state we got in the DB
			if err = e.pushState(&state); err != nil {
				err = errors.Default.Wrap(err, "error storing state for retrieved singer tap records")
			}
		}
	}()
	stream := e.tap.Run()
	e.args.Ctx.SetProgress(0, -1)
	ctx := e.args.Ctx.GetContext()
	for d := range stream {
		if d.Err != nil {
			panic(d.Err)
		}
		select {
		case <-ctx.Done():
			return errors.Convert(ctx.Err())
		default:
		}
		if tapRecord, ok := singer.AsTapRecord[Record](d.Data); ok {
			var results []interface{}
			results, err = e.args.Extract(tapRecord.Record)
			if err != nil {
				return err
			}
			if err = e.pushToDB(divider, results); err != nil {
				return err
			}
			e.args.Ctx.IncProgress(1)
			recordsProcessed++
			continue
		} else if tapState, ok := singer.AsTapState[State](d.Data); ok {
			e.args.Ctx.GetLogger().Info("state: %v", tapState.Value)
			state = *tapState
			continue
		}
	}
	return nil
}

func (e *SingerApiExtractor[Record, State]) pushToDB(divider *BatchSaveDivider, results []any) errors.Error {
	for _, result := range results {
		// get the batch operator for the specific type
		batch, err := divider.ForType(reflect.TypeOf(result))
		if err != nil {
			return errors.Default.Wrap(err, "error getting batch from result")
		}
		err = batch.Add(result)
		if err != nil {
			return errors.Default.Wrap(err, "error adding result to batch")
		}
		e.args.Ctx.IncProgress(1)
	}
	return nil
}
