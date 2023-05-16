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
	"context"
	"fmt"
	"github.com/apache/incubator-devlake/core/models/domainlayer/devops"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/log"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	githubImpl "github.com/apache/incubator-devlake/plugins/github/impl"
	"github.com/apache/incubator-devlake/plugins/github/models"
	githubTasks "github.com/apache/incubator-devlake/plugins/github/tasks"
	"github.com/apache/incubator-devlake/plugins/github_graphql/tasks"
	"github.com/merico-dev/graphql"
	"golang.org/x/oauth2"
)

// make sure interface is implemented
var _ plugin.PluginMeta = (*GithubGraphql)(nil)
var _ plugin.PluginTask = (*GithubGraphql)(nil)
var _ plugin.PluginApi = (*GithubGraphql)(nil)
var _ plugin.PluginModel = (*GithubGraphql)(nil)
var _ plugin.CloseablePluginTask = (*GithubGraphql)(nil)
var _ plugin.PluginSource = (*GithubGraphql)(nil)

type GithubGraphql struct{}

func (p GithubGraphql) Connection() interface{} {
	return &models.GithubConnection{}
}

func (p GithubGraphql) Scope() interface{} {
	return &models.GithubRepo{}
}

func (p GithubGraphql) TransformationRule() interface{} {
	return &models.GithubTransformationRule{}
}

func (p GithubGraphql) Description() string {
	return "collect some GithubGraphql data"
}

func (p GithubGraphql) GetTablesInfo() []dal.Tabler {
	return []dal.Tabler{}
}

func (p GithubGraphql) SubTaskMetas() []plugin.SubTaskMeta {
	return []plugin.SubTaskMeta{
		//tasks.CollectRepoMeta,

		// collect millstones
		githubTasks.CollectMilestonesMeta,
		githubTasks.ExtractMilestonesMeta,

		// collect issue & pr, deps on millstone
		tasks.CollectIssueMeta,
		tasks.CollectPrMeta,

		// collect workflow run & job
		githubTasks.CollectRunsMeta,
		githubTasks.ExtractRunsMeta,
		tasks.CollectGraphqlJobsMeta,

		// collect others
		githubTasks.CollectApiCommentsMeta,
		githubTasks.ExtractApiCommentsMeta,
		//githubTasks.CollectApiEventsMeta,
		//githubTasks.ExtractApiEventsMeta,
		githubTasks.CollectApiPrReviewCommentsMeta,
		githubTasks.ExtractApiPrReviewCommentsMeta,

		// collect account, deps on all before
		tasks.CollectAccountMeta,

		// convert to domain layer
		githubTasks.ConvertRunsMeta,
		githubTasks.ConvertJobsMeta,
		githubTasks.EnrichPullRequestIssuesMeta,
		githubTasks.ConvertRepoMeta,
		githubTasks.ConvertIssuesMeta,
		githubTasks.ConvertCommitsMeta,
		githubTasks.ConvertIssueLabelsMeta,
		githubTasks.ConvertPullRequestCommitsMeta,
		githubTasks.ConvertPullRequestsMeta,
		githubTasks.ConvertPullRequestReviewsMeta,
		githubTasks.ConvertPullRequestLabelsMeta,
		githubTasks.ConvertPullRequestIssuesMeta,
		githubTasks.ConvertIssueCommentsMeta,
		githubTasks.ConvertPullRequestCommentsMeta,
		githubTasks.ConvertMilestonesMeta,
		githubTasks.ConvertAccountsMeta,
	}
}

type GraphQueryRateLimit struct {
	RateLimit struct {
		Limit     graphql.Int
		Remaining graphql.Int
		ResetAt   time.Time
	}
}

func (p GithubGraphql) PrepareTaskData(taskCtx plugin.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	logger := taskCtx.GetLogger()
	logger.Debug("%v", options)
	var op githubTasks.GithubOptions
	err := helper.Decode(options, &op, nil)
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
		return nil, errors.Default.Wrap(err, "unable to get github connection by the given connection ID: %v")
	}

	apiClient, err := githubTasks.CreateApiClient(taskCtx, connection)
	if err != nil {
		return nil, errors.Default.Wrap(err, "unable to get github API client instance")
	}

	err = githubImpl.EnrichOptions(taskCtx, &op, apiClient.ApiClient)
	if err != nil {
		return nil, err
	}

	tokens := strings.Split(connection.Token, ",")
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tokens[0]},
	)
	httpClient := oauth2.NewClient(taskCtx.GetContext(), src)
	endpoint, err := errors.Convert01(url.JoinPath(connection.Endpoint, `graphql`))
	if err != nil {
		return nil, errors.BadInput.Wrap(err, fmt.Sprintf("malformed connection endpoint supplied: %s", connection.Endpoint))
	}
	client := graphql.NewClient(endpoint, httpClient)
	graphqlClient, err := helper.CreateAsyncGraphqlClient(taskCtx, client, taskCtx.GetLogger(),
		func(ctx context.Context, client *graphql.Client, logger log.Logger) (rateRemaining int, resetAt *time.Time, err errors.Error) {
			var query GraphQueryRateLimit
			dataErrors, err := errors.Convert01(client.Query(taskCtx.GetContext(), &query, nil))
			if err != nil {
				return 0, nil, err
			}
			if len(dataErrors) > 0 {
				return 0, nil, errors.Default.Wrap(dataErrors[0], `query rate limit fail`)
			}
			logger.Info(`github graphql init success with remaining %d/%d and will reset at %s`,
				query.RateLimit.Remaining, query.RateLimit.Limit, query.RateLimit.ResetAt)
			return int(query.RateLimit.Remaining), &query.RateLimit.ResetAt, nil
		})
	if err != nil {
		return nil, err
	}

	graphqlClient.SetGetRateCost(func(q interface{}) int {
		v := reflect.ValueOf(q)
		return int(v.Elem().FieldByName(`RateLimit`).FieldByName(`Cost`).Int())
	})

	regexEnricher := helper.NewRegexEnricher()
	if err = regexEnricher.TryAdd(devops.DEPLOYMENT, op.DeploymentPattern); err != nil {
		return nil, errors.BadInput.Wrap(err, "invalid value for `deploymentPattern`")
	}
	if err = regexEnricher.TryAdd(devops.PRODUCTION, op.ProductionPattern); err != nil {
		return nil, errors.BadInput.Wrap(err, "invalid value for `productionPattern`")
	}

	taskData := &githubTasks.GithubTaskData{
		Options:       &op,
		ApiClient:     apiClient,
		GraphqlClient: graphqlClient,
		RegexEnricher: regexEnricher,
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

// PkgPath information lost when compiled as plugin(.so)
func (p GithubGraphql) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/githubGraphql"
}

func (p GithubGraphql) ApiResources() map[string]map[string]plugin.ApiResourceHandler {
	return nil
}

func (p GithubGraphql) Close(taskCtx plugin.TaskContext) errors.Error {
	data, ok := taskCtx.GetData().(*githubTasks.GithubTaskData)
	if !ok {
		return errors.Default.New(fmt.Sprintf("GetData failed when try to close %+v", taskCtx))
	}
	data.ApiClient.Release()
	data.GraphqlClient.Release()
	return nil
}
