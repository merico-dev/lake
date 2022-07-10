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
	"reflect"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"github.com/apache/incubator-devlake/plugins/helper"
)

type store interface {
	findAll(interface{}) error
	save(items []interface{}) error
}

type dbStore struct {
	db     dal.Dal
	driver *helper.BatchSaveDivider
}

func NewDbStore(db dal.Dal, basicRes core.BasicRes) *dbStore {
	driver := helper.NewBatchSaveDivider(
		basicRes,
		1000,
		"",
		"",
	)
	return &dbStore{db: db, driver: driver}
}

func (d *dbStore) findAll(i interface{}) error {
	return d.db.All(i)
}

func (d *dbStore) save(items []interface{}) error {
	for _, item := range items {
		batch, err := d.driver.ForType(reflect.TypeOf(item))
		if err != nil {
			return err
		}
		err = batch.Add(item)
		if err != nil {
			return err
		}
	}
	d.driver.Close()
	return nil
}
