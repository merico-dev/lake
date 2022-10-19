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
	goerror "errors"
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
	"gorm.io/gorm"
	"reflect"
)

// TapExtractorArgs args to initialize a TapExtractor
type TapExtractorArgs[Record any] struct {
	Ctx core.SubTaskContext
	// The function that creates and returns a tap client
	TapProvider func() (Tap, errors.Error)
	// The specific tap stream to invoke at runtime
	StreamName   string
	ConnectionId uint64
	Extract      func(*Record) ([]interface{}, errors.Error)
	BatchSize    int
}

// TapExtractor the extractor that communicates with singer taps
type TapExtractor[Record any] struct {
	*TapExtractorArgs[Record]
	tap           Tap
	streamVersion uint64
}

// NewTapExtractor constructor for TapExtractor
func NewTapExtractor[Record any](args *TapExtractorArgs[Record]) (*TapExtractor[Record], errors.Error) {
	tapClient, err := args.TapProvider()
	if err != nil {
		return nil, err
	}
	extractor := &TapExtractor[Record]{
		TapExtractorArgs: args,
		tap:              tapClient,
	}
	err = extractor.tap.SetConfig()
	if err != nil {
		return nil, err
	}
	extractor.streamVersion, err = extractor.tap.SetProperties(args.StreamName)
	if err != nil {
		return nil, err
	}
	return extractor, nil
}

func (e *TapExtractor[Record]) getState() (*State, errors.Error) {
	db := e.Ctx.GetDal()
	rawState := RawState{
		Id: e.getStateId(),
	}
	if err := db.First(&rawState); err != nil {
		if goerror.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound.Wrap(err, "record not found")
		}
		return nil, err
	}
	return ToState(&rawState), nil
}

func (e *TapExtractor[Record]) pushState(state *State) errors.Error {
	db := e.Ctx.GetDal()
	rawState := FromState(e.ConnectionId, state)
	rawState.Id = e.getStateId()
	return db.CreateOrUpdate(rawState)
}

func (e *TapExtractor[Record]) getStateId() string {
	return fmt.Sprintf("{%s:%d:%d}", fmt.Sprintf("%s::%s", e.tap.GetName(), e.StreamName), e.ConnectionId, e.streamVersion)
}

func (e *TapExtractor[Record]) close(state *State, divider *helper.BatchSaveDivider) errors.Error {
	err := divider.Close()
	if err == nil && state.Value != nil {
		// save the last state we got in the DB
		if err = e.pushState(state); err != nil {
			err = errors.Default.Wrap(err, "error storing state for retrieved singer tap records")
		}
	}
	return err
}

// Execute executes the extractor
func (e *TapExtractor[Record]) Execute() (err errors.Error) {
	initialState, err := e.getState()
	if err != nil && err.GetType() != errors.NotFound {
		return err
	}
	if initialState != nil {
		err = e.tap.SetState(initialState.Value)
		if err != nil {
			return err
		}
	}
	divider := helper.NewNonRawBatchSaveDivider(e.Ctx, e.BatchSize)
	var state State
	recordsProcessed := 0
	defer func() {
		e.Ctx.GetLogger().Info("%s processed %d records", e.tap.GetName(), recordsProcessed)
		err2 := e.close(&state, divider)
		if err2 != nil {
			e.Ctx.GetLogger().Error(err2, "error closing tap executor")
		}
	}()
	resultStream, err := e.tap.Run()
	if err != nil {
		return err
	}
	e.Ctx.SetProgress(0, -1)
	ctx := e.Ctx.GetContext()
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
		if tapRecord, ok := AsTapRecord[Record](result.Data); ok {
			var results []interface{}
			results, err = e.Extract(tapRecord.Record)
			if err != nil {
				return err
			}
			if err = e.pushToDB(divider, results); err != nil {
				return err
			}
			e.Ctx.IncProgress(1)
			recordsProcessed++
			continue
		} else if tapState, ok := AsTapState(result.Data); ok {
			state = *tapState
			continue
		}
	}
	return nil
}

func (e *TapExtractor[Record]) pushToDB(divider *helper.BatchSaveDivider, results []any) errors.Error {
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
		e.Ctx.IncProgress(1)
	}
	return nil
}

var _ core.SubTask = (*TapExtractor[any])(nil)
