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
	"encoding/json"
	"github.com/apache/incubator-devlake/errors"
	"github.com/go-playground/validator/v10"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

// DecodeMapStruct with time.Time and Iso8601Time support
func DecodeMapStruct(input map[string]interface{}, result interface{}) errors.Error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ZeroFields: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
				if data == nil {
					return nil, nil
				}
				if t == reflect.TypeOf(json.RawMessage{}) {
					return json.Marshal(data)
				}

				if t != reflect.TypeOf(Iso8601Time{}) && t != reflect.TypeOf(time.Time{}) {
					return data, nil
				}

				var tt time.Time
				var err error

				switch f.Kind() {
				case reflect.String:
					tt, err = ConvertStringToTime(data.(string))
				case reflect.Float64:
					tt = time.Unix(0, int64(data.(float64))*int64(time.Millisecond))
				case reflect.Int64:
					tt = time.Unix(0, data.(int64)*int64(time.Millisecond))
				}
				if err != nil {
					return data, nil
				}

				if t == reflect.TypeOf(Iso8601Time{}) {
					return Iso8601Time{time: tt}, nil
				}
				return tt, nil
			},
		),
		Result: result,
	})
	if err != nil {
		return errors.Convert(err)
	}

	if err := decoder.Decode(input); err != nil {
		return errors.Convert(err)
	}
	return errors.Convert(err)
}

// Decode decodes `source` into `target`. Pass an optional validator to validate the target.
func Decode(source interface{}, target interface{}, vld *validator.Validate) errors.Error {
	if err := mapstructure.Decode(source, target); err != nil {
		return errors.Default.Wrap(err, "error decoding map into target type")
	}
	if vld != nil {
		if err := vld.Struct(target); err != nil {
			return errors.Default.Wrap(err, "error validating target")
		}
	}
	return nil
}
