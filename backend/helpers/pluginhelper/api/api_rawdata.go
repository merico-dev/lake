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
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-devlake/core/errors"
	plugin "github.com/apache/incubator-devlake/core/plugin"
	"reflect"
	"time"

	"gorm.io/datatypes"
)

// RawData is raw data structure in DB storage
type RawData struct {
	ID        uint64 `gorm:"primaryKey"`
	Params    string `gorm:"type:varchar(255);index"`
	Data      []byte
	Url       string
	Input     datatypes.JSON
	CreatedAt time.Time
}

type TaskOptions interface {
	GetParams() any
}

// RawDataSubTaskArgs FIXME ...
type RawDataSubTaskArgs struct {
	Ctx plugin.SubTaskContext

	//	Table store raw data
	Table string `comment:"Raw data table name"`

	// Deprecated: Use Options instead
	// This struct will be JSONEncoded and stored into database along with raw data itself, to identity minimal set of
	// data to be processed, for example, we process JiraIssues by Board
	Params any `comment:"To identify a set of records with same UrlTemplate, i.e. {ConnectionId, BoardId} for jira entities"`

	Options TaskOptions `comment:"To identify a set of records with same UrlTemplate, i.e. {ConnectionId, BoardId} for jira entities"`
}

// RawDataSubTask is Common features for raw data sub-tasks
type RawDataSubTask struct {
	args   *RawDataSubTaskArgs
	table  string
	params string
}

// NewRawDataSubTask constructor for RawDataSubTask
func NewRawDataSubTask(args RawDataSubTaskArgs) (*RawDataSubTask, errors.Error) {
	if args.Ctx == nil {
		return nil, errors.Default.New("Ctx is required for RawDataSubTask")
	}
	if args.Table == "" {
		return nil, errors.Default.New("Table is required for RawDataSubTask")
	}
	var params any
	if args.Options != nil {
		params = args.Options.GetParams()
	} else { // fallback to old way
		params = args.Params
	}
	paramsString := ""
	if params == nil || reflect.ValueOf(params).IsZero() {
		args.Ctx.GetLogger().Warn(nil, fmt.Sprintf("Missing `Params` for raw data subtask %s", args.Ctx.GetName()))
	} else {
		// TODO: maybe sort it to make it consistent
		paramsBytes, err := json.Marshal(params)
		if err != nil {
			return nil, errors.Default.Wrap(err, "unable to serialize subtask parameters")
		}
		paramsString = string(paramsBytes)
	}
	return &RawDataSubTask{
		args:   &args,
		table:  fmt.Sprintf("_raw_%s", args.Table),
		params: paramsString,
	}, nil
}

// GetTable returns the raw table name
func (r *RawDataSubTask) GetTable() string {
	return r.table
}

// GetParams returns the raw params
func (r *RawDataSubTask) GetParams() string {
	return r.params
}
