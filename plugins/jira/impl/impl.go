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
	"encoding/json"
	goerror "errors"
	"fmt"
	"github.com/apache/incubator-devlake/plugins/jira/tasks/apiv2models"
	"io"
	"net/http"
	"time"

	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/plugins/jira/api"
	"github.com/apache/incubator-devlake/plugins/jira/models"
	"github.com/apache/incubator-devlake/plugins/jira/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/jira/tasks"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var _ core.PluginMeta = (*Jira)(nil)
var _ core.PluginInit = (*Jira)(nil)
var _ core.PluginTask = (*Jira)(nil)
var _ core.PluginApi = (*Jira)(nil)
var _ core.PluginModel = (*Jira)(nil)
var _ core.PluginMigration = (*Jira)(nil)
var _ core.PluginBlueprintV100 = (*Jira)(nil)
var _ core.CloseablePluginTask = (*Jira)(nil)
var _ core.PluginSource = (*Jira)(nil)

type Jira struct {
}

func (plugin Jira) Connection() interface{} {
	return &models.JiraConnection{}
}

func (plugin Jira) Scope() interface{} {
	return &models.JiraBoard{}
}

func (plugin Jira) TransformationRule() interface{} {
	return &models.JiraTransformationRule{}
}

func (plugin *Jira) Init(config *viper.Viper, logger core.Logger, db *gorm.DB) errors.Error {
	api.Init(config, logger, db)
	return nil
}

func (plugin Jira) GetTablesInfo() []core.Tabler {
	return []core.Tabler{
		&models.ApiMyselfResponse{},
		&models.JiraAccount{},
		&models.JiraBoard{},
		&models.JiraBoardIssue{},
		&models.JiraBoardSprint{},
		&models.JiraConnection{},
		&models.JiraIssue{},
		&models.JiraIssueChangelogItems{},
		&models.JiraIssueChangelogs{},
		&models.JiraIssueCommit{},
		&models.JiraIssueLabel{},
		&models.JiraIssueType{},
		&models.JiraProject{},
		&models.JiraRemotelink{},
		&models.JiraServerInfo{},
		&models.JiraSprint{},
		&models.JiraSprintIssue{},
		&models.JiraStatus{},
		&models.JiraWorklog{},
	}
}

func (plugin Jira) Description() string {
	return "To collect and enrich data from JIRA"
}

func (plugin Jira) SubTaskMetas() []core.SubTaskMeta {
	return []core.SubTaskMeta{
		tasks.CollectStatusMeta,
		tasks.ExtractStatusMeta,

		tasks.CollectProjectsMeta,
		tasks.ExtractProjectsMeta,

		tasks.CollectIssueTypesMeta,
		tasks.ExtractIssueTypesMeta,

		tasks.CollectIssuesMeta,
		tasks.ExtractIssuesMeta,

		tasks.CollectIssueChangelogsMeta,
		tasks.ExtractIssueChangelogsMeta,

		tasks.CollectAccountsMeta,

		tasks.CollectWorklogsMeta,
		tasks.ExtractWorklogsMeta,

		tasks.CollectRemotelinksMeta,
		tasks.ExtractRemotelinksMeta,

		tasks.CollectSprintsMeta,
		tasks.ExtractSprintsMeta,

		tasks.ConvertBoardMeta,

		tasks.ConvertIssuesMeta,

		tasks.ConvertWorklogsMeta,

		tasks.ConvertIssueChangelogsMeta,

		tasks.ConvertSprintsMeta,
		tasks.ConvertSprintIssuesMeta,

		tasks.ConvertIssueCommitsMeta,
		tasks.ConvertIssueRepoCommitsMeta,

		tasks.ExtractAccountsMeta,
		tasks.ConvertAccountsMeta,

		tasks.CollectEpicsMeta,
		tasks.ExtractEpicsMeta,
	}
}

func (plugin Jira) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	var op tasks.JiraOptions
	var err errors.Error
	db := taskCtx.GetDal()
	logger := taskCtx.GetLogger()
	logger.Debug("%v", options)
	err = helper.Decode(options, &op, nil)
	if err != nil {
		return nil, errors.Default.Wrap(err, "could not decode Jira options")
	}
	if op.ConnectionId == 0 {
		return nil, errors.BadInput.New("jira connectionId is invalid")
	}
	connection := &models.JiraConnection{}
	connectionHelper := helper.NewConnectionHelper(
		taskCtx,
		nil,
	)
	if err != nil {
		return nil, errors.BadInput.Wrap(err, "could not get connection API instance for Jira")
	}
	err = connectionHelper.FirstById(connection, op.ConnectionId)
	if err != nil {
		return nil, errors.Default.Wrap(err, "unable to get Jira connection")
	}
	jiraApiClient, err := tasks.NewJiraApiClient(taskCtx, connection)
	if err != nil {
		return nil, errors.Default.Wrap(err, "failed to create jira api client")
	}
	info, code, err := tasks.GetJiraServerInfo(jiraApiClient)
	if err != nil || code != http.StatusOK || info == nil {
		return nil, errors.HttpStatus(code).Wrap(err, "fail to get Jira server info")
	}
	if op.BoardId != 0 {
		var jiraBoard models.JiraBoard
		// get repo from db
		err = db.First(&jiraBoard, dal.Where(`connection_id = ? and board_id = ?`, connection.ID, op.BoardId))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = saveJiraScope(taskCtx, jiraApiClient, op)
				if err != nil {
					return nil, errors.Default.Wrap(err, fmt.Sprintf("fail to save scope %d", op.BoardId))
				}
			} else {
				return nil, errors.Default.Wrap(err, fmt.Sprintf("fail to save scope %d", op.BoardId))
			}
		}
	}
	if op.BoardId == 0 && op.ScopeId != "" {
		var jiraBoard models.JiraBoard
		// get repo from db
		err = db.First(&jiraBoard, dal.Where(`connection_id = ? and board_id = ?`, connection.ID, op.ScopeId))
		if err != nil {
			return nil, errors.Default.Wrap(err, fmt.Sprintf("fail to find board%s", op.ScopeId))
		}
		op.BoardId = jiraBoard.BoardId
		if op.TransformationRuleId == 0 {
			op.TransformationRuleId = jiraBoard.TransformationRuleId
		}
	}
	if op.TransformationRules == nil && op.TransformationRuleId != 0 {
		var transformationRule models.JiraTransformationRule
		err = taskCtx.GetDal().First(&transformationRule, dal.Where("id = ?", op.TransformationRuleId))
		if err != nil && !goerror.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.BadInput.Wrap(err, "fail to get transformationRule")
		}
		op.TransformationRules, err = tasks.MakeTransformationRules(transformationRule)
		if err != nil {
			return nil, errors.BadInput.Wrap(err, "fail to make transformationRule")
		}
	}

	var createdDateAfter time.Time
	if op.CreatedDateAfter != "" {
		createdDateAfter, err = errors.Convert01(time.Parse(time.RFC3339, op.CreatedDateAfter))
		if err != nil {
			return nil, errors.BadInput.Wrap(err, "invalid value for `createdDateAfter`")
		}
	}

	taskData := &tasks.JiraTaskData{
		Options:        &op,
		ApiClient:      jiraApiClient,
		JiraServerInfo: *info,
	}
	if !createdDateAfter.IsZero() {
		taskData.CreatedDateAfter = &createdDateAfter
		logger.Debug("collect data created from %s", createdDateAfter)
	}

	return taskData, nil
}

func saveJiraScope(taskCtx core.TaskContext, jiraApiClient *helper.ApiAsyncClient, op tasks.JiraOptions) errors.Error {
	res, err := jiraApiClient.Get(fmt.Sprintf("agile/1.0/board/%d", op.BoardId), nil, nil)
	if err != nil {
		return errors.HttpStatus(res.StatusCode).Wrap(err, fmt.Sprintf("fail to get Jira Board %d", op.BoardId))
	}
	var board apiv2models.Board
	if res.StatusCode != http.StatusOK {
		return errors.HttpStatus(res.StatusCode).New(fmt.Sprintf("unexpected status code when requesting repo detail from %s", res.Request.URL.String()))
	}
	defer res.Body.Close()
	body, err := errors.Convert01(io.ReadAll(res.Body))
	if err != nil {
		return err
	}
	err = errors.Convert(json.Unmarshal(body, &board))
	if err != nil {
		return err
	}
	jiraBoard := *board.ToToolLayer(op.ConnectionId)
	if jiraBoard.BoardId != 0 {
		err = taskCtx.GetDal().CreateIfNotExist(&jiraBoard)
		if err != nil {
			return err
		}
	}
	return nil
}

func (plugin Jira) MakePipelinePlan(connectionId uint64, scope []*core.BlueprintScopeV100) (core.PipelinePlan, errors.Error) {
	return api.MakePipelinePlanV100(plugin.SubTaskMetas(), connectionId, scope)
}

func (plugin Jira) MakeDataSourcePipelinePlanV200(connectionId uint64, scopes []*core.BlueprintScopeV200, syncPolicy core.BlueprintSyncPolicy) (pp core.PipelinePlan, sc []core.Scope, err errors.Error) {
	return api.MakeDataSourcePipelinePlanV200(plugin.SubTaskMetas(), connectionId, scopes, &syncPolicy)
}

func (plugin Jira) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/jira"
}

func (plugin Jira) MigrationScripts() []core.MigrationScript {
	return migrationscripts.All()
}

func (plugin Jira) ApiResources() map[string]map[string]core.ApiResourceHandler {
	return map[string]map[string]core.ApiResourceHandler{
		"test": {
			"POST": api.TestConnection,
		},
		"echo": {
			"POST": func(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
				return &core.ApiResourceOutput{Body: input.Body}, nil
			},
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
		"connections/:connectionId/scopes/:boardId": {
			"GET":   api.GetScope,
			"PATCH": api.UpdateScope,
		},
		"connections/:connectionId/scopes": {
			"GET": api.GetScopeList,
			"PUT": api.PutScope,
		},
		"transformation_rules": {
			"POST": api.CreateTransformationRule,
			"GET":  api.GetTransformationRuleList,
		},
		"transformation_rules/:id": {
			"PATCH": api.UpdateTransformationRule,
			"GET":   api.GetTransformationRule,
		},
	}
}

func (plugin Jira) Close(taskCtx core.TaskContext) errors.Error {
	data, ok := taskCtx.GetData().(*tasks.JiraTaskData)
	if !ok {
		return errors.Default.New(fmt.Sprintf("GetData failed when try to close %+v", taskCtx))
	}
	data.ApiClient.Release()
	return nil
}
