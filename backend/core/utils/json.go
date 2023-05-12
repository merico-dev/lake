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

package utils

import (
	"fmt"
	"reflect"

	"github.com/apache/incubator-devlake/core/errors"
)

type JsonObject = map[string]any
type JsonArray = []any

func GetProperty[T any](object JsonObject, key string) (T, errors.Error) {
	property, ok := object[key]
	if !ok {
		return *new(T), errors.Default.New(fmt.Sprintf("Missing property %s", key))
	}
	return convert[T](property)
}

func GetItem[T any](array JsonArray, index int) (T, errors.Error) {
	if index < 0 || index >= len(array) {
		return *new(T), errors.Default.New(fmt.Sprintf("Index %d out of range", index))
	}
	return convert[T](array[index])
}

func convert[T any](value any) (T, errors.Error) {
	var t T
	tType := reflect.TypeOf(t)
	if tType.Kind() == reflect.Slice {
		valueSlice, ok := value.([]any)
		if !ok {
			return t, errors.Default.New("Value is not a slice")
		}
		elemType := tType.Elem()
		result := reflect.MakeSlice(tType, 0, len(valueSlice))
		for _, v := range valueSlice {
			elem := reflect.ValueOf(v).Convert(elemType)
			result = reflect.Append(result, elem)
		}
		return result.Interface().(T), nil
	} else {
		result, ok := value.(T)
		if !ok {
			return t, errors.Default.New(fmt.Sprintf("Value is not of type %T", t))
		}
		return result, nil
	}
}
