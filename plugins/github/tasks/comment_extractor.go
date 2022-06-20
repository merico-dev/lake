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
	"github.com/apache/incubator-devlake/plugins/core/dal"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/github/models"
	githubUtils "github.com/apache/incubator-devlake/plugins/github/utils"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ExtractApiCommentsMeta = core.SubTaskMeta{
	Name:             "extractApiComments",
	EntryPoint:       ExtractApiComments,
	EnabledByDefault: true,
	Description: "Extract raw comment data  into tool layer table github_pull_request_comments" +
		"and github_issue_comments",
}

type IssueComment struct {
	GithubId int `json:"id"`
	Body     json.RawMessage
	User     struct {
		Login string
		Id    int
	}
	IssueUrl        string             `json:"issue_url"`
	GithubCreatedAt helper.Iso8601Time `json:"created_at"`
	GithubUpdatedAt helper.Iso8601Time `json:"updated_at"`
}

func ExtractApiComments(taskCtx core.SubTaskContext) error {
	data := taskCtx.GetData().(*GithubTaskData)

	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: GithubApiParams{
				ConnectionId: data.Options.ConnectionId,
				Owner:        data.Options.Owner,
				Repo:         data.Options.Repo,
			},
			Table: RAW_COMMENTS_TABLE,
		},
		Extract: func(row *helper.RawData) ([]interface{}, error) {
			apiComment := &IssueComment{}
			err := json.Unmarshal(row.Data, apiComment)
			if err != nil {
				return nil, err
			}
			// need to extract 2 kinds of entities here
			results := make([]interface{}, 0, 2)
			if apiComment.GithubId == 0 {
				return nil, nil
			}
			//If this is a pr, ignore
			issueINumber, err := githubUtils.GetIssueIdByIssueUrl(apiComment.IssueUrl)
			if err != nil {
				return nil, err
			}
			issue := &models.GithubIssue{}
			err = taskCtx.GetDal().All(issue, dal.Where("connection_id = ? and number = ? and repo_id = ?", data.Options.ConnectionId, issueINumber, data.Repo.GithubId))
			if err != nil {
				return nil, err
			}
			//if we can not find issues with issue number above, move the comments to github_pull_request_comments
			if issue.GithubId == 0 {
				pr := &models.GithubPullRequest{}
				err = taskCtx.GetDal().First(pr, dal.Where("connection_id = ? and number = ? and repo_id = ?", data.Options.ConnectionId, issueINumber, data.Repo.GithubId))
				if err != nil {
					return nil, err
				}
				githubPrComment := &models.GithubPullRequestComment{
					ConnectionId:    data.Options.ConnectionId,
					GithubId:        apiComment.GithubId,
					PullRequestId:   pr.GithubId,
					Body:            string(apiComment.Body),
					AuthorUsername:  apiComment.User.Login,
					AuthorUserId:    apiComment.User.Id,
					GithubCreatedAt: apiComment.GithubCreatedAt.ToTime(),
					GithubUpdatedAt: apiComment.GithubUpdatedAt.ToTime(),
				}
				results = append(results, githubPrComment)
			} else {
				githubIssueComment := &models.GithubIssueComment{
					ConnectionId:    data.Options.ConnectionId,
					GithubId:        apiComment.GithubId,
					IssueId:         issue.GithubId,
					Body:            string(apiComment.Body),
					AuthorUsername:  apiComment.User.Login,
					AuthorUserId:    apiComment.User.Id,
					GithubCreatedAt: apiComment.GithubCreatedAt.ToTime(),
					GithubUpdatedAt: apiComment.GithubUpdatedAt.ToTime(),
				}
				results = append(results, githubIssueComment)
			}
			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}
