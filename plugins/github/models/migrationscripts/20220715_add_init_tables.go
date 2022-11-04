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
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/helpers/migrationhelper"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"

	"github.com/apache/incubator-devlake/plugins/github/models/migrationscripts/archived"
)

type addInitTables struct {
	config core.ConfigGetter
}

func (u *addInitTables) SetConfigGetter(config core.ConfigGetter) {
	u.config = config
}

func (u *addInitTables) Up(basicRes core.BasicRes) errors.Error {
	db := basicRes.GetDal()
	err := db.DropTables(
		&archived.GithubRepo{},
		&archived.GithubConnection{},
		&archived.GithubCommit{},
		&archived.GithubRepoCommit{},
		&archived.GithubPullRequest{},
		&archived.GithubReviewer{},
		&archived.GithubPrCommit{},
		&archived.GithubPrLabel{},
		&archived.GithubIssue{},
		&archived.GithubIssueComment{},
		&archived.GithubIssueEvent{},
		&archived.GithubIssueLabel{},
		&archived.GithubPrIssue{},
		&archived.GithubCommitStat{},
		&archived.GithubPrComment{},
		&archived.GithubPrReview{},
		&archived.GithubRepoAccount{},
		&archived.GithubAccountOrg{},
		&archived.GithubAccount{},
		"_tool_github_users",
		"_tool_github_milestones",
		"_raw_github_api_issues",
		"_raw_github_api_comments",
		"_raw_github_api_commits",
		"_raw_github_api_commit_stats",
		"_raw_github_api_events",
		"_raw_github_api_issues",
		"_raw_github_api_pull_requests",
		"_raw_github_api_pull_request_commits",
		"_raw_github_api_pull_request_reviews",
		"_raw_github_api_repositories",
		"_raw_github_api_reviews",
	)

	// create connection
	if err != nil {
		return err
	}

	err = db.AutoMigrate(archived.GithubConnection{})
	if err != nil {
		return err
	}
	encodeKey := basicRes.GetConfig(core.EncodeKeyEnvStr)
	connection := &archived.GithubConnection{}
	connection.Endpoint = basicRes.GetConfig(`GITHUB_ENDPOINT`)
	connection.Proxy = basicRes.GetConfig(`GITHUB_PROXY`)
	connection.Token = basicRes.GetConfig(`GITHUB_AUTH`)
	connection.Name = `GitHub`
	if connection.Endpoint != `` && connection.Token != `` && encodeKey != `` {
		err = helper.UpdateEncryptFields(connection, func(plaintext string) (string, errors.Error) {
			return core.Encrypt(encodeKey, plaintext)
		})
		if err != nil {
			return err
		}
		// update from .env and save to db
		err = db.Create(connection)
		if err != nil {
			return err
		}
	}

	// create other table with connection id
	err = migrationhelper.AutoMigrateTables(
		basicRes,
		&archived.GithubRepo{},
		&archived.GithubCommit{},
		&archived.GithubRepoCommit{},
		&archived.GithubPullRequest{},
		&archived.GithubReviewer{},
		&archived.GithubPrComment{},
		&archived.GithubPrCommit{},
		&archived.GithubPrLabel{},
		&archived.GithubIssue{},
		&archived.GithubIssueComment{},
		&archived.GithubIssueEvent{},
		&archived.GithubIssueLabel{},
		&archived.GithubAccount{},
		&archived.GithubPrIssue{},
		&archived.GithubCommitStat{},
		&archived.GithubMilestone{},
		&archived.GithubPrReview{},
		&archived.GithubRepoAccount{},
		&archived.GithubAccountOrg{},
		&archived.GithubAccount{},
	)
	return err
}

func (*addInitTables) Version() uint64 {
	return 20220715000001
}

func (*addInitTables) Name() string {
	return "Github init schemas 20220707"
}
