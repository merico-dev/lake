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

	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/domainlayer/devops"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/bamboo/api"
	"github.com/apache/incubator-devlake/plugins/bamboo/models"
	"github.com/apache/incubator-devlake/plugins/bamboo/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/bamboo/tasks"
)

// make sure interface is implemented
var _ interface {
	plugin.PluginMeta
	plugin.PluginInit
	plugin.PluginTask
	plugin.PluginModel
	plugin.PluginMigration
	plugin.DataSourcePluginBlueprintV200
	plugin.CloseablePluginTask
	plugin.PluginSource
} = (*Bamboo)(nil)

type Bamboo struct{}

func (p Bamboo) Init(br context.BasicRes) errors.Error {
	api.Init(br, p)

	return nil
}

func (p Bamboo) Connection() dal.Tabler {
	return &models.BambooConnection{}
}

func (p Bamboo) Scope() plugin.ToolLayerScope {
	return &models.BambooPlan{}
}

func (p Bamboo) ScopeConfig() dal.Tabler {
	return &models.BambooScopeConfig{}
}

func (p Bamboo) MakeDataSourcePipelinePlanV200(connectionId uint64, scopes []*plugin.BlueprintScopeV200, syncPolicy plugin.BlueprintSyncPolicy) (plugin.PipelinePlan, []plugin.Scope, errors.Error) {
	return api.MakePipelinePlanV200(p.SubTaskMetas(), connectionId, scopes, &syncPolicy)
}

func (p Bamboo) GetTablesInfo() []dal.Tabler {
	return []dal.Tabler{
		&models.BambooConnection{},
		&models.BambooPlan{},
		&models.BambooJob{},
		&models.BambooPlanBuild{},
		&models.BambooPlanBuildVcsRevision{},
		&models.BambooJobBuild{},
		&models.BambooDeployBuild{},
		&models.BambooDeployEnvironment{},
		&models.BambooScopeConfig{},
	}
}

func (p Bamboo) Description() string {
	return "collect some Bamboo data"
}

func (p Bamboo) Name() string {
	return "bamboo"
}

func (p Bamboo) SubTaskMetas() []plugin.SubTaskMeta {
	// TODO add your sub task here
	return []plugin.SubTaskMeta{
		tasks.ConvertPlansMeta,

		tasks.CollectJobMeta,
		tasks.ExtractJobMeta,

		tasks.CollectPlanBuildMeta,
		tasks.ExtractPlanBuildMeta,

		tasks.CollectJobBuildMeta,
		tasks.ExtractJobBuildMeta,

		tasks.CollectDeployMeta,
		tasks.ExtractDeployMeta,

		tasks.CollectDeployBuildMeta,
		tasks.ExtractDeployBuildMeta,

		tasks.ConvertJobBuildsMeta,
		tasks.ConvertPlanBuildsMeta,
		tasks.ConvertPlanVcsMeta,
		tasks.ConvertDeployBuildsMeta,
	}
}

func (p Bamboo) PrepareTaskData(taskCtx plugin.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	logger := taskCtx.GetLogger()
	logger.Debug("%v", options)

	op, err := tasks.DecodeAndValidateTaskOptions(options)
	if err != nil {
		return nil, err
	}
	connectionHelper := helper.NewConnectionHelper(
		taskCtx,
		nil,
		p.Name(),
	)
	connection := &models.BambooConnection{}
	err = connectionHelper.FirstById(connection, op.ConnectionId)
	if err != nil {
		return nil, errors.Default.Wrap(err, "unable to get Bamboo connection by the given connection ID")
	}

	apiClient, err := tasks.NewBambooApiClient(taskCtx, connection)
	if err != nil {
		return nil, errors.Default.Wrap(err, "unable to get Bamboo API client instance")
	}
	if op.PlanKey != "" {
		var scope *models.BambooPlan
		// support v100 & advance mode
		// If we still cannot find the record in db, we have to request from remote server and save it to db
		db := taskCtx.GetDal()
		err = db.First(&scope, dal.Where("connection_id = ? AND plan_key = ?", op.ConnectionId, op.PlanKey))
		if err != nil {
			return nil, err
		}

		op.ScopeConfigId = scope.ScopeConfigId
		if err != nil {
			return nil, errors.Default.Wrap(err, fmt.Sprintf("fail to find plan: %s", op.PlanKey))
		}
	}

	if op.BambooScopeConfig == nil && op.ScopeConfigId != 0 {
		var scopeConfig models.BambooScopeConfig
		db := taskCtx.GetDal()
		err = db.First(&scopeConfig, dal.Where("id = ?", op.ScopeConfigId))
		if err != nil {
			if db.IsErrorNotFound(err) {
				return nil, errors.Default.Wrap(err, fmt.Sprintf("can not find scopeConfig by scopeConfigId [%d]", op.ScopeConfigId))
			}
			return nil, errors.Default.Wrap(err, fmt.Sprintf("fail to find scopeConfig by scopeConfigId [%d]", op.ScopeConfigId))
		}
		op.BambooScopeConfig = &scopeConfig
	}
	if op.BambooScopeConfig == nil && op.ScopeConfigId == 0 {
		op.BambooScopeConfig = new(models.BambooScopeConfig)
	}
	regexEnricher := helper.NewRegexEnricher()
	if err := regexEnricher.TryAdd(devops.DEPLOYMENT, op.DeploymentPattern); err != nil {
		return nil, errors.BadInput.Wrap(err, "invalid value for `deploymentPattern`")
	}
	if err := regexEnricher.TryAdd(devops.PRODUCTION, op.ProductionPattern); err != nil {
		return nil, errors.BadInput.Wrap(err, "invalid value for `productionPattern`")
	}
	if err := regexEnricher.TryAdd(models.ENV_NAME_PATTERN, op.EnvNamePattern); err != nil {
		return nil, errors.BadInput.Wrap(err, "invalid value for `envNamePattern`")
	}
	return &tasks.BambooTaskData{
		Options:       op,
		ApiClient:     apiClient,
		RegexEnricher: regexEnricher,
	}, nil
}

// PkgPath information lost when compiled as plugin(.so)
func (p Bamboo) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/bamboo"
}

func (p Bamboo) MigrationScripts() []plugin.MigrationScript {
	return migrationscripts.All()
}

func (p Bamboo) ApiResources() map[string]map[string]plugin.ApiResourceHandler {
	return map[string]map[string]plugin.ApiResourceHandler{
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
		"connections/:connectionId/scope-configs": {
			"POST": api.CreateScopeConfig,
			"GET":  api.GetScopeConfigList,
		},
		"connections/:connectionId/scope-configs/:id": {
			"PATCH":  api.UpdateScopeConfig,
			"GET":    api.GetScopeConfig,
			"DELETE": api.DeleteScopeConfig,
		},
		"connections/:connectionId/scopes": {
			"GET": api.GetScopeList,
			"PUT": api.PutScope,
		},
		"connections/:connectionId/remote-scopes": {
			"GET": api.RemoteScopes,
		},
		"connections/:connectionId/search-remote-scopes": {
			"GET": api.SearchRemoteScopes,
		},
		"connections/:connectionId/scopes/:scopeId": {
			"GET":    api.GetScope,
			"PATCH":  api.UpdateScope,
			"DELETE": api.DeleteScope,
		},
	}
}

func (p Bamboo) Close(taskCtx plugin.TaskContext) errors.Error {
	data, ok := taskCtx.GetData().(*tasks.BambooTaskData)
	if !ok {
		return errors.Default.New(fmt.Sprintf("GetData failed when try to close %+v", taskCtx))
	}
	data.ApiClient.Release()
	return nil
}
