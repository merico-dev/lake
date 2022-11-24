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

package models

import (
	"github.com/apache/incubator-devlake/plugins/core"
	"reflect"
)

// DynamicTabler is a core.Tabler that wraps a runtime (anonymously) generated data-model. Due to limitations of
// reflection in Go and the GORM framework, the underlying model and the table have to be explicitly passed into dal.Dal's API
// via Unwrap() and TableName()
type DynamicTabler struct {
	objType reflect.Type
	wrapped any
	table   string
}

func NewDynamicTabler(tableName string, objType reflect.Type) *DynamicTabler {
	return &DynamicTabler{
		objType: objType,
		table:   tableName,
	}
}

func (d *DynamicTabler) New() *DynamicTabler {
	return &DynamicTabler{
		objType: d.objType,
		wrapped: reflect.New(d.objType).Interface(),
		table:   d.table,
	}
}

func (d *DynamicTabler) NewSlice() *DynamicTabler {
	sliceType := reflect.SliceOf(d.objType)
	return &DynamicTabler{
		objType: sliceType,
		wrapped: reflect.New(sliceType).Interface(),
		table:   d.table,
	}
}

func (d *DynamicTabler) Unwrap() any {
	return d.wrapped
}

func (d *DynamicTabler) TableName() string {
	return d.table
}

var _ core.Tabler = (*DynamicTabler)(nil)
