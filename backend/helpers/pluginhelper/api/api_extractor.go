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

package api

import (
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/common"
	plugin "github.com/apache/incubator-devlake/core/plugin"
	"reflect"
)

// ApiExtractorArgs FIXME ...
type ApiExtractorArgs struct {
	RawDataSubTaskArgs
	Params    interface{}
	Extract   func(row *RawData) ([]interface{}, errors.Error)
	BatchSize int
}

// ApiExtractor helps you extract Raw Data from api responses to Tool Layer Data
// It reads rows from specified raw data table, and feed it into `Extract` handler
// you can return arbitrary tool layer entities in this handler, ApiExtractor would
// first delete old data by their RawDataOrigin information, and then perform a
// batch save for you.
type ApiExtractor struct {
	*RawDataSubTask
	args *ApiExtractorArgs
}

// NewApiExtractor creates a new ApiExtractor
func NewApiExtractor(args ApiExtractorArgs) (*ApiExtractor, errors.Error) {
	// process args
	rawDataSubTask, err := NewRawDataSubTask(args.RawDataSubTaskArgs)
	if err != nil {
		return nil, err
	}
	if args.BatchSize == 0 {
		args.BatchSize = 500
	}
	return &ApiExtractor{
		RawDataSubTask: rawDataSubTask,
		args:           &args,
	}, nil
}

// Execute sub-task
func (extractor *ApiExtractor) Execute() errors.Error {
	// load data from database
	db := extractor.args.Ctx.GetDal()
	logger := extractor.args.Ctx.GetLogger()

	clauses := []dal.Clause{
		dal.From(extractor.table),
		dal.Where("params = ?", extractor.params),
		dal.Orderby("id ASC"),
	}

	count, err := db.Count(clauses...)
	if err != nil {
		return errors.Default.Wrap(err, "error getting count of clauses")
	}
	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return errors.Default.Wrap(err, "error running DB query")
	}
	logger.Info("get data from %s where params=%s and got %d", extractor.table, extractor.params, count)
	defer cursor.Close()
	row := &RawData{}

	// batch save divider
	RAW_DATA_ORIGIN := "RawDataOrigin"
	divider := NewBatchSaveDivider(extractor.args.Ctx, extractor.args.BatchSize, extractor.table, extractor.params)

	// prgress
	extractor.args.Ctx.SetProgress(0, -1)
	ctx := extractor.args.Ctx.GetContext()
	// iterate all rows
	for cursor.Next() {
		select {
		case <-ctx.Done():
			return errors.Convert(ctx.Err())
		default:
		}
		err = db.Fetch(cursor, row)
		if err != nil {
			return errors.Default.Wrap(err, "error fetching row")
		}

		results, err := extractor.args.Extract(row)
		if err != nil {
			return errors.Default.Wrap(err, "error calling plugin Extract implementation")
		}
		for _, result := range results {
			// get the batch operator for the specific type
			batch, err := divider.ForType(reflect.TypeOf(result))
			if err != nil {
				return errors.Default.Wrap(err, "error getting batch from result")
			}
			// set raw data origin field
			origin := reflect.ValueOf(result).Elem().FieldByName(RAW_DATA_ORIGIN)
			if origin.IsValid() && origin.IsZero() {
				origin.Set(reflect.ValueOf(common.RawDataOrigin{
					RawDataTable:  extractor.table,
					RawDataId:     row.ID,
					RawDataParams: row.Params,
				}))
			}
			// records get saved into db when slots were max outed
			err = batch.Add(result)
			if err != nil {
				return errors.Default.Wrap(err, "error adding result to batch")
			}
		}
		extractor.args.Ctx.IncProgress(1)
	}

	// save the last batches
	return divider.Close()
}

var _ plugin.SubTask = (*ApiExtractor)(nil)
