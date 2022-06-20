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
	"github.com/apache/incubator-devlake/models/domainlayer/crossdomain"
	"github.com/apache/incubator-devlake/plugins/github/models"
	"testing"

	"github.com/apache/incubator-devlake/helpers/e2ehelper"
	"github.com/apache/incubator-devlake/plugins/github/impl"
	"github.com/apache/incubator-devlake/plugins/github/tasks"
)

func TestPrEnrichIssueDataFlow(t *testing.T) {
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
	dataflowTester.ImportCsvIntoTabler(fmt.Sprintf("./snapshot_tables/%s.csv", models.GithubIssue{}.TableName()), &models.GithubIssue{})
	dataflowTester.ImportCsvIntoTabler(fmt.Sprintf("./snapshot_tables/%s.csv", models.GithubPullRequest{}.TableName()), models.GithubPullRequest{})

	// verify extraction
	dataflowTester.FlushTabler(&models.GithubPullRequestIssue{})
	dataflowTester.Subtask(tasks.EnrichPullRequestIssuesMeta, taskData)
	dataflowTester.VerifyTable(
		models.GithubPullRequestIssue{},
		fmt.Sprintf("./snapshot_tables/%s.csv", models.GithubPullRequestIssue{}.TableName()),
		[]string{"connection_id", "pull_request_id", "issue_id"},
		[]string{
			"pull_request_number",
			"issue_number",
			"_raw_data_params",
			"_raw_data_table",
			"_raw_data_id",
			"_raw_data_remark",
		},
	)

	// verify extraction
	dataflowTester.FlushTabler(&crossdomain.PullRequestIssue{})
	dataflowTester.Subtask(tasks.ConvertPullRequestIssuesMeta, taskData)
	dataflowTester.VerifyTable(
		crossdomain.PullRequestIssue{},
		fmt.Sprintf("./snapshot_tables/%s.csv", crossdomain.PullRequestIssue{}.TableName()),
		[]string{"pull_request_id", "issue_id"},
		[]string{
			"pull_request_number",
			"issue_number",
			"_raw_data_params",
			"_raw_data_table",
			"_raw_data_id",
			"_raw_data_remark",
		},
	)
}
