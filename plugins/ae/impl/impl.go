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

	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/ae/api"
	"github.com/apache/incubator-devlake/plugins/ae/models"
	"github.com/apache/incubator-devlake/plugins/ae/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/ae/tasks"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var _ core.PluginMeta = (*AE)(nil)
var _ core.PluginInit = (*AE)(nil)
var _ core.PluginTask = (*AE)(nil)
var _ core.PluginApi = (*AE)(nil)
var _ core.PluginMigration = (*AE)(nil)
var _ core.CloseablePluginTask = (*AE)(nil)

type AE struct{}

func (plugin AE) Init(config *viper.Viper, logger core.Logger, db *gorm.DB) errors.Error {
	api.Init(config, logger, db)
	return nil
}
func (plugin AE) GetTablesInfo() []core.Tabler {
	return []core.Tabler{
		&models.AECommit{},
		&models.AEProject{},
		&models.AeConnection{},
		&models.AeResponse{},
	}
}

func (plugin AE) Description() string {
	return "To collect and enrich data from AE"
}

func (plugin AE) SubTaskMetas() []core.SubTaskMeta {
	return []core.SubTaskMeta{
		tasks.CollectProjectMeta,
		tasks.CollectCommitsMeta,
		tasks.ExtractProjectMeta,
		tasks.ExtractCommitsMeta,
		tasks.ConvertCommitsMeta,
	}
}

func (plugin AE) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	var op tasks.AeOptions
	err := helper.Decode(options, &op, nil)
	if err != nil {
		return nil, errors.Default.Wrap(err, "AE plugin could not decode options")
	}
	if op.ProjectId <= 0 {
		return nil, errors.Default.New("projectId is required")
	}
	connection := &models.AeConnection{}
	connectionHelper := helper.NewConnectionHelper(
		taskCtx,
		nil,
	)
	err = connectionHelper.FirstById(connection, op.ConnectionId)
	if err != nil {
		return nil, errors.Default.Wrap(err, "error getting connection for AE plugin")
	}
	apiClient, err := tasks.CreateApiClient(taskCtx, connection)
	if err != nil {
		return nil, err
	}
	return &tasks.AeTaskData{
		Options:   &op,
		ApiClient: apiClient,
	}, nil
}

func (plugin AE) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/ae"
}

func (plugin AE) MigrationScripts() []core.MigrationScript {
	return migrationscripts.All()
}

func (plugin AE) ApiResources() map[string]map[string]core.ApiResourceHandler {
	return map[string]map[string]core.ApiResourceHandler{
		"test": {
			"POST": api.TestConnection,
		},
		"connections": {
			"GET":  api.ListConnections,
			"POST": api.PostConnections,
		},
		"connections/:connectionId": {
			"GET":    api.GetConnection,
			"PATCH":  api.PatchConnection,
			"DELETE": api.DeleteConnection,
		},
	}
}

func (plugin AE) Close(taskCtx core.TaskContext) errors.Error {
	data, ok := taskCtx.GetData().(*tasks.AeTaskData)
	if !ok {
		return errors.Default.New(fmt.Sprintf("GetData failed when try to close %+v", taskCtx))
	}
	data.ApiClient.Release()
	return nil
}
