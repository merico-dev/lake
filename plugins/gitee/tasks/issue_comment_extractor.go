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

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ExtractApiIssueCommentsMeta = core.SubTaskMeta{
	Name:             "extractApiIssueComments",
	EntryPoint:       ExtractApiIssueComments,
	EnabledByDefault: true,
	Description: "Extract raw comment data  into tool layer table gitee_pull_request_comments" +
		"and gitee_issue_comments",
}

type IssueComment struct {
	GiteeId int `json:"id"`
	Body    string
	User    struct {
		Login string
		Id    int
	}
	IssueUrl       string             `json:"issue_url"`
	GiteeCreatedAt helper.Iso8601Time `json:"created_at"`
	GiteeUpdatedAt helper.Iso8601Time `json:"updated_at"`
}

func ExtractApiIssueComments(taskCtx core.SubTaskContext) error {
	data := taskCtx.GetData().(*GiteeTaskData)

	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: GiteeApiParams{
				Owner: data.Options.Owner,
				Repo:  data.Options.Repo,
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
			if apiComment.GiteeId == 0 {
				return nil, nil
			}
			//If this is a pr, ignore
			issueINumber, err := GetIssueIdByIssueUrl(apiComment.IssueUrl)
			if err != nil {
				return nil, err
			}
			issue := &models.GiteeIssue{}
			err = taskCtx.GetDb().Where("number = ? and repo_id = ?", issueINumber, data.Repo.GiteeId).Limit(1).Find(issue).Error
			if err != nil {
				return nil, err
			}
			//if we can not find issues with issue number above, move the comments to gitee_pull_request_comments
			if issue.GiteeId == 0 {
				pr := &models.GiteePullRequest{}
				err = taskCtx.GetDb().Where("number = ? and repo_id = ?", issueINumber, data.Repo.GiteeId).Limit(1).Find(pr).Error
				if err != nil {
					return nil, err
				}
				giteePrComment := &models.GiteePullRequestComment{
					GiteeId:        apiComment.GiteeId,
					PullRequestId:  pr.GiteeId,
					Body:           apiComment.Body,
					AuthorUsername: apiComment.User.Login,
					AuthorUserId:   apiComment.User.Id,
					GiteeCreatedAt: apiComment.GiteeCreatedAt.ToTime(),
					GiteeUpdatedAt: apiComment.GiteeUpdatedAt.ToTime(),
				}
				results = append(results, giteePrComment)
			} else {
				giteeIssueComment := &models.GiteeIssueComment{
					GiteeId:        apiComment.GiteeId,
					IssueId:        issue.GiteeId,
					Body:           apiComment.Body,
					AuthorUsername: apiComment.User.Login,
					AuthorUserId:   apiComment.User.Id,
					GiteeCreatedAt: apiComment.GiteeCreatedAt.ToTime(),
					GiteeUpdatedAt: apiComment.GiteeUpdatedAt.ToTime(),
				}
				results = append(results, giteeIssueComment)
			}
			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}
