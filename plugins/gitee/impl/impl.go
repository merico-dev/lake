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

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/api"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/gitee/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/gitee/tasks"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var _ core.PluginMeta = (*Gitee)(nil)
var _ core.PluginInit = (*Gitee)(nil)
var _ core.PluginTask = (*Gitee)(nil)
var _ core.PluginApi = (*Gitee)(nil)
var _ core.PluginMigration = (*Gitee)(nil)
var _ core.CloseablePluginTask = (*Gitee)(nil)

type Gitee string

func (plugin Gitee) Init(config *viper.Viper, logger core.Logger, db *gorm.DB) errors.Error {
	api.Init(config, logger, db)
	return nil
}

func (plugin Gitee) GetTablesInfo() []core.Tabler {
	return []core.Tabler{
		&models.GiteeConnection{},
		&models.GiteeAccount{},
		&models.GiteeCommit{},
		&models.GiteeCommitStat{},
		&models.GiteeIssue{},
		&models.GiteeIssueComment{},
		&models.GiteeIssueLabel{},
		&models.GiteePullRequest{},
		&models.GiteePullRequestComment{},
		&models.GiteePullRequestCommit{},
		&models.GiteePullRequestIssue{},
		&models.GiteePullRequestLabel{},
		&models.GiteeRepo{},
		&models.GiteeRepoCommit{},
		&models.GiteeResponse{},
		&models.GiteeReviewer{},
	}
}

func (plugin Gitee) Description() string {
	return "To collect and enrich data from Gitee"
}

func (plugin Gitee) SubTaskMetas() []core.SubTaskMeta {
	return []core.SubTaskMeta{
		tasks.CollectApiRepoMeta,
		tasks.ExtractApiRepoMeta,
		tasks.CollectApiIssuesMeta,
		tasks.ExtractApiIssuesMeta,
		tasks.CollectCommitsMeta,
		tasks.ExtractCommitsMeta,
		tasks.CollectApiPullRequestsMeta,
		tasks.ExtractApiPullRequestsMeta,
		tasks.CollectApiIssueCommentsMeta,
		tasks.ExtractApiIssueCommentsMeta,
		tasks.CollectApiPullRequestCommitsMeta,
		tasks.ExtractApiPullRequestCommitsMeta,
		tasks.CollectApiPullRequestReviewsMeta,
		tasks.ExtractApiPullRequestReviewsMeta,
		tasks.CollectApiCommitStatsMeta,
		tasks.ExtractApiCommitStatsMeta,
		tasks.EnrichPullRequestIssuesMeta,
		tasks.ConvertRepoMeta,
		tasks.ConvertIssuesMeta,
		tasks.ConvertCommitsMeta,
		tasks.ConvertIssueLabelsMeta,
		tasks.ConvertPullRequestCommitsMeta,
		tasks.ConvertPullRequestsMeta,
		tasks.ConvertPullRequestLabelsMeta,
		tasks.ConvertPullRequestIssuesMeta,
		tasks.ConvertAccountsMeta,
		tasks.ConvertIssueCommentsMeta,
		tasks.ConvertPullRequestCommentsMeta,
		tasks.ConvertPullRequestsMeta,
	}
}

func (plugin Gitee) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	var op tasks.GiteeOptions
	var err errors.Error
	err = helper.Decode(options, &op, nil)
	if err != nil {
		return nil, err
	}

	if op.Owner == "" {
		return nil, errors.BadInput.New("owner is required for Gitee execution")
	}

	if op.Repo == "" {
		return nil, errors.BadInput.New("repo is required for Gitee execution")
	}

	if op.PrType == "" {
		op.PrType = "type/(.*)$"
	}

	if op.PrComponent == "" {
		op.PrComponent = "component/(.*)$"
	}

	if op.IssueSeverity == "" {
		op.IssueSeverity = "severity/(.*)$"
	}

	if op.IssuePriority == "" {
		op.IssuePriority = "^(highest|high|medium|low)$"
	}

	if op.IssueComponent == "" {
		op.IssueComponent = "component/(.*)$"
	}

	if op.IssueTypeBug == "" {
		op.IssueTypeBug = "^(bug|failure|error)$"
	}

	if op.IssueTypeIncident == "" {
		op.IssueTypeIncident = ""
	}

	if op.IssueTypeRequirement == "" {
		op.IssueTypeRequirement = "^(feat|feature|proposal|requirement)$"
	}

	if op.ConnectionId == 0 {
		return nil, errors.BadInput.New("connectionId is invalid")
	}

	connection := &models.GiteeConnection{}
	connectionHelper := helper.NewConnectionHelper(
		taskCtx,
		nil,
	)

	if err != nil {
		return nil, err
	}

	err = connectionHelper.FirstById(connection, op.ConnectionId)

	if err != nil {
		return nil, err
	}
	apiClient, err := tasks.NewGiteeApiClient(taskCtx, connection)

	if err != nil {
		return nil, err
	}

	return &tasks.GiteeTaskData{
		Options:   &op,
		ApiClient: apiClient,
	}, nil
}

func (plugin Gitee) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/gitee"
}

func (plugin Gitee) MigrationScripts() []core.MigrationScript {
	return migrationscripts.All()
}

func (plugin Gitee) ApiResources() map[string]map[string]core.ApiResourceHandler {
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

func (plugin Gitee) MakePipelinePlan(connectionId uint64, scope []*core.BlueprintScopeV100) (core.PipelinePlan, errors.Error) {
	return api.MakePipelinePlan(plugin.SubTaskMetas(), connectionId, scope)
}

func (plugin Gitee) Close(taskCtx core.TaskContext) errors.Error {
	data, ok := taskCtx.GetData().(*tasks.GiteeTaskData)
	if !ok {
		return errors.Default.New(fmt.Sprintf("GetData failed when try to close %+v", taskCtx))
	}
	data.ApiClient.Release()
	return nil
}
