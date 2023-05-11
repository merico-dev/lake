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
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-devlake/impls/dalgorm"
	"reflect"
	"strings"
	"time"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models"
	"gorm.io/datatypes"
)

func LoadTableModel(tableName string, schema map[string]any, encrypt bool, parentModel any) (*models.DynamicTabler, errors.Error) {
	structType, err := GenerateStructType(schema, encrypt, reflect.TypeOf(parentModel))
	if err != nil {
		return nil, err
	}
	return models.NewDynamicTabler(tableName, structType), nil
}

func GenerateStructType(schema map[string]any, encrypt bool, baseType reflect.Type) (reflect.Type, errors.Error) {
	var structFields []reflect.StructField
	propsRaw, ok := schema["properties"]
	if !ok {
		return nil, errors.BadInput.New("Missing properties in JSON schema")
	}
	props, ok := propsRaw.(map[string]any)
	if !ok {
		return nil, errors.BadInput.New("JSON schema properties must be an object")
	}
	if baseType != nil {
		anonymousField := reflect.StructField{
			Name:      baseType.Name(),
			Type:      baseType,
			Tag:       reflect.StructTag("mapstructure:\",squash\""),
			Anonymous: true,
		}
		structFields = append(structFields, anonymousField)
	}
	for k, v := range props {
		if isBaseTypeField(k, baseType) {
			continue
		}
		spec := v.(map[string]any)
		field, err := generateStructField(k, encrypt, spec)
		if err != nil {
			return nil, err
		}
		structFields = append(structFields, *field)
	}
	return reflect.StructOf(structFields), nil
}

func MapTo(x any, y any) errors.Error {
	b, err := json.Marshal(x)
	if err != nil {
		return errors.Convert(err)
	}
	if err = json.Unmarshal(b, y); err != nil {
		return errors.Convert(err)
	}
	return nil
}

func ToDatabaseMap(tableName string, ifc any, createdAt *time.Time, updatedAt *time.Time) (map[string]any, errors.Error) {
	m := map[string]any{}
	err := MapTo(ifc, &m)
	if err != nil {
		return nil, err
	}
	if createdAt != nil {
		m["createdAt"] = createdAt
	}
	if updatedAt != nil {
		m["updatedAt"] = updatedAt
	}
	m = dalgorm.ToDatabaseMap(tableName, m)
	return m, nil
}

func isBaseTypeField(fieldName string, baseType reflect.Type) bool {
	fieldName = canonicalFieldName(fieldName)
	for i := 0; i < baseType.NumField(); i++ {
		baseField := baseType.Field(i)
		if baseField.Anonymous {
			if isBaseTypeField(fieldName, baseField.Type) {
				return true
			}
		}
		if fieldName == canonicalFieldName(baseField.Name) {
			return true
		}
	}
	return false
}

func canonicalFieldName(fieldName string) string {
	return strings.ToLower(strings.Replace(fieldName, "_", "", -1))
}

func generateStructField(name string, encrypt bool, schema map[string]any) (*reflect.StructField, errors.Error) {
	goType, err := getGoType(schema)
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("couldn't resolve type for field: \"%s\"", name))
	}
	sf := &reflect.StructField{
		Name: strings.Title(name), //nolint:staticcheck
		Type: goType,
		Tag:  reflect.StructTag(fmt.Sprintf("json:\"%s\"", name)),
	}
	if encrypt {
		sf.Tag = reflect.StructTag(fmt.Sprintf("json:\"%s\" "+
			"gorm:\"serializer:encdec\"", //just encrypt everything for GORM operations - makes things easy
			name))
	}
	return sf, nil
}

func getGoType(schema map[string]any) (reflect.Type, errors.Error) {
	var goType reflect.Type
	jsonType, ok := schema["type"].(string)
	if !ok {
		return nil, errors.BadInput.New("\"type\" property must be a string")
	}
	switch jsonType {
	//TODO: support more types
	case "integer":
		goType = reflect.TypeOf(uint64(0))
	case "boolean":
		goType = reflect.TypeOf(false)
	case "string":
		goType = reflect.TypeOf("")
	case "object":
		goType = reflect.TypeOf(datatypes.JSONMap{})
	default:
		return nil, errors.BadInput.New(fmt.Sprintf("Unsupported type %s", jsonType))
	}
	return goType, nil
}
