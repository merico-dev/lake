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

package e2e

import (
	"fmt"
	"github.com/apache/incubator-devlake/models/domainlayer/ticket"
	"testing"

	"github.com/apache/incubator-devlake/helpers/e2ehelper"
	"github.com/apache/incubator-devlake/plugins/gitlab/impl"
	"github.com/apache/incubator-devlake/plugins/gitlab/models"
	"github.com/apache/incubator-devlake/plugins/gitlab/tasks"
)

func TestGitlabIssueDataFlow(t *testing.T) {

	var gitlab impl.Gitlab
	dataflowTester := e2ehelper.NewDataFlowTester(t, "gitlab", gitlab)

	taskData := &tasks.GitlabTaskData{
		Options: &tasks.GitlabOptions{
			ConnectionId: 1,
			ProjectId:    11624398,
		},
	}
	// import raw data table
	dataflowTester.ImportCsvIntoRawTable("./raw_tables/_raw_gitlab_api_issues.csv",
		"_raw_gitlab_api_issues")

	// verify extraction
	dataflowTester.FlushTabler(&models.GitlabIssue{})
	dataflowTester.FlushTabler(&models.GitlabIssueLabel{})
	dataflowTester.FlushTabler(&models.GitlabUser{})
	dataflowTester.Subtask(tasks.ExtractApiIssuesMeta, taskData)
	dataflowTester.VerifyTable(
		models.GitlabIssue{},
		fmt.Sprintf("./snapshot_tables/%s.csv", models.GitlabIssue{}.TableName()),
		[]string{"connection_id", "gitlab_id"},
		[]string{
			"project_id",
			"number",
			"state",
			"title",
			"body",
			"priority",
			"type",
			"status",
			"assignee_id",
			"assignee_name",
			"lead_time_minutes",
			"url",
			"closed_at",
			"gitlab_created_at",
			"gitlab_updated_at",
			"severity",
			"component",
			"time_estimate",
			"total_time_spent",
			"_raw_data_params",
			"_raw_data_table",
			"_raw_data_id",
			"_raw_data_remark",
		},
	)

	dataflowTester.VerifyTable(
		models.GitlabIssueLabel{},
		fmt.Sprintf("./snapshot_tables/%s.csv", models.GitlabIssueLabel{}.TableName()),
		[]string{"connection_id", "issue_id", "label_name"},
		[]string{
			"_raw_data_params",
			"_raw_data_table",
			"_raw_data_id",
			"_raw_data_remark",
		},
	)

	// verify conversion
	dataflowTester.FlushTabler(&ticket.Issue{})
	dataflowTester.FlushTabler(&ticket.BoardIssue{})
	dataflowTester.Subtask(tasks.ConvertIssuesMeta, taskData)
	dataflowTester.VerifyTable(
		ticket.Issue{},
		fmt.Sprintf("./snapshot_tables/%s.csv", ticket.Issue{}.TableName()),
		[]string{"id"},
		[]string{
			"_raw_data_params",
			"_raw_data_table",
			"_raw_data_id",
			"_raw_data_remark",
			"url",
			"issue_key",
			"title",
			"description",
			"epic_key",
			"type",
			"status",
			"original_status",
			"story_point",
			"resolution_date",
			"created_date",
			"updated_date",
			"lead_time_minutes",
			"parent_issue_id",
			"priority",
			"original_estimate_minutes",
			"time_spent_minutes",
			"time_remaining_minutes",
			"creator_id",
			"assignee_id",
			"assignee_name",
			"severity",
			"component",
			"icon_url",
			"creator_name",
		},
	)

	dataflowTester.VerifyTable(
		&ticket.BoardIssue{},
		fmt.Sprintf("./snapshot_tables/%s.csv", ticket.BoardIssue{}.TableName()),
		[]string{"board_id", "issue_id"},
		[]string{
			"_raw_data_params",
			"_raw_data_table",
			"_raw_data_id",
			"_raw_data_remark",
		},
	)

}
