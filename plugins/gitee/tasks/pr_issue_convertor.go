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
	"strconv"

	"github.com/apache/incubator-devlake/models/domainlayer/crossdomain"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ConvertPullRequestIssuesMeta = core.SubTaskMeta{
	Name:             "convertPullRequestIssues",
	EntryPoint:       ConvertPullRequestIssues,
	EnabledByDefault: true,
	Description:      "Convert tool layer table gitee_pull_request_issues into  domain layer table pull_request_issues",
}

func ConvertPullRequestIssues(taskCtx core.SubTaskContext) error {
	db := taskCtx.GetDb()
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_PULL_REQUEST_TABLE)
	repoId := data.Repo.GiteeId

	cursor, err := db.Model(&models.GiteePullRequestIssue{}).
		Joins(`left join _tool_gitee_pull_requests on _tool_gitee_pull_requests.gitee_id = _tool_gitee_pull_request_issues.pull_request_id`).
		Where("_tool_gitee_pull_requests.repo_id = ?", repoId).
		Order("pull_request_id ASC").
		Rows()
	if err != nil {
		return err
	}
	defer cursor.Close()
	prIdGen := didgen.NewDomainIdGenerator(&models.GiteePullRequest{})
	issueIdGen := didgen.NewDomainIdGenerator(&models.GiteeIssue{})

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		InputRowType:       reflect.TypeOf(models.GiteePullRequestIssue{}),
		Input:              cursor,
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Convert: func(inputRow interface{}) ([]interface{}, error) {
			giteePrIssue := inputRow.(*models.GiteePullRequestIssue)
			issueNum, _ := strconv.Atoi(giteePrIssue.IssueNumber)
			pullRequestIssue := &crossdomain.PullRequestIssue{
				PullRequestId:     prIdGen.Generate(giteePrIssue.PullRequestId),
				IssueId:           issueIdGen.Generate(giteePrIssue.IssueId),
				IssueNumber:       issueNum,
				PullRequestNumber: giteePrIssue.PullRequestNumber,
			}
			return []interface{}{
				pullRequestIssue,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
