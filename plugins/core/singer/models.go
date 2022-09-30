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
	"encoding/json"
	"time"
)

// TapResponse alias for a generic response from a streamed tap
type TapResponse map[string]interface{}

// TapStateValue alias for a generic value for TapState
type TapStateValue map[string]interface{}

// TapSchema the structure of this is determined by the catalog/properties JSON of this tap
type TapSchema map[string]interface{}

// TapMetadata the structure of this is determined by the catalog/properties JSON of this tap
type TapMetadata map[string]interface{}

type Stream struct {
	Stream        string        `json:"stream"`
	TapStreamId   string        `json:"tap_stream_id"`
	Schema        TapSchema     `json:"schema"`
	Metadata      []TapMetadata `json:"metadata"`
	KeyProperties []string      `json:"key_properties"`
}

type Config struct {
	Mappings             interface{}
	Cmd                  string
	TapType              string
	StreamPropertiesFile string
}

type TapRecord[R any] struct {
	Type          string    `json:"type"`
	Stream        string    `json:"stream"`
	TimeExtracted time.Time `json:"time_extracted"`
	Record        *R        `json:"record"`
}

type TapState[V any] struct {
	Type  string `json:"type"`
	Value *V     `json:"value"`
}

func NewTapState[V any](v *V) *TapState[V] {
	return &TapState[V]{
		Type:  "STATE",
		Value: v,
	}
}

func (TapState[V]) TableName() string {
	return "_singer_github_state"
}

func AsTapState[V any](src map[string]interface{}) (*TapState[V], bool) {
	if src["type"] == "STATE" {
		state := TapState[V]{}
		if err := convert(src, &state); err != nil {
			panic(err)
		}
		return &state, true
	}
	return nil, false
}

func AsTapRecord[R any](src TapResponse) (*TapRecord[R], bool) {
	if src["type"] == "RECORD" {
		record := TapRecord[R]{}
		if err := convert(src, &record); err != nil {
			panic(err)
		}
		return &record, true
	}
	return nil, false
}

// Convert inefficient, but there's no other practical way
func convert(src any, dest any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, dest); err != nil {
		return err
	}
	return nil
}
