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

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ExtractApiRepoMeta = core.SubTaskMeta{
	Name:        "extractApiRepo",
	EntryPoint:  ExtractApiRepositories,
	Required:    true,
	Description: "Extract raw Repositories data into tool layer table gitee_repos",
}

type GiteeApiRepoResponse struct {
	Name        string                `json:"name"`
	GiteeId     int                   `json:"id"`
	HTMLUrl     string                `json:"html_url"`
	Language    string                `json:"language"`
	Description string                `json:"description"`
	Owner       models.GiteeUser      `json:"owner"`
	Parent      *GiteeApiRepoResponse `json:"parent"`
	CreatedAt   helper.Iso8601Time    `json:"created_at"`
	UpdatedAt   *helper.Iso8601Time   `json:"updated_at"`
}

func ExtractApiRepositories(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_REPOSITORIES_TABLE)
	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Extract: func(row *helper.RawData) ([]interface{}, error) {
			repo := &GiteeApiRepoResponse{}
			err := json.Unmarshal(row.Data, repo)
			if err != nil {
				return nil, err
			}
			if repo.GiteeId == 0 {
				return nil, fmt.Errorf("repo %s/%s not found", data.Options.Owner, data.Options.Repo)
			}
			results := make([]interface{}, 0, 1)
			giteeRepository := &models.GiteeRepo{
				GiteeId:     repo.GiteeId,
				Name:        repo.Name,
				HTMLUrl:     repo.HTMLUrl,
				Description: repo.Description,
				OwnerId:     repo.Owner.Id,
				OwnerLogin:  repo.Owner.Login,
				Language:    repo.Language,
				CreatedDate: repo.CreatedAt.ToTime(),
				UpdatedDate: helper.Iso8601TimeToTime(repo.UpdatedAt),
			}
			data.Repo = giteeRepository

			if repo.Parent != nil {
				giteeRepository.ParentGiteeId = repo.Parent.GiteeId
				giteeRepository.ParentHTMLUrl = repo.Parent.HTMLUrl
			}
			results = append(results, giteeRepository)
			taskCtx.TaskContext().GetData().(*GiteeTaskData).Repo = giteeRepository
			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}
