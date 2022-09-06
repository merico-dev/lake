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
	"fmt"
	"github.com/apache/incubator-devlake/migration"
	"github.com/apache/incubator-devlake/plugins/core"
    "github.com/apache/incubator-devlake/plugins/zentao/api"
    "github.com/apache/incubator-devlake/plugins/zentao/models"
    "github.com/apache/incubator-devlake/plugins/zentao/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/zentao/tasks"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// make sure interface is implemented
var _ core.PluginMeta = (*Zentao)(nil)
var _ core.PluginInit = (*Zentao)(nil)
var _ core.PluginTask = (*Zentao)(nil)
var _ core.PluginApi = (*Zentao)(nil)
var _ core.PluginBlueprintV100 = (*Zentao)(nil)
var _ core.CloseablePluginTask = (*Zentao)(nil)



type Zentao struct{}

func (plugin Zentao) Description() string {
	return "collect some Zentao data"
}

func (plugin Zentao) Init(config *viper.Viper, logger core.Logger, db *gorm.DB) error {
	api.Init(config, logger, db)
	return nil
}

func (plugin Zentao) SubTaskMetas() []core.SubTaskMeta {
	// TODO add your sub task here
	return []core.SubTaskMeta{
		tasks.CollectProjectMeta,
	}
}

func (plugin Zentao) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, error) {
	op, err := tasks.DecodeAndValidateTaskOptions(options)
    if err != nil {
        return nil, err
    }
    connectionHelper := helper.NewConnectionHelper(
        taskCtx,
        nil,
    )
    connection := &models.ZentaoConnection{}
    err = connectionHelper.FirstById(connection, op.ConnectionId)
    if err != nil {
        return nil, fmt.Errorf("unable to get Zentao connection by the given connection ID: %v", err)
    }

    apiClient, err := tasks.NewZentaoApiClient(taskCtx, connection)
    if err != nil {
        return nil, fmt.Errorf("unable to get Zentao API client instance: %v", err)
    }

    return &tasks.ZentaoTaskData{
        Options:   op,
        ApiClient: apiClient,
    }, nil
}

// PkgPath information lost when compiled as plugin(.so)
func (plugin Zentao) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/zentao"
}

func (plugin Zentao) MigrationScripts() []migration.Script {
	return migrationscripts.All()
}

func (plugin Zentao) ApiResources() map[string]map[string]core.ApiResourceHandler {
    return map[string]map[string]core.ApiResourceHandler{
        "test": {
            "POST": api.TestConnection,
        },
        "connections": {
            "POST": api.PostConnections,
            "GET":  api.ListConnections,
        },
        "connections/:connectionId": {
            "GET":    api.GetConnection,
            "PATCH":  api.PatchConnection,
            "DELETE": api.DeleteConnection,
        },
    }
}

func (plugin Zentao) MakePipelinePlan(connectionId uint64, scope []*core.BlueprintScopeV100) (core.PipelinePlan, error) {
	return api.MakePipelinePlan(plugin.SubTaskMetas(), connectionId, scope)
}

func (plugin Zentao) Close(taskCtx core.TaskContext) error {
	data, ok := taskCtx.GetData().(*tasks.ZentaoTaskData)
	if !ok {
		return fmt.Errorf("GetData failed when try to close %+v", taskCtx)
	}
	data.ApiClient.Release()
	return nil
}
