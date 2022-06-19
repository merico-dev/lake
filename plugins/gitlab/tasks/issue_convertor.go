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
	"strconv"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"

	"github.com/apache/incubator-devlake/models/domainlayer"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/models/domainlayer/ticket"
	gitlabModels "github.com/apache/incubator-devlake/plugins/gitlab/models"
)

var ConvertIssuesMeta = core.SubTaskMeta{
	Name:             "convertIssues",
	EntryPoint:       ConvertIssues,
	EnabledByDefault: true,
	Description:      "Convert tool layer table gitlab_issues into  domain layer table issues",
}

func ConvertIssues(taskCtx core.SubTaskContext) error {
	db := taskCtx.GetDal()
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_ISSUE_TABLE)
	projectId := data.Options.ProjectId

	clauses := []dal.Clause{
		dal.Select("issues.*"),
		dal.From("_tool_gitlab_issues issues"),
		dal.Where("project_id = ? and connection_id = ?", projectId, data.Options.ConnectionId),
	}
	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	}
	defer cursor.Close()

	issueIdGen := didgen.NewDomainIdGenerator(&gitlabModels.GitlabIssue{})
	boardIdGen := didgen.NewDomainIdGenerator(&gitlabModels.GitlabProject{})

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		InputRowType:       reflect.TypeOf(gitlabModels.GitlabIssue{}),
		Input:              cursor,
		Convert: func(inputRow interface{}) ([]interface{}, error) {
			issue := inputRow.(*gitlabModels.GitlabIssue)
			domainIssue := &ticket.Issue{
				DomainEntity:            domainlayer.DomainEntity{Id: issueIdGen.Generate(data.Options.ConnectionId, issue.GitlabId)},
				IssueKey:                strconv.Itoa(issue.Number),
				Title:                   issue.Title,
				Description:             issue.Body,
				Priority:                issue.Priority,
				Type:                    issue.Type,
				AssigneeName:            issue.AssigneeName,
				LeadTimeMinutes:         issue.LeadTimeMinutes,
				Url:                     issue.Url,
				CreatedDate:             &issue.GitlabCreatedAt,
				UpdatedDate:             &issue.GitlabUpdatedAt,
				ResolutionDate:          issue.ClosedAt,
				Severity:                issue.Severity,
				Component:               issue.Component,
				OriginalStatus:          issue.Status,
				OriginalEstimateMinutes: issue.TimeEstimate,
				TimeSpentMinutes:        issue.TotalTimeSpent,
			}
			if issue.State == "opened" {
				domainIssue.Status = ticket.TODO
			} else {
				domainIssue.Status = ticket.DONE
			}
			boardIssue := &ticket.BoardIssue{
				BoardId: boardIdGen.Generate(data.Options.ConnectionId, projectId),
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
