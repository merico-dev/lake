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
	"github.com/apache/incubator-devlake/models/domainlayer/code"
	"github.com/apache/incubator-devlake/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/plugins/github/models"
	"github.com/apache/incubator-devlake/plugins/github/tasks"
	"testing"

	"github.com/apache/incubator-devlake/helpers/e2ehelper"
	"github.com/apache/incubator-devlake/plugins/github/impl"
)

func TestCommentDataFlow(t *testing.T) {
	var plugin impl.Github
	dataflowTester := e2ehelper.NewDataFlowTester(t, "gitlab", plugin)

	githubRepository := &models.GithubRepo{
		GithubId: 134018330,
	}
	taskData := &tasks.GithubTaskData{
		Options: &tasks.GithubOptions{
			ConnectionId: 1,
			Owner:        "panjf2000",
			Repo:         "ants",
			Config: models.Config{
				PrType:               "type/(.*)$",
				PrComponent:          "component/(.*)$",
				PrBodyClosePattern:   "(?mi)(fix|close|resolve|fixes|closes|resolves|fixed|closed|resolved)[\\s]*.*(((and )?(#|https:\\/\\/github.com\\/%s\\/%s\\/issues\\/)\\d+[ ]*)+)",
				IssueSeverity:        "severity/(.*)$",
				IssuePriority:        "^(highest|high|medium|low)$",
				IssueComponent:       "component/(.*)$",
				IssueTypeBug:         "^(bug|failure|error)$",
				IssueTypeIncident:    "",
				IssueTypeRequirement: "^(feat|feature|proposal|requirement)$",
			},
		},
		Repo: githubRepository,
	}

	// import raw data table
	dataflowTester.ImportCsvIntoRawTable("./raw_tables/_raw_github_api_comments.csv", "_raw_github_api_comments")
	dataflowTester.ImportCsvIntoTabler("./snapshot_tables/_tool_github_issues.csv", &models.GithubIssue{})
	dataflowTester.ImportCsvIntoTabler("./snapshot_tables/_tool_github_issue_labels.csv", &models.GithubIssueLabel{})
	dataflowTester.ImportCsvIntoTabler("./snapshot_tables/_tool_github_pull_requests.csv", models.GithubPullRequest{})

	// verify extraction
	dataflowTester.FlushTabler(&models.GithubIssueComment{})
	dataflowTester.FlushTabler(&models.GithubPullRequestComment{})
	dataflowTester.Subtask(tasks.ExtractApiCommentsMeta, taskData)
	dataflowTester.VerifyTable(
		models.GithubIssueComment{},
		fmt.Sprintf("./snapshot_tables/%s.csv", models.GithubIssueComment{}.TableName()),
		[]string{"connection_id", "github_id"},
		[]string{
			"issue_id",
			"body",
			"author_username",
			"author_user_id",
			"github_created_at",
			"github_updated_at",
			"_raw_data_params",
			"_raw_data_table",
			"_raw_data_id",
			"_raw_data_remark",
		},
	)
	dataflowTester.VerifyTable(
		models.GithubPullRequestComment{},
		fmt.Sprintf("./snapshot_tables/%s.csv", models.GithubPullRequestComment{}.TableName()),
		[]string{"connection_id", "github_id"},
		[]string{
			"pull_request_id",
			"body",
			"author_username",
			"author_user_id",
			"github_created_at",
			"github_updated_at",
			"_raw_data_params",
			"_raw_data_table",
			"_raw_data_id",
			"_raw_data_remark",
		},
	)

	// verify comment conversion
	dataflowTester.FlushTabler(&ticket.IssueComment{})
	dataflowTester.Subtask(tasks.ConvertIssueCommentsMeta, taskData)
	dataflowTester.VerifyTable(
		ticket.IssueComment{},
		fmt.Sprintf("./snapshot_tables/%s.csv", ticket.IssueComment{}.TableName()),
		[]string{"id"},
		[]string{
			"issue_id",
			"body",
			"user_id",
			"created_date",
		},
	)

	// verify relation in pr and comment conversion
	dataflowTester.FlushTabler(&code.PullRequestComment{})
	dataflowTester.Subtask(tasks.ConvertPullRequestCommentsMeta, taskData)
	dataflowTester.VerifyTable(
		code.PullRequestComment{},
		fmt.Sprintf("./snapshot_tables/%s.csv", code.PullRequestComment{}.TableName()),
		[]string{"id"},
		[]string{
			"pull_request_id",
			"body",
			"user_id",
			"created_date",
			"commit_sha",
			"position",
		},
	)
}
