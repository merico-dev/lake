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
	"github.com/apache/incubator-devlake/plugins/github_singer/api"
	"github.com/apache/incubator-devlake/plugins/github_singer/models"
	"github.com/apache/incubator-devlake/plugins/github_singer/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/github_singer/tasks"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"strings"
	"time"
)

// make sure interface is implemented
var _ core.PluginMeta = (*GithubSinger)(nil)
var _ core.PluginInit = (*GithubSinger)(nil)
var _ core.PluginTask = (*GithubSinger)(nil)
var _ core.PluginApi = (*GithubSinger)(nil)
var _ core.PluginBlueprintV100 = (*GithubSinger)(nil)
var _ core.CloseablePluginTask = (*GithubSinger)(nil)

type GithubSinger struct{}

func (plugin GithubSinger) Description() string {
	return "collect some GithubSinger data"
}

func (plugin GithubSinger) Init(config *viper.Viper, logger core.Logger, db *gorm.DB) errors.Error {
	api.Init(config, logger, db)
	return nil
}

func (plugin GithubSinger) SubTaskMetas() []core.SubTaskMeta {
	// TODO add your sub task here
	return []core.SubTaskMeta{
		tasks.ExtractPrMeta,
		tasks.ExtractIssuesMeta,
		tasks.ExtractAssigneeMeta,
	}
}

func (plugin GithubSinger) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	op, err := tasks.DecodeAndValidateTaskOptions(options)
	if err != nil {
		return nil, err
	}
	connectionHelper := helper.NewConnectionHelper(
		taskCtx,
		nil,
	)
	connection := &models.GithubSingerConnection{}
	err = connectionHelper.FirstById(connection, op.ConnectionId)
	if err != nil {
		return nil, errors.Default.Wrap(err, "unable to get GithubSinger connection by the given connection ID")
	}
	//var since time.Time
	//if op.Since != "" {
	//	since, err = errors.Convert01(time.Parse("2006-01-02T15:04:05Z", op.Since))
	//	if err != nil {
	//		return nil, errors.BadInput.Wrap(err, "invalid value for `since`")
	//	}
	//}
	endpoint := strings.TrimSuffix(connection.Endpoint, "/")
	config := &models.GithubConfig{
		AccessToken:    connection.Token,
		Repository:     options["repo"].(string),
		StartDate:      options["start_date"].(time.Time),
		RequestTimeout: 300,
		BaseUrl:        endpoint,
	}
	return &tasks.GithubSingerTaskData{
		Options: op,
		Config:  config,
	}, nil
}

// PkgPath information lost when compiled as plugin(.so)
func (plugin GithubSinger) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/github_singer"
}

func (plugin GithubSinger) MigrationScripts() []migration.Script {
	return migrationscripts.All()
}

func (plugin GithubSinger) ApiResources() map[string]map[string]core.ApiResourceHandler {
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

func (plugin GithubSinger) MakePipelinePlan(connectionId uint64, scope []*core.BlueprintScopeV100) (core.PipelinePlan, errors.Error) {
	return api.MakePipelinePlan(plugin.SubTaskMetas(), connectionId, scope)
}

func (plugin GithubSinger) Close(taskCtx core.TaskContext) errors.Error {
	_, ok := taskCtx.GetData().(*tasks.GithubSingerTaskData)
	if !ok {
		return errors.Default.New(fmt.Sprintf("GetData failed when try to close %+v", taskCtx))
	}
	return nil
}
