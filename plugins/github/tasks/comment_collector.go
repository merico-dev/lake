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
	"encoding/json"
	"fmt"
	. "github.com/apache/incubator-devlake/plugins/core/dal"
	"net/http"
	"net/url"

	"github.com/apache/incubator-devlake/plugins/helper"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/github/models"
)

const RAW_COMMENTS_TABLE = "github_api_comments"

// this struct should be moved to `github_api_common.go`

func CollectApiComments(taskCtx core.SubTaskContext) error {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*GithubTaskData)

	since := data.Since
	incremental := false
	if since == nil {
		var latestUpdatedIssueComt models.GithubIssueComment
		err := db.All(
			&latestUpdatedIssueComt,
			Join("left join _tool_github_issues on _tool_github_issues.github_id = _tool_github_issue_comments.issue_id"),
			Where(
				"_tool_github_issues.repo_id = ? AND _tool_github_issues.connection_id = ?", data.Repo.GithubId, data.Repo.ConnectionId,
			),
			Orderby("github_updated_at DESC"),
			Limit(1),
		)
		if err != nil {
			return fmt.Errorf("failed to get latest github issue record: %w", err)
		}
		var latestUpdatedPrComt models.GithubPullRequestComment
		err = db.All(
			&latestUpdatedPrComt,
			Join("left join _tool_github_pull_requests on _tool_github_pull_requests.github_id = _tool_github_pull_request_comments.pull_request_id"),
			Where("_tool_github_pull_requests.repo_id = ? AND _tool_github_pull_requests.connection_id = ?", data.Repo.GithubId, data.Repo.ConnectionId),
			Orderby("github_updated_at DESC"),
			Limit(1),
		)
		if err != nil {
			return fmt.Errorf("failed to get latest github issue record: %w", err)
		}
		if latestUpdatedIssueComt.GithubId > 0 && latestUpdatedPrComt.GithubId > 0 {
			if latestUpdatedIssueComt.GithubUpdatedAt.Before(latestUpdatedPrComt.GithubUpdatedAt) {
				since = &latestUpdatedPrComt.GithubUpdatedAt
			} else {
				since = &latestUpdatedIssueComt.GithubUpdatedAt
			}
			incremental = true
		} else if latestUpdatedIssueComt.GithubId > 0 {
			since = &latestUpdatedIssueComt.GithubUpdatedAt
			incremental = true
		} else if latestUpdatedPrComt.GithubId > 0 {
			since = &latestUpdatedPrComt.GithubUpdatedAt
			incremental = true
		}
	}

	collector, err := helper.NewApiCollector(helper.ApiCollectorArgs{
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: GithubApiParams{
				Owner: data.Options.Owner,
				Repo:  data.Options.Repo,
			},
			Table: RAW_COMMENTS_TABLE,
		},
		ApiClient:   data.ApiClient,
		PageSize:    100,
		Incremental: incremental,

		UrlTemplate: "repos/{{ .Params.Owner }}/{{ .Params.Repo }}/issues/comments",
		Query: func(reqData *helper.RequestData) (url.Values, error) {
			query := url.Values{}
			query.Set("state", "all")
			if since != nil {
				query.Set("since", since.String())
			}
			query.Set("page", fmt.Sprintf("%v", reqData.Pager.Page))
			query.Set("direction", "asc")
			query.Set("per_page", fmt.Sprintf("%v", reqData.Pager.Size))

			return query, nil
		},
		GetTotalPages: GetTotalPagesFromResponse,
		ResponseParser: func(res *http.Response) ([]json.RawMessage, error) {
			var items []json.RawMessage
			err := helper.UnmarshalResponse(res, &items)
			if err != nil {
				return nil, err
			}
			return items, nil
		},
	})

	if err != nil {
		return err
	}

	return collector.Execute()
}

var CollectApiCommentsMeta = core.SubTaskMeta{
	Name:             "collectApiComments",
	EntryPoint:       CollectApiComments,
	EnabledByDefault: true,
	Description:      "Collect comments data from Github api",
}
