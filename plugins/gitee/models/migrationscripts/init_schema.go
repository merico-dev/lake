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
	"context"
	"fmt"

	"github.com/apache/incubator-devlake/config"
	"gorm.io/gorm/clause"

	"github.com/apache/incubator-devlake/plugins/core"

	"github.com/apache/incubator-devlake/plugins/gitee/models/migrationscripts/archived"
	"gorm.io/gorm"
)

type InitSchemas struct{}

func (*InitSchemas) Up(ctx context.Context, db *gorm.DB) error {
	rawTableList := []string{
		"_raw_gitee_api_commit",
		"_raw_gitee_api_issues",
		"_raw_gitee_api_pull_requests",
		"_raw_gitee_api_pull_request_commits",
		"_raw_gitee_api_pull_request_reviews",
		"_raw_gitee_api_repo",
		"_raw_gitee_api_comments",
		"_raw_gitee_api_commits",
		"_raw_gitee_issue_comments",
	}
	for _, v := range rawTableList {
		err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", v)).Error
		if err != nil {
			return err
		}
	}

	err := db.Migrator().DropTable(
		&archived.GiteeRepo{},
		&archived.GiteeCommit{},
		&archived.GiteeRepoCommit{},
		&archived.GiteePullRequest{},
		&archived.GiteePullRequestLabel{},
		&archived.GiteeUser{},
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

	err = db.Migrator().AutoMigrate(
		&archived.GiteeRepo{},
		&archived.GiteeCommit{},
		&archived.GiteeRepoCommit{},
		&archived.GiteePullRequest{},
		&archived.GiteePullRequestLabel{},
		&archived.GiteeUser{},
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
	v := config.GetConfig()
	encKey := v.GetString(core.EncodeKeyEnvStr)

	conn.Name = "init gitee connection"
	conn.ID = 1
	conn.Endpoint = v.GetString("GITEE_ENDPOINT")
	conn.Token, err = core.Encrypt(encKey, v.GetString("GITEE_AUTH"))
	if err != nil {
		return err
	}
	conn.Proxy = v.GetString("GITEE_PROXY")
	conn.RateLimit = v.GetInt("GITEE_API_REQUESTS_PER_HOUR")

	err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(conn).Error

	if err != nil {
		return err
	}
	return nil
}

func (*InitSchemas) Version() uint64 {
	return 20220617231243
}

func (*InitSchemas) Name() string {
	return "Gitee init schemas"
}
