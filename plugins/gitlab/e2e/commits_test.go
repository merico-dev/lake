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
	"testing"

	"github.com/apache/incubator-devlake/helpers/e2ehelper"
	"github.com/apache/incubator-devlake/plugins/gitlab/impl"
	"github.com/apache/incubator-devlake/plugins/gitlab/models"
	"github.com/apache/incubator-devlake/plugins/gitlab/tasks"
)

func TestGitlabCommitDataFlow(t *testing.T) {

	var gitlab impl.Gitlab
	dataflowTester := e2ehelper.NewDataFlowTester(t, "gitlab", gitlab)

	taskData := &tasks.GitlabTaskData{
		Options: &tasks.GitlabOptions{
			ConnectionId: 1,
			ProjectId:    20171709,
		},
	}
	// import raw data table
	dataflowTester.ImportCsvIntoRawTable("./raw_tables/_raw_gitlab_api_commit.csv",
		"_raw_gitlab_api_commit")

	// verify extraction
	dataflowTester.FlushTabler(&models.GitlabCommit{})
	dataflowTester.FlushTabler(&models.GitlabProjectCommit{})
	dataflowTester.FlushTabler(&models.GitlabUser{})
	dataflowTester.Subtask(tasks.ExtractApiCommitsMeta, taskData)
	dataflowTester.VerifyTable(
		models.GitlabCommit{},
		fmt.Sprintf("./snapshot_tables/%s.csv", models.GitlabCommit{}.TableName()),
		[]string{"connection_id", "sha"},
		[]string{
			"title",
			"message",
			"short_id",
			"author_name",
			"author_email",
			"authored_date",
			"committer_name",
			"committer_email",
			"committed_date",
			"web_url",
			"additions",
			"deletions",
			"total",
			"_raw_data_params",
			"_raw_data_table",
			"_raw_data_id",
			"_raw_data_remark",
		},
	)

	dataflowTester.VerifyTable(
		models.GitlabProjectCommit{},
		fmt.Sprintf("./snapshot_tables/%s.csv", models.GitlabProjectCommit{}.TableName()),
		[]string{"connection_id", "gitlab_project_id", "commit_sha"},
		[]string{
			"_raw_data_params",
			"_raw_data_table",
			"_raw_data_id",
			"_raw_data_remark",
		},
	)

	dataflowTester.VerifyTable(
		models.GitlabUser{},
		fmt.Sprintf("./snapshot_tables/%s.csv", models.GitlabUser{}.TableName()),
		[]string{"connection_id", "email"},
		[]string{
			"name",
			"_raw_data_params",
			"_raw_data_table",
			"_raw_data_id",
			"_raw_data_remark",
		},
	)

	// verify conversion
	dataflowTester.FlushTabler(&code.Commit{})
	dataflowTester.FlushTabler(&code.RepoCommit{})
	dataflowTester.Subtask(tasks.ConvertCommitsMeta, taskData)
	dataflowTester.VerifyTable(
		code.Commit{},
		fmt.Sprintf("./snapshot_tables/%s.csv", code.Commit{}.TableName()),
		[]string{"sha"},
		[]string{
			"_raw_data_params",
			"_raw_data_table",
			"_raw_data_id",
			"_raw_data_remark",
			"additions",
			"deletions",
			"dev_eq",
			"message",
			"author_name",
			"author_email",
			"authored_date",
			"author_id",
			"committer_name",
			"committer_email",
			"committed_date",
			"committer_id",
		},
	)

	dataflowTester.VerifyTable(
		code.RepoCommit{},
		fmt.Sprintf("./snapshot_tables/%s.csv", code.RepoCommit{}.TableName()),
		[]string{"repo_id", "commit_sha"},
		[]string{
			"_raw_data_params",
			"_raw_data_table",
			"_raw_data_id",
			"_raw_data_remark",
		},
	)

}
