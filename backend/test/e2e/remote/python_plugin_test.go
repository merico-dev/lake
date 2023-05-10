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
	"testing"

	"github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/test/helper"
	"github.com/stretchr/testify/require"
)

func TestCreateConnection(t *testing.T) {
	client := CreateClient(t)

	CreateTestConnection(client)

	conns := client.ListConnections(PLUGIN_NAME)
	require.Equal(t, 1, len(conns))
	require.Equal(t, TOKEN, conns[0].Token)
}

func TestRemoteScopeGroups(t *testing.T) {
	client := CreateClient(t)
	connection := CreateTestConnection(client)

	output := client.RemoteScopes(helper.RemoteScopesQuery{
		PluginName:   PLUGIN_NAME,
		ConnectionId: connection.ID,
	})

	scopeGroups := output.Children
	require.Equal(t, 1, len(scopeGroups))
	scope := scopeGroups[0]
	require.Equal(t, "Group 1", scope.Name)
	require.Equal(t, "group1", scope.Id)
	require.Equal(t, "group", scope.Type)
	require.Nil(t, scope.ParentId)
	require.Nil(t, scope.Data)
}

func TestRemoteScopes(t *testing.T) {
	client := CreateClient(t)
	connection := CreateTestConnection(client)
	output := client.RemoteScopes(helper.RemoteScopesQuery{
		PluginName:   PLUGIN_NAME,
		ConnectionId: connection.ID,
		GroupId:      "group1",
	})
	scopes := output.Children
	require.Equal(t, 1, len(scopes))
	scope := scopes[0]
	require.Equal(t, "Project 1", scope.Name)
	require.Equal(t, "p1", scope.Id)
	require.Equal(t, "group1", *scope.ParentId)
	require.Equal(t, "scope", scope.Type)
	require.NotNil(t, scope.Data)
	cicdScope := helper.Cast[FakeProject](scope.Data)
	require.Equal(t, connection.ID, cicdScope.ConnectionId)
	require.Equal(t, "p1", cicdScope.Id)
	require.Equal(t, "Project 1", cicdScope.Name)
	require.Equal(t, "http://fake.org/api/project/p1", cicdScope.Url)
}

func TestCreateScope(t *testing.T) {
	client := CreateClient(t)
	conn := CreateTestConnection(client)
	rule := CreateTestTransformationRule(client, conn.ID)
	scope := CreateTestScope(client, rule, conn.ID)

	scopes := client.ListScopes(PLUGIN_NAME, conn.ID, false)
	require.Equal(t, 1, len(scopes))

	cicdScope := helper.Cast[FakeProject](scopes[0].Scope)
	require.Equal(t, conn.ID, cicdScope.ConnectionId)
	require.Equal(t, "p1", cicdScope.Id)
	require.Equal(t, "Project 1", cicdScope.Name)
	require.Equal(t, "http://fake.org/api/project/p1", cicdScope.Url)

	cicdScope.Name = "scope-name-2"
	client.UpdateScope(PLUGIN_NAME, conn.ID, cicdScope.Id, scope)
}

func TestRunPipeline(t *testing.T) {
	client := CreateClient(t)
	conn := CreateTestConnection(client)
	rule := CreateTestTransformationRule(client, conn.ID)
	scope := CreateTestScope(client, rule, conn.ID)
	pipeline := client.RunPipeline(models.NewPipeline{
		Name: "remote_test",
		Plan: []plugin.PipelineStage{
			{
				{
					Plugin:   PLUGIN_NAME,
					Subtasks: nil,
					Options: map[string]interface{}{
						"connectionId": conn.ID,
						"scopeId":      scope.Id,
					},
				},
			},
		},
	})
	require.Equal(t, models.TASK_COMPLETED, pipeline.Status)
	require.Equal(t, 1, pipeline.FinishedTasks)
	require.Equal(t, "", pipeline.ErrorName)
}

func TestBlueprintV200(t *testing.T) {
	client := CreateClient(t)
	connection := CreateTestConnection(client)
	projectName := "Test project"
	client.CreateProject(&helper.ProjectConfig{
		ProjectName: projectName,
	})
	rule := CreateTestTransformationRule(client, connection.ID)
	scope := CreateTestScope(client, rule, connection.ID)
	blueprint := client.CreateBasicBlueprintV2(
		"Test blueprint",
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

	project := client.GetProject(projectName)
	require.Equal(t, blueprint.Name, project.Blueprint.Name)
	client.TriggerBlueprint(blueprint.ID)
}

func TestCreateTxRule(t *testing.T) {
	client := CreateClient(t)
	connection := CreateTestConnection(client)

	res := client.CreateTransformationRule(PLUGIN_NAME, connection.ID, FakeTxRule{Name: "Tx rule", Env: "test env"})
	txRule := helper.Cast[FakeTxRule](res)

	res = client.GetTransformationRule(PLUGIN_NAME, connection.ID, txRule.Id)
	txRule = helper.Cast[FakeTxRule](res)
	require.Equal(t, "Tx rule", txRule.Name)
	require.Equal(t, "test env", txRule.Env)
}

func TestUpdateTxRule(t *testing.T) {
	client := CreateClient(t)
	connection := CreateTestConnection(client)
	res := client.CreateTransformationRule(PLUGIN_NAME, connection.ID, FakeTxRule{Name: "old name", Env: "old env"})
	oldTxRule := helper.Cast[FakeTxRule](res)

	client.PatchTransformationRule(PLUGIN_NAME, connection.ID, oldTxRule.Id, FakeTxRule{Name: "new name", Env: "new env"})

	res = client.GetTransformationRule(PLUGIN_NAME, connection.ID, oldTxRule.Id)
	txRule := helper.Cast[FakeTxRule](res)
	require.Equal(t, "new name", txRule.Name)
	require.Equal(t, "new env", txRule.Env)
}
