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

package tasks

import (
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/helper"
)

type SonarqubeApiParams struct {
}

type SonarqubeOptions struct {
	// TODO add some custom options here if necessary
	// options means some custom params required by plugin running.
	// Such As How many rows do your want
	// You can use it in sub tasks and you need pass it in main.go and pipelines.
    ConnectionId               uint64   `json:"connectionId"`
    Tasks                      []string `json:"tasks,omitempty"`
    Since                      string
}

type SonarqubeTaskData struct {
	Options   *SonarqubeOptions
	ApiClient *helper.ApiAsyncClient
}

func DecodeAndValidateTaskOptions(options map[string]interface{}) (*SonarqubeOptions, errors.Error) {
	var op SonarqubeOptions
	if err := helper.Decode(options, &op, nil); err != nil {
		return nil, err
	}
	if op.ConnectionId == 0 {
		return nil, errors.Default.New("connectionId is invalid")
	}
	return &op, nil
}