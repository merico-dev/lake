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
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/mitchellh/mapstructure"
)

type ZentaoApiParams struct {
	ProductId   uint64
	ExecutionId uint64
	ProjectId   uint64
}

type ZentaoOptions struct {
	// TODO add some custom options here if necessary
	// options means some custom params required by plugin running.
	// Such As How many rows do your want
	// You can use it in sub tasks and you need pass it in main.go and pipelines.
	ConnectionId uint64 `json:"connectionId"`
	ProductId    uint64
	ExecutionId  uint64
	ProjectId    uint64
	Tasks        []string `json:"tasks,omitempty"`
	Since        string
}

type ZentaoTaskData struct {
	Options   *ZentaoOptions
	ApiClient *helper.ApiAsyncClient
}

func DecodeAndValidateTaskOptions(options map[string]interface{}) (*ZentaoOptions, error) {
	var op ZentaoOptions
	err := mapstructure.Decode(options, &op)
	if err != nil {
		return nil, err
	}

	if op.ConnectionId == 0 {
		return nil, fmt.Errorf("connectionId is invalid")
	}
	return &op, nil
}
