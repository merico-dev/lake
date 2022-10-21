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
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
	"gorm.io/gorm"
	"reflect"
)

// ExtractorArgs args to initialize a Extractor
type ExtractorArgs[Record any] struct {
	Ctx core.SubTaskContext
	// The function that creates and returns a tap client
	TapProvider func() (Tap, errors.Error)
	// The specific tap stream to invoke at runtime
	StreamName   string
	ConnectionId uint64
	Extract      func(*Record) ([]interface{}, errors.Error)
}

// Extractor the extractor that communicates with singer taps
type Extractor[Record any] struct {
	*ExtractorArgs[Record]
	tap           Tap
	streamVersion uint64
}

// NewTapExtractor constructor for Extractor
func NewTapExtractor[Record any](args *ExtractorArgs[Record]) (*Extractor[Record], errors.Error) {
	tapClient, err := args.TapProvider()
	if err != nil {
		return nil, err
	}
	extractor := &Extractor[Record]{
		ExtractorArgs: args,
		tap:           tapClient,
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

func (e *Extractor[Record]) getState() (*State, errors.Error) {
	db := e.Ctx.GetDal()
	rawState := RawState{
		Id: e.getStateId(),
	}
	if err := db.First(&rawState); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound.Wrap(err, "record not found")
		}
		return nil, err
	}
	return ToState(&rawState), nil
}

func (e *Extractor[Record]) pushState(state *State) errors.Error {
	db := e.Ctx.GetDal()
	rawState := FromState(e.ConnectionId, state)
	rawState.Id = e.getStateId()
	return db.CreateOrUpdate(rawState)
}

func (e *Extractor[Record]) getStateId() string {
	return fmt.Sprintf("{%s:%d:%d}", fmt.Sprintf("%s::%s", e.tap.GetName(), e.StreamName), e.ConnectionId, e.streamVersion)
}

// Execute executes the extractor
func (e *Extractor[Record]) Execute() (err errors.Error) {
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
	resultStream, err := e.tap.Run()
	if err != nil {
		return err
	}
	e.Ctx.SetProgress(0, -1)
	ctx := e.Ctx.GetContext()
	var batchedResults []interface{}
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
			var extractedResults []interface{}
			extractedResults, err = e.Extract(tapRecord.Record)
			if err != nil {
				return err
			}
			batchedResults = append(batchedResults, extractedResults...)
			e.Ctx.IncProgress(1)
			continue
		} else if tapState, ok := AsTapState(result.Data); ok {
			err = e.pushResults(batchedResults)
			if err != nil {
				return err
			}
			err = e.pushState(tapState)
			if err != nil {
				return errors.Default.Wrap(err, "error saving tap state")
			}
			batchedResults = nil
			continue
		}
	}
	return nil
}

func (e *Extractor[Record]) pushResults(results []any) errors.Error {
	if len(results) == 0 {
		return nil
	}
	e.Ctx.GetLogger().Info("%s flushing %d records", e.tap.GetName(), len(results))
	divider := helper.NewNonRawBatchSaveDivider(e.Ctx, 1+len(results))
	for _, result := range results {
		batch, err := divider.ForType(reflect.TypeOf(result))
		if err != nil {
			return err
		}
		err = batch.Add(result)
		if err != nil {
			return err
		}
	}
	err := divider.Close()
	if err != nil {
		return errors.Default.Wrap(err, "error flushing tap records to DB")
	}
	return nil
}

var _ core.SubTask = (*Extractor[any])(nil)
