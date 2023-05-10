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
