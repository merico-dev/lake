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

package remote

import (
	"fmt"
	"github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/apache/incubator-devlake/test/helper"
)

const (
	PLUGIN_NAME     = "fake"
	TOKEN           = "this_is_a_valid_token"
	FAKE_PLUGIN_DIR = "python/test/fakeplugin"
)

type (
	FakePluginConnection struct {
		Id    uint64 `json:"id"`
		Name  string `json:"name"`
		Token string `json:"token"`
	}
	FakeProject struct {
		Id                   string `json:"id"`
		Name                 string `json:"name"`
		ConnectionId         uint64 `json:"connectionId"`
		TransformationRuleId uint64 `json:"transformationRuleId"`
		Url                  string `json:"url"`
	}
	FakeTxRule struct {
		Id   uint64 `json:"id"`
		Name string `json:"name"`
		Env  string `json:"env"`
	}
	BlueprintTestParams struct {
		connection *helper.Connection
		projects   []models.ApiOutputProject
		blueprints []models.Blueprint
		rule       *FakeTxRule
		scope      *FakeProject
	}
)

func ConnectLocalServer(t *testing.T) *helper.DevlakeClient {
	fmt.Println("Connect to server")
	client := helper.StartDevLakeServer(t, nil)
	client.SetTimeout(30 * time.Second)
	return client
}

func CreateClient(t *testing.T) *helper.DevlakeClient {
	path := filepath.Join(helper.ProjectRoot, FAKE_PLUGIN_DIR)
	_ = os.Setenv("REMOTE_PLUGIN_DIR", path)
	client := ConnectLocalServer(t)
	client.AwaitPluginAvailability(PLUGIN_NAME, 60*time.Second)
	return client
}

func CreateTestConnection(client *helper.DevlakeClient) *helper.Connection {
	connection := client.CreateConnection(PLUGIN_NAME,
		FakePluginConnection{
			Name:  "Test connection",
			Token: TOKEN,
		},
	)
	return connection
}

func CreateTestScope(client *helper.DevlakeClient, rule *FakeTxRule, connectionId uint64) *FakeProject {
	scopes := helper.Cast[[]FakeProject](client.CreateScope(PLUGIN_NAME,
		connectionId,
		FakeProject{
			Id:                   "p1",
			Name:                 "Project 1",
			ConnectionId:         connectionId,
			Url:                  "http://fake.org/api/project/p1",
			TransformationRuleId: rule.Id,
		},
	))
	return &scopes[0]
}

func CreateTestTransformationRule(client *helper.DevlakeClient, connectionId uint64) *FakeTxRule {
	rule := helper.Cast[FakeTxRule](client.CreateTransformationRule(PLUGIN_NAME, connectionId, FakeTxRule{Name: "Tx rule", Env: "test env"}))
	return &rule
}

func CreateTestBlueprints(t *testing.T, client *helper.DevlakeClient, count int) *BlueprintTestParams {
	t.Helper()
	connection := CreateTestConnection(client)
	rule := CreateTestTransformationRule(client, connection.ID)
	scope := CreateTestScope(client, rule, connection.ID)
	var bps []models.Blueprint
	var projects []models.ApiOutputProject
	for i := 1; i <= count; i++ {
		projectName := fmt.Sprintf("Test project %d", i)
		client.CreateProject(&helper.ProjectConfig{
			ProjectName: projectName,
		})
		blueprint := client.CreateBasicBlueprintV2(
			fmt.Sprintf("Test blueprint %d", i),
			&helper.BlueprintV2Config{
				Connection: &plugin.BlueprintConnectionV200{
					Plugin:       "fake",
					ConnectionId: connection.ID,
					Scopes: []*plugin.BlueprintScopeV200{
						{
							Id:   scope.Id,
							Name: "Test scope",
							Entities: []string{
								plugin.DOMAIN_TYPE_CICD,
							},
						},
					},
				},
				SkipOnFail:  true,
				ProjectName: projectName,
			},
		)
		plan, err := blueprint.UnmarshalPlan()
		require.NoError(t, err)
		_ = plan
		bps = append(bps, blueprint)
		project := client.GetProject(projectName)
		require.Equal(t, blueprint.Name, project.Blueprint.Name)
		projects = append(projects, project)
	}
	return &BlueprintTestParams{
		connection: connection,
		projects:   projects,
		blueprints: bps,
		rule:       rule,
		scope:      scope,
	}
}
