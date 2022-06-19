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
	"github.com/apache/incubator-devlake/plugins/gitlab/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ExtractApiCommitsMeta = core.SubTaskMeta{
	Name:             "extractApiCommits",
	EntryPoint:       ExtractApiCommits,
	EnabledByDefault: true,
	Description:      "Extract raw commit data into tool layer table GitlabCommit,GitlabUser and GitlabProjectCommit",
}

func ExtractApiCommits(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_COMMIT_TABLE)

	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Extract: func(row *helper.RawData) ([]interface{}, error) {
			// need to extract 3 kinds of entities here
			results := make([]interface{}, 0, 3)

			// create gitlab commit
			gitlabApiCommit := &GitlabApiCommit{}
			err := json.Unmarshal(row.Data, gitlabApiCommit)
			if err != nil {
				return nil, err
			}
			gitlabCommit, err := ConvertCommit(gitlabApiCommit)
			if err != nil {
				return nil, err
			}

			// create project/commits relationship
			gitlabProjectCommit := &models.GitlabProjectCommit{GitlabProjectId: data.Options.ProjectId}
			gitlabProjectCommit.CommitSha = gitlabCommit.Sha

			// create gitlab user
			gitlabUserAuthor := &models.GitlabUser{}
			gitlabUserAuthor.Email = gitlabCommit.AuthorEmail
			gitlabUserAuthor.Name = gitlabCommit.AuthorName

			gitlabProjectCommit.ConnectionId = data.Options.ConnectionId
			gitlabUserAuthor.ConnectionId = data.Options.ConnectionId
			results = append(results, gitlabCommit)
			results = append(results, gitlabProjectCommit)
			results = append(results, gitlabUserAuthor)

			// For Commiter Email is not same as AuthorEmail
			if gitlabCommit.CommitterEmail != gitlabUserAuthor.Email {
				gitlabUserCommitter := &models.GitlabUser{}
				gitlabUserCommitter.Email = gitlabCommit.CommitterEmail
				gitlabUserCommitter.Name = gitlabCommit.CommitterName
				gitlabUserCommitter.ConnectionId = data.Options.ConnectionId
				results = append(results, gitlabUserCommitter)
			}

			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}

// Convert the API response to our DB model instance
func ConvertCommit(commit *GitlabApiCommit) (*models.GitlabCommit, error) {
	gitlabCommit := &models.GitlabCommit{
		Sha:            commit.GitlabId,
		Title:          commit.Title,
		Message:        commit.Message,
		ShortId:        commit.ShortId,
		AuthorName:     commit.AuthorName,
		AuthorEmail:    commit.AuthorEmail,
		AuthoredDate:   commit.AuthoredDate.ToTime(),
		CommitterName:  commit.CommitterName,
		CommitterEmail: commit.CommitterEmail,
		CommittedDate:  commit.CommittedDate.ToTime(),
		WebUrl:         commit.WebUrl,
		Additions:      commit.Stats.Additions,
		Deletions:      commit.Stats.Deletions,
		Total:          commit.Stats.Total,
	}
	return gitlabCommit, nil
}
