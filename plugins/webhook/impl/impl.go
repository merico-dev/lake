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

package impl

import (
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/webhook/api"
	"github.com/apache/incubator-devlake/plugins/webhook/models/migrationscripts"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// make sure interface is implemented
var _ core.PluginMeta = (*Webhook)(nil)
var _ core.PluginInit = (*Webhook)(nil)
var _ core.PluginApi = (*Webhook)(nil)
var _ core.PluginMigration = (*Webhook)(nil)

type Webhook struct{}

func (plugin Webhook) Description() string {
	return "collect some Webhook data"
}

func (plugin Webhook) Init(config *viper.Viper, logger core.Logger, db *gorm.DB) errors.Error {
	api.Init(config, logger, db)
	return nil
}

// PkgPath information lost when compiled as plugin(.so)
func (plugin Webhook) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/webhook"
}

func (plugin Webhook) MigrationScripts() []core.MigrationScript {
	return migrationscripts.All()
}

func (plugin Webhook) ApiResources() map[string]map[string]core.ApiResourceHandler {
	return map[string]map[string]core.ApiResourceHandler{
		"connections": {
			"POST": api.PostConnections,
			"GET":  api.ListConnections,
		},
		"connections/:connectionId": {
			"GET":    api.GetConnection,
			"PATCH":  api.PatchConnection,
			"DELETE": api.DeleteConnection,
		},
		":connectionId/cicd_tasks": {
			"POST": api.PostCicdTask,
		},
		":connectionId/cicd_pipeline/:pipelineName/finish": {
			"POST": api.PostPipelineFinish,
		},
		":connectionId/deployments": {
			"POST": api.PostDeploymentCicdTask,
		},
		":connectionId/issues": {
			"POST": api.PostIssue,
		},
		":connectionId/issue/:boardKey/:issueKey/close": {
			"POST": api.CloseIssue,
		},
	}
}
