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

package main

import (
	"context"
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/github/models"
	githubTasks "github.com/apache/incubator-devlake/plugins/github/tasks"
	"github.com/apache/incubator-devlake/plugins/github_graphql/tasks"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/runner"
	"github.com/merico-dev/graphql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"reflect"
	"strings"
	"time"
)

// make sure interface is implemented
var _ core.PluginMeta = (*GithubGraphql)(nil)
var _ core.PluginInit = (*GithubGraphql)(nil)
var _ core.PluginTask = (*GithubGraphql)(nil)
var _ core.PluginApi = (*GithubGraphql)(nil)

// PluginEntry exports a symbol for Framework to load
var PluginEntry GithubGraphql //nolint

type GithubGraphql struct{}

func (plugin GithubGraphql) Description() string {
	return "collect some GithubGraphql data"
}

func (plugin GithubGraphql) Init(config *viper.Viper, logger core.Logger, db *gorm.DB) errors.Error {
	return nil
}

func (plugin GithubGraphql) SubTaskMetas() []core.SubTaskMeta {
	return []core.SubTaskMeta{
		tasks.CollectRepoMeta,
		tasks.CollectIssueMeta,
		tasks.CollectPrMeta,

		githubTasks.CollectApiCommentsMeta,
		githubTasks.ExtractApiCommentsMeta,
		githubTasks.CollectApiEventsMeta,
		githubTasks.ExtractApiEventsMeta,
		githubTasks.CollectMilestonesMeta,
		githubTasks.ExtractMilestonesMeta,
		githubTasks.CollectApiPrReviewCommentsMeta,
		githubTasks.ExtractApiPrReviewCommentsMeta,

		tasks.CollectAccountMeta,
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

func (plugin GithubGraphql) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	var op githubTasks.GithubOptions
	err := helper.Decode(options, &op, nil)
	if err != nil {
		return nil, err
	}
	if op.Owner == "" {
		return nil, errors.Default.New("owner is required for GitHub execution")
	}
	if op.Repo == "" {
		return nil, errors.Default.New("repo is required for GitHub execution")
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

	tokens := strings.Split(connection.Token, ",")
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tokens[0]},
	)
	httpClient := oauth2.NewClient(taskCtx.GetContext(), src)
	client := graphql.NewClient(connection.Endpoint+`graphql`, httpClient)
	graphqlClient, err := helper.CreateAsyncGraphqlClient(taskCtx, client, taskCtx.GetLogger(),
		func(ctx context.Context, client *graphql.Client, logger core.Logger) (rateRemaining int, resetAt *time.Time, err errors.Error) {
			var query GraphQueryRateLimit
			dataErrors, err := errors.Convert01(client.Query(taskCtx.GetContext(), &query, nil))
			if err != nil {
				return 0, nil, err
			}
			if dataErrors != nil && len(dataErrors) > 0 {
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

	apiClient, err := githubTasks.CreateApiClient(taskCtx, connection)
	if err != nil {
		return nil, errors.Default.Wrap(err, "unable to get github API client instance")
	}

	return &githubTasks.GithubTaskData{
		Options:       &op,
		ApiClient:     apiClient,
		GraphqlClient: graphqlClient,
	}, nil
}

// PkgPath information lost when compiled as plugin(.so)
func (plugin GithubGraphql) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/githubGraphql"
}

func (plugin GithubGraphql) ApiResources() map[string]map[string]core.ApiResourceHandler {
	return nil
}

func (plugin GithubGraphql) Close(taskCtx core.TaskContext) errors.Error {
	data, ok := taskCtx.GetData().(*githubTasks.GithubTaskData)
	if !ok {
		return errors.Default.New(fmt.Sprintf("GetData failed when try to close %+v", taskCtx))
	}
	data.ApiClient.Release()
	data.GraphqlClient.Release()
	return nil
}

// standalone mode for debugging
func main() {
	cmd := &cobra.Command{Use: "githubGraphql"}
	connectionId := cmd.Flags().Uint64P("connectionId", "c", 0, "github connection id")
	owner := cmd.Flags().StringP("owner", "o", "", "github owner")
	repo := cmd.Flags().StringP("repo", "r", "", "github repo")
	_ = cmd.MarkFlagRequired("connectionId")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		runner.DirectRun(cmd, args, PluginEntry, map[string]interface{}{
			"connectionId": *connectionId,
			"owner":        *owner,
			"repo":         *repo,
		})
	}
	runner.RunCmd(cmd)
}
