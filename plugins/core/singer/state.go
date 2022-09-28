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
	"gorm.io/datatypes"
)

type RawTapState struct {
	Id           string `gorm:"primaryKey;type:varchar(255)"`
	ConnectionId uint64
	Type         string `gorm:"type:varchar(255)"`
	Value        datatypes.JSON
}

func FromState[State any](connectionId uint64, t *TapState[State]) *RawTapState {
	b, err := json.Marshal(t.Value)
	if err != nil {
		panic(err)
	}
	return &RawTapState{
		ConnectionId: connectionId,
		Type:         t.Type,
		Value:        b,
	}
}

func ToState[State any](raw *RawTapState) *TapState[State] {
	val := new(State)
	err := json.Unmarshal(raw.Value, val)
	if err != nil {
		panic(err)
	}
	return &TapState[State]{
		Type:  raw.Type,
		Value: val,
	}
}

func (*RawTapState) TableName() string {
	return "singer_tap_state"
}
