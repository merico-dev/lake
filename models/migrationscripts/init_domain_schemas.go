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

	"github.com/apache/incubator-devlake/models/migrationscripts/archived"
	"gorm.io/gorm"
)

type initDomainSchemas struct{}

func (*initDomainSchemas) Up(ctx context.Context, db *gorm.DB) error {
	err := db.Migrator().DropTable(
		"issue_assignee_history",
		"issue_status_history",
		"issue_sprints_history",
		"users",
		&archived.Repo{},
		&archived.Commit{},
		&archived.CommitParent{},
		&archived.PullRequest{},
		&archived.PullRequestCommit{},
		&archived.PullRequestComment{},
		&archived.PullRequestLabel{},
		&archived.Note{},
		&archived.RepoCommit{},
		&archived.Ref{},
		&archived.RefsCommitsDiff{},
		&archived.CommitFile{},
		&archived.Board{},
		&archived.Issue{},
		&archived.IssueLabel{},
		&archived.IssueComment{},
		&archived.BoardIssue{},
		&archived.BoardSprint{},
		&archived.IssueChangelogs{},
		&archived.Sprint{},
		&archived.SprintIssue{},
		&archived.Job{},
		&archived.Build{},
		&archived.IssueWorklog{},
		&archived.BoardRepo{},
		&archived.PullRequestIssue{},
		&archived.IssueCommit{},
		&archived.IssueRepoCommit{},
		&archived.RefsIssuesDiffs{},
		&archived.RefsPrCherrypick{},
	)
	if err != nil {
		return err
	}

	return db.Migrator().AutoMigrate(
		&archived.Repo{},
		&archived.Commit{},
		&archived.CommitParent{},
		&archived.PullRequest{},
		&archived.PullRequestCommit{},
		&archived.PullRequestComment{},
		&archived.PullRequestLabel{},
		&archived.Note{},
		&archived.RepoCommit{},
		&archived.Ref{},
		&archived.RefsCommitsDiff{},
		&archived.CommitFile{},
		&archived.Board{},
		&archived.Issue{},
		&archived.IssueLabel{},
		&archived.IssueComment{},
		&archived.BoardIssue{},
		&archived.BoardSprint{},
		&archived.IssueChangelogs{},
		&archived.Sprint{},
		&archived.SprintIssue{},
		&archived.Job{},
		&archived.Build{},
		&archived.IssueWorklog{},
		&archived.BoardRepo{},
		&archived.PullRequestIssue{},
		&archived.IssueCommit{},
		&archived.IssueRepoCommit{},
		&archived.RefsIssuesDiffs{},
		&archived.RefsPrCherrypick{},
		&archived.Account{},
		&archived.User{},
		&archived.Team{},
		&archived.UserAccount{},
		&archived.TeamUser{},
	)
}

<<<<<<< HEAD
<<<<<<< HEAD:models/migrationscripts/init_domain_schemas.go
func (*initDomainSchemas) Version() uint64 {
	return 20220707232344
=======
func (*initSchemas) Version() uint64 {
	return 20220711232344
>>>>>>> fc9539cf (feat(github): collect comments/reviews):models/migrationscripts/init_schema.go
=======
func (*initDomainSchemas) Version() uint64 {
	return 20220707232344
>>>>>>> b8097ca4 (feat(github): collect comments/reviews)
}

func (*initDomainSchemas) Owner() string {
	return "Framework"
}

func (*initDomainSchemas) Name() string {
	return "create init schemas"
}
