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

	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/plugins/core"
	gitlabModels "github.com/apache/incubator-devlake/plugins/gitlab/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ConvertIssueLabelsMeta = core.SubTaskMeta{
	Name:             "convertIssueLabels",
	EntryPoint:       ConvertIssueLabels,
	EnabledByDefault: true,
	Description:      "Convert tool layer table gitlab_issue_labels into  domain layer table issue_labels",
}

func ConvertIssueLabels(taskCtx core.SubTaskContext) error {
	db := taskCtx.GetDal()
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_ISSUE_TABLE)
	projectId := data.Options.ProjectId
	clauses := []dal.Clause{
		dal.Select("*"),
		dal.From(&gitlabModels.GitlabIssueLabel{}),
		dal.Join("left join _tool_gitlab_issues on " +
			"_tool_gitlab_issues.gitlab_id = _tool_gitlab_issue_labels.issue_id"),
		dal.Where("_tool_gitlab_issues.project_id = ? "+
			"and _tool_gitlab_issues.connection_id = ?",
			projectId, data.Options.ConnectionId),
		dal.Orderby("issue_id ASC"),
	}

	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	}
	defer cursor.Close()

	issueIdGen := didgen.NewDomainIdGenerator(&gitlabModels.GitlabIssue{})

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		InputRowType:       reflect.TypeOf(gitlabModels.GitlabIssueLabel{}),
		Input:              cursor,
		Convert: func(inputRow interface{}) ([]interface{}, error) {
			issueLabel := inputRow.(*gitlabModels.GitlabIssueLabel)
			domainIssueLabel := &ticket.IssueLabel{
				IssueId:   issueIdGen.Generate(data.Options.ConnectionId, issueLabel.IssueId),
				LabelName: issueLabel.LabelName,
			}
			return []interface{}{
				domainIssueLabel,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
