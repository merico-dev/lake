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

package tap

import (
	"encoding/json"
	"github.com/apache/incubator-devlake/plugins/core"
	"gorm.io/datatypes"
	"time"
)

// abstract tap types
type (
	// Result alias for a generic response from a streamed tap
	Result map[string]interface{}
	// Response wraps a unit of data returned by a tap

	// Record the fields embedded in a singer-tap record. The specifics of the record are tap-implementation specific.
	Record[R any] struct {
		Type          string    `json:"type"`
		Stream        string    `json:"stream"`
		TimeExtracted time.Time `json:"time_extracted"`
		Record        *R        `json:"record"`
	}
	// State the fields embedded in a singer-tap state. The specifics of the value are tap-implementation specific.
	State struct {
		Type  string                 `json:"type"`
		Value map[string]interface{} `json:"value"`
	}

	// RawState The raw-database version of State
	RawState struct {
		Id           string `gorm:"primaryKey;type:varchar(255)"`
		ConnectionId uint64
		Type         string `gorm:"type:varchar(255)"`
		Value        datatypes.JSON
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}
)

// TableName the table name
func (*RawState) TableName() string {
	return "tap_state"
}

// NewTapState creates a new tap State variable
func NewTapState(v map[string]any) *State {
	return &State{
		Type:  "STATE",
		Value: v,
	}
}

// FromState converts State to RawState
func FromState(connectionId uint64, t *State) *RawState {
	b, err := json.Marshal(t.Value)
	if err != nil {
		panic(err)
	}
	return &RawState{
		ConnectionId: connectionId,
		Type:         t.Type,
		Value:        b,
	}
}

// ToState converts RawState to State
func ToState(raw *RawState) *State {
	val := new(map[string]any)
	err := json.Unmarshal(raw.Value, val)
	if err != nil {
		panic(err)
	}
	return &State{
		Type:  raw.Type,
		Value: *val,
	}
}

// AsTapState tries to convert the map object to a State. Returns false if it can't be done.
func AsTapState(src Result) (*State, bool) {
	if src["type"] == "STATE" {
		state := State{}
		if err := convert(src, &state); err != nil {
			panic(err)
		}
		return &state, true
	}
	return nil, false
}

// AsTapRecord tries to convert the map object to a Record. Returns false if it can't be done.
func AsTapRecord[R any](src Result) (*Record[R], bool) {
	if src["type"] == "RECORD" {
		record := Record[R]{}
		if err := convert(src, &record); err != nil {
			panic(err)
		}
		return &record, true
	}
	return nil, false
}

// Convert a generic converter between two types (not fast, but practical here)
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

var _ core.Tabler = (*RawState)(nil)
