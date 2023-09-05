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
	"time"

	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	coreModels "github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/tapd/api"
	"github.com/apache/incubator-devlake/plugins/tapd/models"
	"github.com/apache/incubator-devlake/plugins/tapd/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/tapd/tasks"
)

var _ interface {
	plugin.PluginMeta
	plugin.PluginInit
	plugin.PluginTask
	plugin.PluginApi
	plugin.PluginModel
	plugin.PluginMigration
	plugin.CloseablePluginTask
	plugin.PluginSource
} = (*Tapd)(nil)

type Tapd struct{}

func (p Tapd) Init(basicRes context.BasicRes) errors.Error {
	api.Init(basicRes, p)

	return nil
}

func (p Tapd) Connection() dal.Tabler {
	return &models.TapdConnection{}
}

func (p Tapd) Scope() plugin.ToolLayerScope {
	return &models.TapdWorkspace{}
}

func (p Tapd) ScopeConfig() dal.Tabler {
	return &models.TapdScopeConfig{}
}

func (p Tapd) GetTablesInfo() []dal.Tabler {
	return []dal.Tabler{
		&models.TapdAccount{},
		&models.TapdBug{},
		&models.TapdBugChangelog{},
		&models.TapdBugChangelogItem{},
		&models.TapdBugCommit{},
		&models.TapdBugCustomFields{},
		&models.TapdBugLabel{},
		&models.TapdBugStatus{},
		&models.TapdConnection{},
		&models.TapdIteration{},
		&models.TapdIterationBug{},
		&models.TapdIterationStory{},
		&models.TapdIterationTask{},
		&models.TapdStory{},
		&models.TapdStoryBug{},
		&models.TapdStoryCategory{},
		&models.TapdStoryChangelog{},
		&models.TapdStoryChangelogItem{},
		&models.TapdStoryCommit{},
		&models.TapdStoryCustomFields{},
		&models.TapdStoryLabel{},
		&models.TapdStoryStatus{},
		&models.TapdTask{},
		&models.TapdTaskChangelog{},
		&models.TapdTaskChangelogItem{},
		&models.TapdTaskCommit{},
		&models.TapdTaskCustomFields{},
		&models.TapdTaskLabel{},
		&models.TapdWorkSpaceBug{},
		&models.TapdWorkSpaceStory{},
		&models.TapdWorkSpaceTask{},
		&models.TapdWorklog{},
		&models.TapdWorkspace{},
		&models.TapdWorkspaceIteration{},
		&models.TapdStoryCustomFieldValue{},
		&models.TapdTaskCustomFieldValue{},
		&models.TapdBugCustomFieldValue{},
		&models.TapdScopeConfig{},
		&models.TapdWorkitemType{},
	}
}

func (p Tapd) Description() string {
	return "To collect and enrich data from Tapd"
}

func (p Tapd) Name() string {
	return "tapd"
}

func (p Tapd) SubTaskMetas() []plugin.SubTaskMeta {
	return []plugin.SubTaskMeta{
		tasks.ConvertWorkspaceMeta,
		tasks.CollectWorkitemTypesMeta,
		tasks.ExtractWorkitemTypesMeta,
		tasks.CollectStoryCustomFieldsMeta,
		tasks.ExtractStoryCustomFieldsMeta,
		tasks.CollectTaskCustomFieldsMeta,
		tasks.ExtractTaskCustomFieldsMeta,
		tasks.CollectBugCustomFieldsMeta,
		tasks.ExtractBugCustomFieldsMeta,
		tasks.CollectStoryCategoriesMeta,
		tasks.ExtractStoryCategoriesMeta,
		tasks.CollectStoryStatusMeta,
		tasks.ExtractStoryStatusMeta,
		tasks.CollectStoryStatusLastStepMeta,
		tasks.EnrichStoryStatusLastStepMeta,
		tasks.CollectBugStatusMeta,
		tasks.ExtractBugStatusMeta,
		tasks.CollectBugStatusLastStepMeta,
		tasks.EnrichBugStatusLastStepMeta,
		tasks.CollectAccountsMeta,
		tasks.ExtractAccountsMeta,
		tasks.CollectIterationMeta,
		tasks.ExtractIterationMeta,
		tasks.CollectStoryMeta,
		tasks.CollectBugMeta,
		tasks.CollectTaskMeta,
		tasks.ExtractStoryMeta,
		tasks.ExtractBugMeta,
		tasks.ExtractTaskMeta,
		tasks.CollectBugChangelogMeta,
		tasks.ExtractBugChangelogMeta,
		tasks.CollectStoryChangelogMeta,
		tasks.ExtractStoryChangelogMeta,
		tasks.CollectTaskChangelogMeta,
		tasks.ExtractTaskChangelogMeta,
		tasks.CollectWorklogMeta,
		tasks.ExtractWorklogMeta,
		tasks.CollectBugCommitMeta,
		tasks.ExtractBugCommitMeta,
		tasks.CollectStoryCommitMeta,
		tasks.ExtractStoryCommitMeta,
		tasks.CollectTaskCommitMeta,
		tasks.ExtractTaskCommitMeta,
		tasks.CollectStoryBugMeta,
		tasks.ExtractStoryBugsMeta,
		tasks.ConvertAccountsMeta,
		tasks.ConvertIterationMeta,
		tasks.ConvertStoryMeta,
		tasks.ConvertBugMeta,
		tasks.ConvertTaskMeta,
		tasks.ConvertWorklogMeta,
		tasks.ConvertBugChangelogMeta,
		tasks.ConvertStoryChangelogMeta,
		tasks.ConvertTaskChangelogMeta,
		tasks.ConvertBugCommitMeta,
		tasks.ConvertStoryCommitMeta,
		tasks.ConvertTaskCommitMeta,
		tasks.ConvertStoryLabelsMeta,
		tasks.ConvertTaskLabelsMeta,
		tasks.ConvertBugLabelsMeta,
		tasks.EnrichStoryCustomFieldMeta,
		tasks.EnrichBugCustomFieldMeta,
		tasks.EnrichTaskCustomFieldMeta,
	}
}

func (p Tapd) PrepareTaskData(taskCtx plugin.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	logger := taskCtx.GetLogger()
	logger.Debug("%v", options)
	op, err := tasks.DecodeAndValidateTaskOptions(options)
	if err != nil {
		return nil, err
	}
	connection := &models.TapdConnection{}
	connectionHelper := helper.NewConnectionHelper(
		taskCtx,
		nil,
		p.Name(),
	)
	err = connectionHelper.FirstById(connection, op.ConnectionId)
	if err != nil {
		return nil, err
	}
	if connection.RateLimitPerHour == 0 {
		connection.RateLimitPerHour = 3600
	}
	tapdApiClient, err := tasks.NewTapdApiClient(taskCtx, connection)
	if err != nil {
		return nil, errors.Default.Wrap(err, "failed to create tapd api client")
	}

	if op.WorkspaceId != 0 {
		var scope *models.TapdWorkspace
		// support v100 & advance mode
		// If we still cannot find the record in db, we have to request from remote server and save it to db
		db := taskCtx.GetDal()
		err = db.First(&scope, dal.Where("connection_id = ? AND id = ?", op.ConnectionId, op.WorkspaceId))
		if err != nil && db.IsErrorNotFound(err) {
			scope, err = api.GetApiWorkspace(op, tapdApiClient)
			if err != nil {
				return nil, err
			}
			logger.Debug(fmt.Sprintf("Current workspace: %d", scope.Id))
			err = db.CreateIfNotExist(&scope)
			if err != nil {
				return nil, err
			}
		}
		if err != nil {
			return nil, errors.Default.Wrap(err, fmt.Sprintf("fail to find workspace: %d", op.WorkspaceId))
		}
		op.WorkspaceId = scope.Id
		if op.ScopeConfigId == 0 {
			op.ScopeConfigId = scope.ScopeConfigId
		}
	}

	if op.ScopeConfig == nil && op.ScopeConfigId != 0 {
		err = taskCtx.GetDal().First(&op.ScopeConfig, dal.Where("id = ?", op.ScopeConfigId))
		if err != nil && taskCtx.GetDal().IsErrorNotFound(err) {
			return nil, errors.BadInput.Wrap(err, "fail to get scopeConfig")
		}
	}

	if op.PageSize == 0 {
		op.PageSize = 100
	}
	cstZone, err1 := time.LoadLocation("Asia/Shanghai")
	if err1 != nil {
		return nil, errors.Default.Wrap(err1, "fail to get CST Location")
	}
	op.CstZone = cstZone
	taskData := &tasks.TapdTaskData{
		Options:    op,
		ApiClient:  tapdApiClient,
		Connection: connection,
	}
	if op.TimeAfter != "" {
		var timeAfter time.Time
		timeAfter, err = errors.Convert01(time.Parse(time.RFC3339, op.TimeAfter))
		if err != nil {
			return nil, errors.BadInput.Wrap(err, "invalid value for `timeAfter`")
		}
		taskData.TimeAfter = &timeAfter
		logger.Debug("collect data updated timeAfter %s", timeAfter)
	}
	return taskData, nil
}

func (p Tapd) MakeDataSourcePipelinePlanV200(
	connectionId uint64,
	scopes []*coreModels.BlueprintScope,
	syncPolicy coreModels.BlueprintSyncPolicy,
) (pp coreModels.PipelinePlan, sc []plugin.Scope, err errors.Error) {
	return api.MakeDataSourcePipelinePlanV200(p.SubTaskMetas(), connectionId, scopes, &syncPolicy)
}

func (p Tapd) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/tapd"
}

func (p Tapd) MigrationScripts() []plugin.MigrationScript {
	return migrationscripts.All()
}

func (p Tapd) ApiResources() map[string]map[string]plugin.ApiResourceHandler {
	return map[string]map[string]plugin.ApiResourceHandler{
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
		"connections/:connectionId/proxy/rest/*path": {
			"GET": api.Proxy,
		},
		"connections/:connectionId/scopes/:scopeId": {
			"GET":    api.GetScope,
			"PATCH":  api.UpdateScope,
			"DELETE": api.DeleteScope,
		},
		"connections/:connectionId/remote-scopes-prepare-token": {
			"GET": api.PrepareFirstPageToken,
		},
		"connections/:connectionId/remote-scopes": {
			"GET": api.RemoteScopes,
		},
		"connections/:connectionId/scopes": {
			"GET": api.GetScopeList,
			"PUT": api.PutScope,
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
	}
}

func (p Tapd) Close(taskCtx plugin.TaskContext) errors.Error {
	data, ok := taskCtx.GetData().(*tasks.TapdTaskData)
	if !ok {
		return errors.Default.New(fmt.Sprintf("GetData failed when try to close %+v", taskCtx))
	}
	data.ApiClient.Release()
	return nil
}
