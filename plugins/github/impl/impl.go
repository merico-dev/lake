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
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"strings"
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
var _ core.PluginModel = (*Github)(nil)
var _ core.PluginBlueprintV100 = (*Github)(nil)
var _ core.CloseablePluginTask = (*Github)(nil)
var _ core.PluginSource = (*Github)(nil)

type Github struct{}

func (plugin Github) Connection() interface{} {
	return &models.GithubConnection{}
}

func (plugin Github) Scope() interface{} {
	return &models.GithubRepo{}
}

func (plugin Github) TransformationRule() interface{} {
	return &models.GithubTransformationRule{}
}

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
	op, err := tasks.DecodeTaskOptions(options)
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
	apiClient, err := tasks.CreateApiClient(taskCtx, connection)
	if err != nil {
		return nil, errors.Default.Wrap(err, "unable to get github API client instance")
	}
	err = EnrichOptions(taskCtx, op, apiClient.ApiClient)
	if err != nil {
		return nil, err
	}

	var createdDateAfter time.Time
	if op.CreatedDateAfter != "" {
		createdDateAfter, err = errors.Convert01(time.Parse(time.RFC3339, op.CreatedDateAfter))
		if err != nil {
			return nil, errors.BadInput.Wrap(err, "invalid value for `createdDateAfter`")
		}
	}
	taskData := &tasks.GithubTaskData{
		Options:   op,
		ApiClient: apiClient,
	}

	if !createdDateAfter.IsZero() {
		taskData.CreatedDateAfter = &createdDateAfter
		logger.Debug("collect data updated createdDateAfter %s", createdDateAfter)
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
		"connections/:connectionId/scopes/:repoId": {
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
		"connections/:connectionId/proxy/rest/*path": {
			"GET": api.Proxy,
		},
	}
}

func (plugin Github) MakePipelinePlan(connectionId uint64, scope []*core.BlueprintScopeV100) (core.PipelinePlan, errors.Error) {
	return api.MakePipelinePlan(plugin.SubTaskMetas(), connectionId, scope)
}

func (plugin Github) MakeDataSourcePipelinePlanV200(connectionId uint64, scopes []*core.BlueprintScopeV200, syncPolicy core.BlueprintSyncPolicy) (pp core.PipelinePlan, sc []core.Scope, err errors.Error) {
	return api.MakeDataSourcePipelinePlanV200(plugin.SubTaskMetas(), connectionId, scopes, &syncPolicy)
}

func (plugin Github) Close(taskCtx core.TaskContext) errors.Error {
	data, ok := taskCtx.GetData().(*tasks.GithubTaskData)
	if !ok {
		return errors.Default.New(fmt.Sprintf("GetData failed when try to close %+v", taskCtx))
	}
	data.ApiClient.Release()
	return nil
}

func EnrichOptions(taskCtx core.TaskContext,
	op *tasks.GithubOptions,
	apiClient *helper.ApiClient) errors.Error {
	var githubRepo models.GithubRepo
	var err errors.Error
	log := taskCtx.GetLogger()
	// for advanced mode or others which we already set value to onwer/repo
	if op.Owner != "" && op.Repo != "" {
		// Lets try to use owner_login/name to find the record first
		// In our db._tool_github_repos, record might be two kinds:
		// 1. owner_login  = op.owner and name = op.repo
		// 2. name = op.owner/op.repo
		err := taskCtx.GetDal().First(&githubRepo, dal.Where(
			"connection_id = ? AND ((name = ? AND owner_login = ?) OR name = ?)",
			op.ConnectionId, op.Repo, op.Owner, fmt.Sprintf("%s/%s", op.Owner, op.Repo)))
		if err == nil {
			op.GithubId = githubRepo.GithubId
			op.TransformationRuleId = githubRepo.TransformationRuleId
		}
		// If we still cannot find the record in db, we have to request from remote server and save it to db
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				var repo *tasks.GithubApiRepo
				repo, err = api.MemorizedGetApiRepo(repo, op, apiClient)
				if err != nil {
					return err
				}
				log.Debug(fmt.Sprintf("Current repo: %s", repo.FullName))
				var scope models.GithubRepo
				scope.ConnectionId = op.ConnectionId
				scope.GithubId = repo.GithubId
				scope.CreatedDate = repo.CreatedAt.ToNullableTime()
				scope.Language = repo.Language
				scope.Description = repo.Description
				scope.HTMLUrl = repo.HTMLUrl
				scope.ConnectionId = op.ConnectionId
				scope.Name = repo.FullName
				err = taskCtx.GetDal().CreateIfNotExist(&scope)
				if err != nil {
					return err
				}

				op.GithubId = repo.GithubId
			} else {
				return errors.Default.Wrap(err, fmt.Sprintf("fail to find repo %s/%s", op.Owner, op.Repo))
			}
		}
	}
	// for bp v200 which we only set ScopeId for options
	if githubRepo.GithubId == 0 && op.ScopeId != "" {
		log.Debug(fmt.Sprintf("Getting githubRepo by op.ScopeId: %s", op.ScopeId))
		err = taskCtx.GetDal().First(&githubRepo, dal.Where(`connection_id = ? AND github_id = ?`, op.ConnectionId, op.ScopeId))
		if err != nil {
			return errors.Default.Wrap(err, fmt.Sprintf("fail to find repo %s", op.ScopeId))
		}
		ownerName := strings.Split(githubRepo.Name, "/")
		if len(ownerName) != 2 {
			return errors.Default.New("Fail to set owner/repo for github options.")
		}
		op.Owner = ownerName[0]
		op.Repo = ownerName[1]
		op.GithubId = githubRepo.GithubId
		op.TransformationRuleId = githubRepo.TransformationRuleId
	}
	// Set GithubTransformationRule if it's nil, this has lower priority
	if op.GithubTransformationRule == nil && op.TransformationRuleId != 0 {
		var transformationRule models.GithubTransformationRule
		err = taskCtx.GetDal().First(&transformationRule, dal.Where("id = ?", githubRepo.TransformationRuleId))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.BadInput.Wrap(err, "fail to get transformationRule")
		}
		op.GithubTransformationRule = &transformationRule
	}
	err = tasks.ValidateTaskOptions(op)
	return err
}
