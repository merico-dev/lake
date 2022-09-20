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
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/logger"
	"github.com/apache/incubator-devlake/mocks"
	"github.com/apache/incubator-devlake/models/common"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/github/models"
	"github.com/apache/incubator-devlake/plugins/github/tasks"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/runner"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessScope(t *testing.T) {
	cfg := config.GetConfig()
	log := logger.Global.Nested("github")
	db, _ := runner.NewGormDb(cfg, log)
	Init(cfg, log, db)
	connection := &models.GithubConnection{
		RestConnection: helper.RestConnection{
			BaseConnection: helper.BaseConnection{
				Name: "github-test",
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
	mockMeta := mocks.NewPluginMeta(t)
	mockMeta.On("RootPkgPath").Return("github.com/apache/incubator-devlake/plugins/github")
	err := core.RegisterPlugin("github", mockMeta)
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
              "dora": {
                "environment": "pattern",
                "environmentRegex": "xxxx"
              }
            }`),
	}
	apiRepo := &tasks.GithubApiRepo{
		GithubId: 123,
		CloneUrl: "HttpUrlToRepo",
	}
	scopes := make([]*core.BlueprintScopeV100, 0)
	scopes = append(scopes, bs)
	plan := make(core.PipelinePlan, len(scopes))
	for i, scopeElem := range scopes {
		plan, err = processScope(nil, 1, scopeElem, i, plan, apiRepo, connection)
		assert.Nil(t, err)
	}
	planJson, err1 := json.Marshal(plan)
	assert.Nil(t, err1)
	expectPlan := `[[{"plugin":"github","subtasks":[],"options":{"connectionId":1,"owner":"test","repo":"testRepo","transformationRules":{"prType":"hey,man,wasup"}}},{"plugin":"gitextractor","subtasks":null,"options":{"proxy":"","repoId":"github:GithubRepo:1:123","url":"//git:123@HttpUrlToRepo"}}],[{"plugin":"refdiff","subtasks":null,"options":{"tagsLimit":10,"tagsOrder":"reverse semver","tagsPattern":"pattern"}}],[{"plugin":"dora","subtasks":null,"options":{"repoId":"github:GithubRepo:1:123","tasks":["EnrichTaskEnv"],"transformation":{"environment":"pattern","environmentRegex":"xxxx"}}}]]`
	assert.Equal(t, expectPlan, string(planJson))
}
