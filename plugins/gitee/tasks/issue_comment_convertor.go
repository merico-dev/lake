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

	"github.com/apache/incubator-devlake/models/domainlayer"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ConvertIssueCommentsMeta = core.SubTaskMeta{
	Name:             "convertIssueComments",
	EntryPoint:       ConvertIssueComments,
	EnabledByDefault: true,
	Description:      "ConvertIssueComments data from Gitee api",
}

func ConvertIssueComments(taskCtx core.SubTaskContext) error {
	db := taskCtx.GetDb()
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_COMMENTS_TABLE)
	repoId := data.Repo.GiteeId

	cursor, err := db.Model(&models.GiteeIssueComment{}).
		Joins("left join _tool_gitee_issues "+
			"on _tool_gitee_issues.gitee_id = _tool_gitee_issue_comments.issue_id").
		Where("repo_id = ?", repoId).Rows()
	if err != nil {
		return err
	}
	defer cursor.Close()

	issueIdGen := didgen.NewDomainIdGenerator(&models.GiteeIssue{})
	userIdGen := didgen.NewDomainIdGenerator(&models.GiteeUser{})

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		InputRowType:       reflect.TypeOf(models.GiteeIssueComment{}),
		Input:              cursor,
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Convert: func(inputRow interface{}) ([]interface{}, error) {
			giteeIssueComment := inputRow.(*models.GiteeIssueComment)
			domainIssueComment := &ticket.IssueComment{
				DomainEntity: domainlayer.DomainEntity{
					Id: issueIdGen.Generate(giteeIssueComment.GiteeId),
				},
				IssueId:     issueIdGen.Generate(giteeIssueComment.IssueId),
				Body:        giteeIssueComment.Body,
				UserId:      userIdGen.Generate(giteeIssueComment.AuthorUserId),
				CreatedDate: giteeIssueComment.GiteeCreatedAt,
			}
			return []interface{}{
				domainIssueComment,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
