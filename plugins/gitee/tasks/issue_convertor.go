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

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"

	"github.com/apache/incubator-devlake/models/domainlayer"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/models/domainlayer/ticket"
	giteeModels "github.com/apache/incubator-devlake/plugins/gitee/models"
)

var ConvertIssuesMeta = core.SubTaskMeta{
	Name:             "convertIssues",
	EntryPoint:       ConvertIssues,
	EnabledByDefault: true,
	Description:      "Convert tool layer table gitee_issues into  domain layer table issues",
}

func ConvertIssues(taskCtx core.SubTaskContext) error {
	db := taskCtx.GetDb()
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_ISSUE_TABLE)
	repoId := data.Repo.GiteeId

	issue := &giteeModels.GiteeIssue{}
	cursor, err := db.Model(issue).Where("repo_id = ?", repoId).Rows()

	if err != nil {
		return err
	}
	defer cursor.Close()

	issueIdGen := didgen.NewDomainIdGenerator(&giteeModels.GiteeIssue{})
	userIdGen := didgen.NewDomainIdGenerator(&giteeModels.GiteeUser{})
	boardIdGen := didgen.NewDomainIdGenerator(&giteeModels.GiteeRepo{})

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		InputRowType:       reflect.TypeOf(giteeModels.GiteeIssue{}),
		Input:              cursor,
		Convert: func(inputRow interface{}) ([]interface{}, error) {
			issue := inputRow.(*giteeModels.GiteeIssue)
			domainIssue := &ticket.Issue{
				DomainEntity:    domainlayer.DomainEntity{Id: issueIdGen.Generate(issue.GiteeId)},
				IssueKey:        issue.Number,
				Title:           issue.Title,
				Description:     issue.Body,
				Priority:        issue.Priority,
				Type:            issue.Type,
				AssigneeId:      userIdGen.Generate(issue.AssigneeId),
				AssigneeName:    issue.AssigneeName,
				CreatorId:       userIdGen.Generate(issue.AuthorId),
				CreatorName:     issue.AuthorName,
				LeadTimeMinutes: issue.LeadTimeMinutes,
				Url:             issue.Url,
				CreatedDate:     &issue.GiteeCreatedAt,
				UpdatedDate:     &issue.GiteeUpdatedAt,
				ResolutionDate:  issue.ClosedAt,
				Severity:        issue.Severity,
				Component:       issue.Component,
			}
			if issue.State == "closed" {
				domainIssue.Status = ticket.DONE
			} else {
				domainIssue.Status = ticket.TODO
			}
			boardIssue := &ticket.BoardIssue{
				BoardId: boardIdGen.Generate(repoId),
				IssueId: domainIssue.Id,
			}
			return []interface{}{
				domainIssue,
				boardIssue,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
