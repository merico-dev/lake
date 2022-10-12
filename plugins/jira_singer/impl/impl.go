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
	"github.com/apache/incubator-devlake/migration"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/tap"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/plugins/jira_singer/api"
	"github.com/apache/incubator-devlake/plugins/jira_singer/models"
	"github.com/apache/incubator-devlake/plugins/jira_singer/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/jira_singer/tasks"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"strings"
	"time"
)

// make sure interface is implemented
var _ core.PluginMeta = (*JiraSinger)(nil)
var _ core.PluginInit = (*JiraSinger)(nil)
var _ core.PluginTask = (*JiraSinger)(nil)
var _ core.PluginApi = (*JiraSinger)(nil)
var _ core.PluginBlueprintV100 = (*JiraSinger)(nil)
var _ core.CloseablePluginTask = (*JiraSinger)(nil)

type JiraSinger struct{}

func (plugin JiraSinger) Description() string {
	return "collect some JiraSinger data"
}

func (plugin JiraSinger) Init(config *viper.Viper, logger core.Logger, db *gorm.DB) errors.Error {
	api.Init(config, logger, db)
	return nil
}

func (plugin JiraSinger) SubTaskMetas() []core.SubTaskMeta {
	// TODO add your sub task here
	return []core.SubTaskMeta{
		tasks.ExtractProjectsMeta,
	}
}

func (plugin JiraSinger) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	op, err := tasks.DecodeAndValidateTaskOptions(options)
	if err != nil {
		return nil, err
	}
	connectionHelper := helper.NewConnectionHelper(
		taskCtx,
		nil,
	)
	connection := &models.JiraSingerConnection{}
	err = connectionHelper.FirstById(connection, op.ConnectionId)
	if err != nil {
		return nil, errors.Default.Wrap(err, "unable to get JiraSinger connection by the given connection ID")
	}
	endpoint := strings.Split(connection.Endpoint, "/rest")[0]
	config := &models.JiraConfig{
		StartDate:      options["start_date"].(time.Time),
		Username:       connection.Username,
		Password:       connection.Password,
		BaseUrl:        endpoint,
		RequestTimeout: 300,
		Groups:         "jira-administrators, site-admins, jira-software-users",
	}
	op.TapProvider = func() (tap.Tap, errors.Error) {
		return helper.NewSingerTapClient(&helper.SingerTapArgs{
			Mappings:             config,
			TapClass:             "TAP_JIRA",
			StreamPropertiesFile: "jira.json",
		})
	}
	if err != nil {
		return nil, err
	}
	return &tasks.JiraSingerTaskData{
		Options: op,
		Config:  config,
	}, nil
}

// PkgPath information lost when compiled as plugin(.so)
func (plugin JiraSinger) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/jira_singer"
}

func (plugin JiraSinger) MigrationScripts() []migration.Script {
	return migrationscripts.All()
}

func (plugin JiraSinger) ApiResources() map[string]map[string]core.ApiResourceHandler {
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

func (plugin JiraSinger) MakePipelinePlan(connectionId uint64, scope []*core.BlueprintScopeV100) (core.PipelinePlan, errors.Error) {
	return api.MakePipelinePlan(plugin.SubTaskMetas(), connectionId, scope)
}

func (plugin JiraSinger) Close(taskCtx core.TaskContext) errors.Error {
	_, ok := taskCtx.GetData().(*tasks.JiraSingerTaskData)
	if !ok {
		return errors.Default.New(fmt.Sprintf("GetData failed when try to close %+v", taskCtx))
	}
	return nil
}
