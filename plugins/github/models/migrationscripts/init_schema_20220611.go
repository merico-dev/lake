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
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"

	commonArchived "github.com/apache/incubator-devlake/models/migrationscripts/archived"
	"github.com/apache/incubator-devlake/plugins/github/models/migrationscripts/archived"
	"gorm.io/gorm"
)

type GithubConnection20220609 struct {
	commonArchived.Model
	Name      string `gorm:"type:varchar(100);uniqueIndex" json:"name" validate:"required"`
	Endpoint  string `mapstructure:"endpoint" env:"GITHUB_ENDPOINT" validate:"required"`
	Proxy     string `mapstructure:"proxy" env:"GITHUB_PROXY"`
	RateLimit int    `comment:"api request rate limit per hour"`
	Token     string `mapstructure:"token" env:"GITHUB_AUTH" validate:"required" encrypt:"yes"`
}

func (GithubConnection20220609) TableName() string {
	return "_tool_github_connections"
}

type InitSchemas struct {
	config core.ConfigGetter
}

func (u *InitSchemas) SetConfigGetter(config core.ConfigGetter) {
	u.config = config
}

func (u *InitSchemas) Up(ctx context.Context, db *gorm.DB) error {
	if db.Migrator().HasTable(GithubConnection20220609{}) {
		err := db.Migrator().DropTable(GithubConnection20220609{})
		if err != nil {
			return err
		}
	}
	err := db.Migrator().DropTable(
		&archived.GithubRepo{},
		&archived.GithubCommit{},
		&archived.GithubRepoCommit{},
		&archived.GithubPullRequest{},
		&archived.GithubReviewer{},
		&archived.GithubPullRequestComment{},
		&archived.GithubPullRequestCommit{},
		&archived.GithubPullRequestLabel{},
		&archived.GithubIssue{},
		&archived.GithubIssueComment{},
		&archived.GithubIssueEvent{},
		&archived.GithubIssueLabel{},
		&archived.GithubUser{},
		&archived.GithubPullRequestIssue{},
		&archived.GithubCommitStat{},
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
	)

	// create connection
	if err != nil {
		return err
	}
	err = db.Migrator().CreateTable(GithubConnection20220609{})
	if err != nil {
		return err
	}
	connection := &GithubConnection20220609{}
	connection.Endpoint = u.config.GetString(`GITHUB_ENDPOINT`)
	connection.Proxy = u.config.GetString(`GITHUB_PROXY`)
	connection.Token = u.config.GetString(`GITHUB_AUTH`)
	connection.Name = `GitHub`
	if err != nil {
		return err
	}
	err = helper.UpdateEncryptFields(connection, func(plaintext string) (string, error) {
		return core.Encrypt(u.config.GetString(core.EncodeKeyEnvStr), plaintext)
	})
	if err != nil {
		return err
	}
	// update from .env and save to db
	if connection.Endpoint != `` && connection.Token != `` {
		db.Create(connection)
	}

	// create other table with connection id
	return db.Migrator().AutoMigrate(
		&archived.GithubRepo{},
		&archived.GithubCommit{},
		&archived.GithubRepoCommit{},
		&archived.GithubPullRequest{},
		&archived.GithubReviewer{},
		&archived.GithubPullRequestComment{},
		&archived.GithubPullRequestCommit{},
		&archived.GithubPullRequestLabel{},
		&archived.GithubIssue{},
		&archived.GithubIssueComment{},
		&archived.GithubIssueEvent{},
		&archived.GithubIssueLabel{},
		&archived.GithubUser{},
		&archived.GithubPullRequestIssue{},
		&archived.GithubCommitStat{},
	)
}

func (*InitSchemas) Version() uint64 {
	return 20220613000001
}

func (*InitSchemas) Name() string {
	return "Github init schemas 20220611"
}
