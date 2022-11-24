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
	"fmt"
	"github.com/apache/incubator-devlake/models"
	"reflect"
	"strings"
)

func LoadTableModel(conn map[string]any) *models.DynamicTabler {
	table := ""
	if tableRaw, ok := conn["table"]; ok {
		delete(conn, "table")
		table = tableRaw.(string)
	} else {
		panic("table field not provided")
	}
	structType := LoadModelType(conn)
	return models.NewDynamicTabler(table, structType)
}

func LoadModelType(conn map[string]any) reflect.Type {
	var structFields []reflect.StructField
	for k, v := range conn {
		sf := transformFields(k, v)
		structFields = append(structFields, *sf)
	}
	return reflect.StructOf(structFields)
}

// nolint:staticcheck
func transformFields(k string, v any) *reflect.StructField {
	sf := reflect.StructField{
		Name: strings.Title(k),
		Type: reflect.TypeOf(v),
	}
	if m, ok := v.(map[string]any); ok {
		if tags, ok := m["tags"]; ok {
			sf.Tag = reflect.StructTag(tags.(string))
			value := m["value"]
			sf.Type = reflect.TypeOf(value)
		}
	}
	jsonTag := reflect.StructTag(fmt.Sprintf("json:\"%s\"", k))
	if sf.Tag == "" {
		sf.Tag = jsonTag
	} else {
		sf.Tag = reflect.StructTag(fmt.Sprintf("%s %s", sf.Tag, jsonTag))
	}
	if typeName := sf.Tag.Get("gotype"); typeName != "" {
		if cached, ok := supportedTypes[typeName]; ok {
			sf.Type = reflect.TypeOf(cached)
		} else {
			panic(fmt.Sprintf("unsupported gotype: %s on field: %s", typeName, k))
		}
	}
	return &sf
}

// keep around for if needed later (for any default fields to inherit)
// nolint:unused
func inheritFields(parent any) (fields []reflect.StructField) {
	var f func(rtype reflect.Type)
	f = func(rtype reflect.Type) {
		for i := 0; i < rtype.NumField(); i++ {
			sf := rtype.Field(i)
			if sf.Anonymous {
				f(sf.Type)
			} else {
				fields = append(fields, sf)
			}
		}
	}
	f(reflect.TypeOf(parent))
	return fields
}
