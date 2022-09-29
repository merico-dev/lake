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
	"github.com/mitchellh/hashstructure"
	"gorm.io/gorm"
	"reflect"
	"time"
)

// SingerExtractorArgs add doc
type SingerExtractorArgs[Record, State, Schema any] struct {
	Ctx             core.SubTaskContext
	SingerConfig    interface{}
	ConnectionId    uint64
	Extract         func(*Record) ([]interface{}, errors.Error)
	BatchSize       int
	TapType         string
	TapClass        string
	TapSchemaSetter func(stream *singer.Stream[Schema]) bool
	TapStateSetter  func(time time.Time) *State
}

// SingerApiExtractor add doc
type SingerApiExtractor[Record, State, Schema any] struct {
	args          *SingerExtractorArgs[Record, State, Schema]
	tap           *singer.Tap[Schema]
	streamVersion uint64
}

// NewSingerApiExtractor add doc
func NewSingerApiExtractor[Record, State, Schema any](args *SingerExtractorArgs[Record, State, Schema]) *SingerApiExtractor[Record, State, Schema] {
	env := config.GetConfig()
	cmd := env.GetString(args.TapClass)
	if cmd == "" {
		panic("singer tap command not provided")
	}
	tap := singer.NewTap[Schema](&singer.Config{
		Mappings: args.SingerConfig,
		Cmd:      cmd,
		TapType:  args.TapType,
	})
	extractor := &SingerApiExtractor[Record, State, Schema]{
		args: args,
		tap:  tap,
	}
	extractor.setupSingerTap()
	return extractor
}

func (e *SingerApiExtractor[Record, State, Schema]) setupSingerTap() {
	e.tap.WriteConfig()
	e.tap.DiscoverProperties()
	e.tap.SetProperties(func(s *singer.Stream[Schema]) bool {
		b := e.args.TapSchemaSetter(s)
		e.streamVersion = hash(s)
		return b
	})
}

// TODO fix this...
func (e *SingerApiExtractor[Record, State, Schema]) getState() (*singer.TapState[State], errors.Error) {
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
func (e *SingerApiExtractor[Record, State, Schema]) pushState(state *singer.TapState[State]) errors.Error {
	db := e.args.Ctx.GetDal()
	rawState := singer.FromState[State](e.args.ConnectionId, state)
	rawState.Id = e.getStateId()
	return db.CreateOrUpdate(rawState)
}

func (e *SingerApiExtractor[Record, State, Schema]) getStateId() string {
	//TODO should this account for the schema state?
	return fmt.Sprintf("{singer:%d:%s:%d}", e.args.ConnectionId, e.args.TapType, e.streamVersion)
}

// Execute add doc
func (e *SingerApiExtractor[Record, State, Schema]) Execute() (err errors.Error) {
	initialState, err := e.getState()
	if err != nil && err.GetType() != errors.NotFound {
		return err
	}
	if initialState != nil {
		e.tap.WriteState(initialState.Value)
	}
	divider := NewBatchSaveDivider2(e.args.Ctx, e.args.BatchSize)
	var state *singer.TapState[State]
	defer func() {
		err = divider.Close()
		if err == nil && state != nil {
			// save the last state we got in the DB
			if err = e.pushState(state); err != nil {
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
			if e.args.TapStateSetter != nil {
				state = singer.NewTapState(e.args.TapStateSetter(tapRecord.TimeExtracted))
			}
			e.args.Ctx.IncProgress(1)
			continue
		} else if state, ok = singer.AsTapState[State](d.Data); ok {
			continue
		}
	}
	return nil
}

func (e *SingerApiExtractor[Record, State, Schema]) pushToDB(divider *BatchSaveDivider, results []any) errors.Error {
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

func hash(x any) uint64 {
	version, err := hashstructure.Hash(x, nil)
	if err != nil {
		panic(err)
	}
	return version
}
