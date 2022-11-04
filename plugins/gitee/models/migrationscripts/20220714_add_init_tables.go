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

package migrationscripts

import (
	"strconv"

	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/helpers/migrationhelper"

	"github.com/apache/incubator-devlake/plugins/core"

	"github.com/apache/incubator-devlake/plugins/gitee/models/migrationscripts/archived"
)

type addInitTables struct{}

func (*addInitTables) Up(baseRes core.BasicRes) errors.Error {
	db := baseRes.GetDal()
	err := db.DropTables(
		&archived.GiteeRepo{},
		&archived.GiteeCommit{},
		&archived.GiteeRepoCommit{},
		&archived.GiteePullRequest{},
		&archived.GiteePullRequestLabel{},
		&archived.GiteePullRequestComment{},
		&archived.GiteeIssue{},
		&archived.GiteeIssueComment{},
		&archived.GiteeCommitStat{},
		&archived.GiteeIssueLabel{},
		&archived.GiteePullRequestCommit{},
		&archived.GiteePullRequestIssue{},
		&archived.GiteeReviewer{},
		&archived.GiteeConnection{},
		"_tool_gitee_users",
		"_raw_gitee_api_commit",
		"_raw_gitee_api_issues",
		"_raw_gitee_api_pull_requests",
		"_raw_gitee_api_pull_request_commits",
		"_raw_gitee_api_pull_request_reviews",
		"_raw_gitee_api_repo",
		"_raw_gitee_api_comments",
		"_raw_gitee_api_commits",
		"_raw_gitee_issue_comments",
	)

	if err != nil {
		return err
	}

	err = migrationhelper.AutoMigrateTables(
		baseRes,
		&archived.GiteeRepo{},
		&archived.GiteeCommit{},
		&archived.GiteeRepoCommit{},
		&archived.GiteePullRequest{},
		&archived.GiteePullRequestLabel{},
		&archived.GiteeAccount{},
		&archived.GiteePullRequestComment{},
		&archived.GiteeIssue{},
		&archived.GiteeIssueComment{},
		&archived.GiteeCommitStat{},
		&archived.GiteeIssueLabel{},
		&archived.GiteePullRequestCommit{},
		&archived.GiteePullRequestIssue{},
		&archived.GiteeReviewer{},
		&archived.GiteeConnection{},
	)

	if err != nil {
		return err
	}

	conn := &archived.GiteeConnection{}
	encKey := baseRes.GetConfig(core.EncodeKeyEnvStr)

	conn.Name = "init gitee connection"
	conn.ID = 1
	conn.Endpoint = baseRes.GetConfig("GITEE_ENDPOINT")
	conn.Token, err = core.Encrypt(encKey, baseRes.GetConfig("GITEE_AUTH"))
	if err != nil {
		return err
	}
	conn.Proxy = baseRes.GetConfig("GITEE_PROXY")

	var err1 error
	conn.RateLimitPerHour, err1 = strconv.Atoi(baseRes.GetConfig("GITEE_API_REQUESTS_PER_HOUR"))
	if err1 != nil {
		conn.RateLimitPerHour = 1000
	}

	err = db.CreateIfNotExist(conn)
	if err != nil {
		return err
	}
	return nil
}

func (*addInitTables) Version() uint64 {
	return 20220714231268
}

func (*addInitTables) Name() string {
	return "Gitee init schemas"
}
