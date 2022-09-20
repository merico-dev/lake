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

package singer

import (
	"github.com/apache/incubator-devlake/plugins/helper"
	"time"
)

type TapResponse map[string]interface{}

type Stream[Sch any] struct {
	Stream      string `json:"stream"`
	TapStreamId string `json:"tap_stream_id"`
	Schema      *Sch   `json:"schema"`
	Metadata    []struct {
		Breadcrumb []string `json:"breadcrumb"`
		Metadata   struct {
			TableKeyProperties []string `json:"table-key-properties,omitempty"`
			Inclusion          string   `json:"inclusion,omitempty"`
		} `json:"metadata"`
	} `json:"metadata"`
	KeyProperties []string `json:"key_properties"`
}

type Config struct {
	Mappings interface{}
	Cmd      string
	TapType  string
}

type TapState[V any] struct {
	Type  string `json:"type"`
	Value *V     `json:"value"`
}

func AsTapState[V any](src map[string]interface{}) (*TapState[V], bool) {
	if src["type"] == "STATE" {
		state := TapState[V]{}
		if err := helper.Convert(src, &state); err != nil {
			panic(err)
		}
		return &state, true
	}
	return nil, false
}

type TapRecord[R any] struct {
	Type          string    `json:"type"`
	Stream        string    `json:"stream"`
	TimeExtracted time.Time `json:"time_extracted"`
	Record        *R        `json:"record"`
}

func AsTapRecord[R any](src map[string]interface{}) (*TapRecord[R], bool) {
	if src["type"] == "RECORD" {
		record := TapRecord[R]{}
		if err := helper.Convert(src, &record); err != nil {
			panic(err)
		}
		return &record, true
	}
	return nil, false
}
