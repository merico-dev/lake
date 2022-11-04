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

	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/github/api"
	"github.com/apache/incubator-devlake/plugins/github/models"
	"github.com/apache/incubator-devlake/plugins/github/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/github/tasks"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var _ core.PluginMeta = (*Github)(nil)
var _ core.PluginInit = (*Github)(nil)
var _ core.PluginTask = (*Github)(nil)
var _ core.PluginApi = (*Github)(nil)

// var _ core.Migratable = (*Github)(nil)
var _ core.PluginBlueprintV100 = (*Github)(nil)
var _ core.CloseablePluginTask = (*Github)(nil)

type Github struct{}

func (plugin Github) Init(config *viper.Viper, logger core.Logger, db *gorm.DB) errors.Error {
	api.Init(config, logger, db)
	return nil
}

func (plugin Github) GetTablesInfo() []core.Tabler {
	return []core.Tabler{
		&models.GithubConnection{},
		&models.GithubAccount{},
		&models.GithubAccountOrg{},
		&models.GithubCommit{},
		&models.GithubCommitStat{},
		&models.GithubIssue{},
		&models.GithubIssueComment{},
		&models.GithubIssueEvent{},
		&models.GithubIssueLabel{},
		&models.GithubJob{},
		&models.GithubMilestone{},
		&models.GithubPrComment{},
		&models.GithubPrCommit{},
		&models.GithubPrIssue{},
		&models.GithubPrLabel{},
		&models.GithubPrReview{},
		&models.GithubPullRequest{},
		&models.GithubRepo{},
		&models.GithubRepoAccount{},
		&models.GithubRepoCommit{},
		&models.GithubReviewer{},
		&models.GithubRun{},
	}
}

func (plugin Github) Description() string {
	return "To collect and enrich data from GitHub"
}

func (plugin Github) SubTaskMetas() []core.SubTaskMeta {
	return []core.SubTaskMeta{
		tasks.CollectApiRepoMeta,
		tasks.ExtractApiRepoMeta,
		tasks.CollectApiIssuesMeta,
		tasks.ExtractApiIssuesMeta,
		tasks.CollectApiPullRequestsMeta,
		tasks.ExtractApiPullRequestsMeta,
		tasks.CollectApiCommentsMeta,
		tasks.ExtractApiCommentsMeta,
		tasks.CollectApiEventsMeta,
		tasks.ExtractApiEventsMeta,
		tasks.CollectApiPullRequestCommitsMeta,
		tasks.ExtractApiPullRequestCommitsMeta,
		tasks.CollectApiPullRequestReviewsMeta,
		tasks.ExtractApiPullRequestReviewsMeta,
		tasks.CollectApiPrReviewCommentsMeta,
		tasks.ExtractApiPrReviewCommentsMeta,
		tasks.CollectApiCommitsMeta,
		tasks.ExtractApiCommitsMeta,
		tasks.CollectApiCommitStatsMeta,
		tasks.ExtractApiCommitStatsMeta,
		tasks.CollectMilestonesMeta,
		tasks.ExtractMilestonesMeta,
		tasks.CollectAccountsMeta,
		tasks.ExtractAccountsMeta,
		tasks.CollectAccountOrgMeta,
		tasks.ExtractAccountOrgMeta,
		tasks.CollectRunsMeta,
		tasks.ExtractRunsMeta,
		tasks.ConvertRunsMeta,
		tasks.CollectJobsMeta,
		tasks.ExtractJobsMeta,
		tasks.ConvertJobsMeta,
		tasks.EnrichPullRequestIssuesMeta,
		tasks.ConvertRepoMeta,
		tasks.ConvertIssuesMeta,
		tasks.ConvertCommitsMeta,
		tasks.ConvertIssueLabelsMeta,
		tasks.ConvertPullRequestCommitsMeta,
		tasks.ConvertPullRequestsMeta,
		tasks.ConvertPullRequestReviewsMeta,
		tasks.ConvertPullRequestLabelsMeta,
		tasks.ConvertPullRequestIssuesMeta,
		tasks.ConvertIssueCommentsMeta,
		tasks.ConvertPullRequestCommentsMeta,
		tasks.ConvertMilestonesMeta,
		tasks.ConvertAccountsMeta,
	}
}

func (plugin Github) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	logger := taskCtx.GetLogger()
	logger.Debug("%v", options)
	op, err := tasks.DecodeAndValidateTaskOptions(options)
	if err != nil {
		return nil, err
	}
	connectionHelper := helper.NewConnectionHelper(
		taskCtx,
		nil,
	)
	connection := &models.GithubConnection{}
	err = connectionHelper.FirstById(connection, op.ConnectionId)
	if err != nil {
		return nil, errors.Default.Wrap(err, "unable to get github connection by the given connection ID")
	}

	var since time.Time
	if op.Since != "" {
		since, err = errors.Convert01(time.Parse("2006-01-02T15:04:05Z", op.Since))
		if err != nil {
			return nil, errors.BadInput.Wrap(err, "invalid value for `since`")
		}
	}

	apiClient, err := tasks.CreateApiClient(taskCtx, connection)
	if err != nil {
		return nil, errors.Default.Wrap(err, "unable to get github API client instance")
	}
	taskData := &tasks.GithubTaskData{
		Options:   op,
		ApiClient: apiClient,
	}

	if !since.IsZero() {
		taskData.Since = &since
		logger.Debug("collect data updated since %s", since)
	}
	return taskData, nil
}

func (plugin Github) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/github"
}

func (plugin Github) MigrationScripts() []core.MigrationScript {
	return migrationscripts.All()
}

func (plugin Github) ApiResources() map[string]map[string]core.ApiResourceHandler {
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

func (plugin Github) MakePipelinePlan(connectionId uint64, scope []*core.BlueprintScopeV100) (core.PipelinePlan, errors.Error) {
	return api.MakePipelinePlan(plugin.SubTaskMetas(), connectionId, scope)
}

func (plugin Github) Close(taskCtx core.TaskContext) errors.Error {
	data, ok := taskCtx.GetData().(*tasks.GithubTaskData)
	if !ok {
		return errors.Default.New(fmt.Sprintf("GetData failed when try to close %+v", taskCtx))
	}
	data.ApiClient.Release()
	return nil
}
