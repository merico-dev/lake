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
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/tap"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/github/models"
	"github.com/apache/incubator-devlake/plugins/github_keon_singer/models/generated"
	"strconv"
)

var _ core.SubTaskEntryPoint = ExtractPr

func ExtractPr(taskCtx core.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*GithubSingerTaskData)
	extractor, err := tap.NewTapExtractor(
		&tap.ExtractorArgs[generated.PullRequests]{
			Ctx:          taskCtx,
			TapProvider:  data.Options.TapProvider,
			ConnectionId: data.Options.ConnectionId,
			StreamName:   "pull_requests",
			Extract: func(resData *generated.PullRequests) ([]interface{}, errors.Error) {
				//extractedModels := make([]interface{}, 0)
				//println(resData.Data)
				//println(resData.Input)
				// TODO decode some db models from api result
				// extractedModels = append(extractedModels, &models.XXXXXX)
				//return extractedModels, nil
				githubId, _ := strconv.Atoi(*resData.Id)
				pr := &models.GithubPullRequest{
					ConnectionId:    data.Options.ConnectionId,
					GithubId:        githubId,
					RepoId:          123,
					HeadRepoId:      0,
					Number:          0,
					State:           "DUMMY",
					Title:           "",
					GithubCreatedAt: *resData.CreatedAt,
					GithubUpdatedAt: *resData.UpdatedAt,
					ClosedAt:        nil,
					Additions:       0,
					Deletions:       0,
					Comments:        0,
					Commits:         0,
					ReviewComments:  0,
					Merged:          false,
					MergedAt:        nil,
					Body:            *resData.Body,
					Type:            "",
					Component:       "",
					MergeCommitSha:  "",
					HeadRef:         "",
					BaseRef:         "",
					BaseCommitSha:   "",
					HeadCommitSha:   "",
					Url:             "",
					AuthorName:      "",
					AuthorId:        0,
				}
				return []interface{}{pr}, nil
			},
		},
	)
	if err != nil {
		return err
	}
	return extractor.Execute()
}

var ExtractPrMeta = core.SubTaskMeta{
	Name:             "ExtractPr",
	EntryPoint:       ExtractPr,
	EnabledByDefault: true,
	Description:      "Extract raw data into tool layer table github_singer_pr",
}
