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

	"github.com/spf13/viper"
	"gorm.io/gorm"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/feishu/api"
	"github.com/apache/incubator-devlake/plugins/feishu/models"
	"github.com/apache/incubator-devlake/plugins/feishu/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/feishu/tasks"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var _ core.PluginMeta = (*Feishu)(nil)
var _ core.PluginInit = (*Feishu)(nil)
var _ core.PluginTask = (*Feishu)(nil)
var _ core.PluginApi = (*Feishu)(nil)
var _ core.PluginMigration = (*Feishu)(nil)
var _ core.CloseablePluginTask = (*Feishu)(nil)

type Feishu struct{}

func (plugin Feishu) Init(config *viper.Viper, logger core.Logger, db *gorm.DB) errors.Error {
	api.Init(config, logger, db)

	// FIXME after config-ui support feishu plugin
	// save env to db where name=feishu
	connection := &models.FeishuConnection{}
	if db.Migrator().HasTable(connection) {
		if err := db.Find(connection, map[string]string{"name": "Feishu"}).Error; err != nil {
			return errors.Convert(err)
		}
		if connection.ID != 0 {
			encodeKey := config.GetString(core.EncodeKeyEnvStr)
			connection.Endpoint = config.GetString(`FEISHU_ENDPOINT`)
			connection.AppId = config.GetString(`FEISHU_APPID`)
			connection.SecretKey = config.GetString(`FEISHU_APPSCRECT`)
			if connection.Endpoint != `` && connection.AppId != `` && connection.SecretKey != `` && encodeKey != `` {
				err := helper.UpdateEncryptFields(connection, func(plaintext string) (string, errors.Error) {
					return core.Encrypt(encodeKey, plaintext)
				})
				if err != nil {
					return err
				}
				// update from .env and save to db
				err = errors.Convert(db.Updates(connection).Error)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (plugin Feishu) GetTablesInfo() []core.Tabler {
	return []core.Tabler{
		&models.FeishuConnection{},
		&models.FeishuMeetingTopUserItem{},
	}
}

func (plugin Feishu) Description() string {
	return "To collect and enrich data from Feishu"
}

func (plugin Feishu) SubTaskMetas() []core.SubTaskMeta {
	return []core.SubTaskMeta{
		tasks.CollectMeetingTopUserItemMeta,
		tasks.ExtractMeetingTopUserItemMeta,
	}
}

func (plugin Feishu) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	var op tasks.FeishuOptions
	if err := helper.Decode(options, &op, nil); err != nil {
		return nil, err
	}

	connectionHelper := helper.NewConnectionHelper(
		taskCtx,
		nil,
	)
	connection := &models.FeishuConnection{}
	err := connectionHelper.FirstById(connection, op.ConnectionId)
	if err != nil {
		return nil, err
	}

	apiClient, err := tasks.NewFeishuApiClient(taskCtx, connection)
	if err != nil {
		return nil, err
	}
	return &tasks.FeishuTaskData{
		Options:   &op,
		ApiClient: apiClient,
	}, nil
}

func (plugin Feishu) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/feishu"
}

func (plugin Feishu) MigrationScripts() []core.MigrationScript {
	return migrationscripts.All()
}

func (plugin Feishu) ApiResources() map[string]map[string]core.ApiResourceHandler {
	return map[string]map[string]core.ApiResourceHandler{
		"test": {
			"POST": api.TestConnection,
		},
		"connections": {
			"POST": api.PostConnections,
			"GET":  api.ListConnections,
		},
		"connections/:connectionId": {
			"PATCH":  api.PatchConnection,
			"DELETE": api.DeleteConnection,
			"GET":    api.GetConnection,
		},
	}
}

func (plugin Feishu) Close(taskCtx core.TaskContext) errors.Error {
	data, ok := taskCtx.GetData().(*tasks.FeishuTaskData)
	if !ok {
		return errors.Default.New(fmt.Sprintf("GetData failed when try to close %+v", taskCtx))
	}
	data.ApiClient.Release()
	return nil
}
