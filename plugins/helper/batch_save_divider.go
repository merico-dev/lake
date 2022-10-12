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
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"reflect"

	"github.com/apache/incubator-devlake/models/common"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
)

// BatchSaveDivider creates and caches BatchSave, this is helpful when dealing with massive amount of data records
// with arbitrary types.
type BatchSaveDivider struct {
	basicRes  core.BasicRes
	log       core.Logger
	db        dal.Dal
	batches   map[reflect.Type]*BatchSave
	batchSize int
	rawTable  string
	rawParams string
	isRaw     bool
}

// NewBatchSaveDivider create a new BatchInsertDivider instance. The table param cannot be empty.
func NewBatchSaveDivider(basicRes core.BasicRes, batchSize int, table string, params string) *BatchSaveDivider {
	log := basicRes.GetLogger().Nested("batch divider")
	return &BatchSaveDivider{
		basicRes:  basicRes,
		log:       log,
		db:        basicRes.GetDal(),
		batches:   make(map[reflect.Type]*BatchSave),
		batchSize: batchSize,
		rawTable:  table,
		rawParams: params,
		isRaw:     true,
	}
}

// NewNonRawBatchSaveDivider create a new BatchInsertDivider instance with no raw associations
func NewNonRawBatchSaveDivider(basicRes core.BasicRes, batchSize int) *BatchSaveDivider {
	log := basicRes.GetLogger().Nested("batch divider")
	return &BatchSaveDivider{
		basicRes:  basicRes,
		log:       log,
		db:        basicRes.GetDal(),
		batches:   make(map[reflect.Type]*BatchSave),
		batchSize: batchSize,
	}
}

// ForType returns a `BatchSave` instance for specific type
func (d *BatchSaveDivider) ForType(rowType reflect.Type) (*BatchSave, errors.Error) {
	// get the cache for the specific type
	batch := d.batches[rowType]
	var err errors.Error
	// create one if not exists
	if batch == nil {
		batch, err = NewBatchSave(d.basicRes, rowType, d.batchSize)
		if err != nil {
			return nil, err
		}
		d.batches[rowType] = batch
		if d.isRaw {
			if err = d.handleRawOrigin(rowType); err != nil {
				return nil, err
			}
		}
	}
	return batch, nil
}

func (d *BatchSaveDivider) handleRawOrigin(rowType reflect.Type) errors.Error {
	// delete outdated records if rowType was not PartialUpdate
	rowElemType := rowType.Elem()
	row := reflect.New(rowElemType).Interface()
	field, hasField := rowElemType.FieldByName("RawDataOrigin")
	if !hasField || field.Type != reflect.TypeOf(common.RawDataOrigin{}) {
		return errors.Default.New(fmt.Sprintf("type %s must have RawDataOrigin embeded", rowElemType.Name()))
	}
	// all good, delete outdated records before we insertion
	d.log.Debug("deleting outdated records for %s", rowElemType.Name())
	err := d.db.Delete(
		row,
		dal.Where("_raw_data_table = ? AND _raw_data_params = ?", d.rawTable, d.rawParams),
	)
	return err
}

// Close all batches so the rest records get saved into db
func (d *BatchSaveDivider) Close() errors.Error {
	for _, batch := range d.batches {
		err := batch.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
