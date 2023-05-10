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

package helper

import (
	"github.com/apache/incubator-devlake/core/models"
	"time"

	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
)

type (
	ProjectPlugin struct {
		Name    string
		Options any
	}
	ProjectConfig struct {
		ProjectName        string
		ProjectDescription string
		EnableDora         bool
		MetricPlugins      []ProjectPlugin
	}

	ScopeResponse struct {
		Scope                  any
		TransformationRuleName string
		Blueprints             []*models.Blueprint
	}
)

type Connection struct {
	api.BaseConnection `mapstructure:",squash"`
	api.RestConnection `mapstructure:",squash"`
	api.AccessToken    `mapstructure:",squash"`
}

type BlueprintV2Config struct {
	Connection  *plugin.BlueprintConnectionV200
	TimeAfter   *time.Time
	SkipOnFail  bool
	ProjectName string
}
type RemoteScopesChild struct {
	Type     string      `json:"type"`
	ParentId *string     `json:"parentId"`
	Id       string      `json:"id"`
	Name     string      `json:"name"`
	Data     interface{} `json:"data"`
}

type RemoteScopesOutput struct {
	Children      []RemoteScopesChild `json:"children"`
	NextPageToken string              `json:"nextPageToken"`
}

type RemoteScopesQuery struct {
	PluginName   string
	ConnectionId uint64
	GroupId      string
	PageToken    string
	Params       map[string]string
}

type SearchRemoteScopesOutput struct {
	Children []RemoteScopesChild `json:"children"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
}

type SearchRemoteScopesQuery struct {
	PluginName   string
	ConnectionId uint64
	Search       string
	Page         int
	PageSize     int
	Params       map[string]string
}
