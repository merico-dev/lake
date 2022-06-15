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
	"reflect"

	"github.com/apache/incubator-devlake/models/domainlayer/code"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ConvertPullRequestCommitsMeta = core.SubTaskMeta{
	Name:             "convertPullRequestCommits",
	EntryPoint:       ConvertPullRequestCommits,
	EnabledByDefault: true,
	Description:      "Convert tool layer table gitee_pull_request_commits into  domain layer table pull_request_commits",
}

func ConvertPullRequestCommits(taskCtx core.SubTaskContext) (err error) {
	db := taskCtx.GetDb()
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_PULL_REQUEST_COMMIT_TABLE)
	repoId := data.Repo.GiteeId

	pullIdGen := didgen.NewDomainIdGenerator(&models.GiteePullRequest{})

	cursor, err := db.Model(&models.GiteePullRequestCommit{}).
		Joins(`left join _tool_gitee_pull_requests on _tool_gitee_pull_requests.gitee_id = _tool_gitee_pull_request_commits.pull_request_id`).
		Where("_tool_gitee_pull_requests.repo_id = ?", repoId).
		Order("pull_request_id ASC").Rows()
	if err != nil {
		return err
	}
	defer cursor.Close()

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		InputRowType:       reflect.TypeOf(models.GiteePullRequestCommit{}),
		Input:              cursor,
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Convert: func(inputRow interface{}) ([]interface{}, error) {
			giteePullRequestCommit := inputRow.(*models.GiteePullRequestCommit)
			domainPrCommit := &code.PullRequestCommit{
				CommitSha:     giteePullRequestCommit.CommitSha,
				PullRequestId: pullIdGen.Generate(giteePullRequestCommit.PullRequestId),
			}
			return []interface{}{
				domainPrCommit,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
