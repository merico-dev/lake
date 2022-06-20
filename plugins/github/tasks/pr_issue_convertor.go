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

package tasks

import (
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"reflect"

	"github.com/apache/incubator-devlake/models/domainlayer/crossdomain"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/plugins/core"
	githubModels "github.com/apache/incubator-devlake/plugins/github/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ConvertPullRequestIssuesMeta = core.SubTaskMeta{
	Name:             "convertPullRequestIssues",
	EntryPoint:       ConvertPullRequestIssues,
	EnabledByDefault: true,
	Description:      "Convert tool layer table github_pull_request_issues into  domain layer table pull_request_issues",
}

func ConvertPullRequestIssues(taskCtx core.SubTaskContext) error {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*GithubTaskData)
	repoId := data.Repo.GithubId

	cursor, err := db.Cursor(
		dal.From(&githubModels.GithubPullRequestIssue{}),
		dal.Join(`left join _tool_github_pull_requests on _tool_github_pull_requests.github_id = _tool_github_pull_request_issues.pull_request_id`),
		dal.Where("_tool_github_pull_requests.repo_id = ? and _tool_github_pull_requests.connection_id = ?", repoId, data.Options.ConnectionId),
		dal.Orderby("pull_request_id ASC"),
	)
	if err != nil {
		return err
	}
	defer cursor.Close()
	prIdGen := didgen.NewDomainIdGenerator(&githubModels.GithubPullRequest{})
	issueIdGen := didgen.NewDomainIdGenerator(&githubModels.GithubIssue{})

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		InputRowType: reflect.TypeOf(githubModels.GithubPullRequestIssue{}),
		Input:        cursor,
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: GithubApiParams{
				ConnectionId: data.Options.ConnectionId,
				Owner:        data.Options.Owner,
				Repo:         data.Options.Repo,
			},
			Table: RAW_PULL_REQUEST_TABLE,
		},
		Convert: func(inputRow interface{}) ([]interface{}, error) {
			githubPrIssue := inputRow.(*githubModels.GithubPullRequestIssue)
			pullRequestIssue := &crossdomain.PullRequestIssue{
				PullRequestId:     prIdGen.Generate(data.Options.ConnectionId, githubPrIssue.PullRequestId),
				IssueId:           issueIdGen.Generate(data.Options.ConnectionId, githubPrIssue.IssueId),
				IssueNumber:       githubPrIssue.IssueNumber,
				PullRequestNumber: githubPrIssue.PullRequestNumber,
			}
			return []interface{}{
				pullRequestIssue,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
