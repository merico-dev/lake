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
	"bytes"
	"encoding/json"
	"github.com/apache/incubator-devlake/mocks"
	"github.com/apache/incubator-devlake/models/common"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/gitee/tasks"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"testing"
)

func TestMakePipelinePlan(t *testing.T) {
	connection := &models.GiteeConnection{
		RestConnection: helper.RestConnection{
			BaseConnection: helper.BaseConnection{
				Name: "gitee-test",
				Model: common.Model{
					ID: 1,
				},
			},
			Endpoint:         "https://api.github.com/",
			Proxy:            "",
			RateLimitPerHour: 0,
		},
		AccessToken: helper.AccessToken{
			Token: "123",
		},
	}
	mockApiCLient := mocks.NewApiClientGetter(t)
	repo := &tasks.GiteeApiRepoResponse{
		GiteeId: 12345,
		HTMLUrl: "https://this_is_cloneUrl",
	}
	js, err := json.Marshal(repo)
	assert.Nil(t, err)
	res := &http.Response{}
	res.Body = io.NopCloser(bytes.NewBuffer(js))
	res.StatusCode = http.StatusOK
	mockApiCLient.On("Get", "repos/test/testRepo", mock.Anything, mock.Anything).Return(res, nil)
	mockMeta := mocks.NewPluginMeta(t)
	mockMeta.On("RootPkgPath").Return("github.com/apache/incubator-devlake/plugins/gitee")
	err = core.RegisterPlugin("gitee", mockMeta)
	assert.Nil(t, err)
	bs := &core.BlueprintScopeV100{
		Entities: []string{"CODE"},
		Options: json.RawMessage(`{
              "owner": "test",
              "repo": "testRepo"
            }`),
		Transformation: json.RawMessage(`{
              "prType": "hey,man,wasup",
              "refdiff": {
                "tagsPattern": "pattern",
                "tagsLimit": 10,
                "tagsOrder": "reverse semver"
              },
              "productionPattern": "xxxx"
            }`),
	}
	scopes := make([]*core.BlueprintScopeV100, 0)
	scopes = append(scopes, bs)
	plan, err := makePipelinePlan(nil, scopes, mockApiCLient, connection)
	assert.Nil(t, err)

	expectPlan := core.PipelinePlan{
		core.PipelineStage{
			{
				Plugin:   "gitee",
				Subtasks: []string{},
				Options: map[string]interface{}{
					"connectionId": uint64(1),
					"owner":        "test",
					"repo":         "testRepo",
					"transformationRules": map[string]interface{}{
						"prType": "hey,man,wasup",
					},
				},
			},
			{
				Plugin: "gitextractor",
				Options: map[string]interface{}{
					"proxy":  "",
					"repoId": "gitee:GiteeRepo:1:12345",
					"url":    "https://git:123@this_is_cloneUrl",
				},
			},
		},
		core.PipelineStage{
			{
				Plugin: "refdiff",
				Options: map[string]interface{}{
					"tagsLimit":   float64(10),
					"tagsOrder":   "reverse semver",
					"tagsPattern": "pattern",
				},
			},
		},
		core.PipelineStage{
			{
				Plugin:   "dora",
				Subtasks: []string{"EnrichTaskEnv"},
				Options: map[string]interface{}{
					"repoId": "gitee:GiteeRepo:1:12345",
					"transformationRules": map[string]interface{}{
						"productionPattern": "xxxx",
					},
				},
			},
		},
	}
	assert.Equal(t, expectPlan, plan)
}
