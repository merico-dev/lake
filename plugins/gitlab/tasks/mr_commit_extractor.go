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

var ExtractApiMergeRequestsCommitsMeta = core.SubTaskMeta{
	Name:             "extractApiMergeRequestsCommits",
	EntryPoint:       ExtractApiMergeRequestsCommits,
	EnabledByDefault: true,
	Description:      "Extract raw merge requests commit data into tool layer table GitlabMergeRequestCommit and GitlabCommit",
}

func ExtractApiMergeRequestsCommits(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_MERGE_REQUEST_COMMITS_TABLE)

	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Extract: func(row *helper.RawData) ([]interface{}, error) {
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
			// get input info
			input := &GitlabInput{}
			err = json.Unmarshal(row.Input, input)
			if err != nil {
				return nil, err
			}

			gitlabMrCommit := &models.GitlabMergeRequestCommit{
				CommitSha:      gitlabApiCommit.GitlabId,
				MergeRequestId: input.GitlabId,
				ConnectionId:   data.Options.ConnectionId,
			}
			gitlabProjectCommit := &models.GitlabProjectCommit{
				ConnectionId:    data.Options.ConnectionId,
				GitlabProjectId: data.Options.ProjectId,
				CommitSha:       gitlabCommit.Sha,
			}

			// need to extract 2 kinds of entities here
			results := make([]interface{}, 0, 3)

			results = append(results, gitlabCommit, gitlabProjectCommit, gitlabMrCommit)
			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}
