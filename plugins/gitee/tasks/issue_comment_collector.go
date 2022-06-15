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
	"fmt"
	"net/url"

	"github.com/apache/incubator-devlake/plugins/helper"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
)

const RAW_COMMENTS_TABLE = "gitee_issue_comments"

var CollectApiIssueCommentsMeta = core.SubTaskMeta{
	Name:             "collectApiIssueComments",
	EntryPoint:       CollectApiIssueComments,
	EnabledByDefault: true,
	Description:      "Collect comments data from Gitee api",
}

func CollectApiIssueComments(taskCtx core.SubTaskContext) error {
	db := taskCtx.GetDb()
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_COMMENTS_TABLE)

	since := data.Since
	incremental := false
	// user didn't specify a time range to sync, try load from database
	// actually, for gitee pull, since doesn't make any sense, gitee pull api doesn't support it
	if since == nil {
		var latestUpdatedIssueComment models.GiteeIssueComment
		err := db.Model(&latestUpdatedIssueComment).
			Joins("left join _tool_gitee_issues on _tool_gitee_issues.gitee_id = _tool_gitee_issue_comments.issue_id").
			Where("_tool_gitee_issues.repo_id = ?", data.Repo.GiteeId).
			Order("gitee_updated_at DESC").Limit(1).Find(&latestUpdatedIssueComment).Error
		if err != nil {
			return fmt.Errorf("failed to get latest gitee issue record: %w", err)
		}
		var latestUpdatedPrComt models.GiteePullRequestComment
		err = db.Model(&latestUpdatedPrComt).
			Joins("left join _tool_gitee_pull_requests on _tool_gitee_pull_requests.gitee_id = _tool_gitee_pull_request_comments.pull_request_id").
			Where("_tool_gitee_pull_requests.repo_id = ?", data.Repo.GiteeId).
			Order("gitee_updated_at DESC").Limit(1).Find(&latestUpdatedPrComt).Error
		if err != nil {
			return fmt.Errorf("failed to get latest gitee issue record: %w", err)
		}
		if latestUpdatedIssueComment.GiteeId > 0 && latestUpdatedPrComt.GiteeId > 0 {
			if latestUpdatedIssueComment.GiteeUpdatedAt.Before(latestUpdatedPrComt.GiteeUpdatedAt) {
				since = &latestUpdatedPrComt.GiteeUpdatedAt
			} else {
				since = &latestUpdatedIssueComment.GiteeUpdatedAt
			}
			incremental = true
		} else if latestUpdatedIssueComment.GiteeId > 0 {
			since = &latestUpdatedIssueComment.GiteeUpdatedAt
			incremental = true
		} else if latestUpdatedPrComt.GiteeId > 0 {
			since = &latestUpdatedPrComt.GiteeUpdatedAt
			incremental = true
		}

	}

	collector, err := helper.NewApiCollector(helper.ApiCollectorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		ApiClient:          data.ApiClient,
		PageSize:           100,
		Incremental:        incremental,

		UrlTemplate: "repos/{{ .Params.Owner }}/{{ .Params.Repo }}/issues/comments",
		Query: func(reqData *helper.RequestData) (url.Values, error) {
			query := url.Values{}
			query.Set("access_token", data.Options.Token)
			query.Set("state", "all")
			if since != nil {
				query.Set("since", since.String())
			}
			query.Set("page", fmt.Sprintf("%v", reqData.Pager.Page))
			query.Set("direction", "asc")
			query.Set("per_page", fmt.Sprintf("%v", reqData.Pager.Size))

			return query, nil
		},
		GetTotalPages:  GetTotalPagesFromResponse,
		ResponseParser: GetRawMessageFromResponse,
	})

	if err != nil {
		return err
	}

	return collector.Execute()
}
