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

	"github.com/apache/incubator-devlake/models/domainlayer/code"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitlab/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ConvertApiMergeRequestsCommitsMeta = core.SubTaskMeta{
	Name:             "convertApiMergeRequestsCommits",
	EntryPoint:       ConvertApiMergeRequestsCommits,
	EnabledByDefault: true,
	Description:      "Update domain layer PullRequestCommit according to GitlabMergeRequestCommit",
}

func ConvertApiMergeRequestsCommits(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_MERGE_REQUEST_COMMITS_TABLE)
	db := taskCtx.GetDal()

	clauses := []dal.Clause{
		dal.From(&models.GitlabMergeRequestCommit{}),
		dal.Join(`left join _tool_gitlab_merge_requests 
			on _tool_gitlab_merge_requests.gitlab_id = 
			_tool_gitlab_merge_request_commits.merge_request_id`),
		dal.Where(`_tool_gitlab_merge_requests.project_id = ? 
			and _tool_gitlab_merge_requests.connection_id = ?`,
			data.Options.ProjectId, data.Options.ConnectionId),
		dal.Orderby("merge_request_id ASC"),
	}

	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	}

	// TODO: adopt batch indate operation
	domainIdGenerator := didgen.NewDomainIdGenerator(&models.GitlabMergeRequest{})

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		InputRowType:       reflect.TypeOf(models.GitlabMergeRequestCommit{}),
		Input:              cursor,

		Convert: func(inputRow interface{}) ([]interface{}, error) {
			gitlabMergeRequestCommit := inputRow.(*models.GitlabMergeRequestCommit)
			domainPrcommit := &code.PullRequestCommit{
				CommitSha:     gitlabMergeRequestCommit.CommitSha,
				PullRequestId: domainIdGenerator.Generate(data.Options.ConnectionId, gitlabMergeRequestCommit.MergeRequestId),
			}
			return []interface{}{
				domainPrcommit,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
