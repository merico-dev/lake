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

package domaininfo

import (
	"github.com/apache/incubator-devlake/core/models/domainlayer/code"
	"github.com/apache/incubator-devlake/core/models/domainlayer/crossdomain"
	"github.com/apache/incubator-devlake/core/models/domainlayer/devops"
	"github.com/apache/incubator-devlake/core/models/domainlayer/ticket"
)

type Tabler interface {
	TableName() string
}

func GetDomainTablesInfo() []Tabler {
	return []Tabler{
		// code
		&code.Commit{},
		&code.CommitFile{},
		&code.CommitFileComponent{},
		&code.CommitParent{},
		&code.Component{},
		&code.PullRequest{},
		&code.PullRequestComment{},
		&code.PullRequestCommit{},
		&code.PullRequestLabel{},
		&code.Ref{},
		&code.CommitsDiff{},
		&code.RefCommit{},
		&code.FinishedCommitsDiff{},
		&code.RefsPrCherrypick{},
		&code.Repo{},
		&code.RepoCommit{},
		//&code.RepoLanguage{},
		// crossdomain
		&crossdomain.Account{},
		&crossdomain.BoardRepo{},
		&crossdomain.IssueCommit{},
		&crossdomain.IssueRepoCommit{},
		&crossdomain.ProjectMapping{},
		&crossdomain.PullRequestIssue{},
		&crossdomain.RefsIssuesDiffs{},
		&crossdomain.Team{},
		&crossdomain.TeamUser{},
		&crossdomain.User{},
		&crossdomain.UserAccount{},
		// devops
		&devops.CICDPipeline{},
		&devops.CICDTask{},
		// didgen no table
		// ticket
		&ticket.Board{},
		&ticket.BoardIssue{},
		&ticket.BoardSprint{},
		&ticket.Issue{},
		&ticket.IssueChangelogs{},
		&ticket.IssueComment{},
		&ticket.IssueLabel{},
		&ticket.IssueWorklog{},
		&ticket.Sprint{},
		&ticket.SprintIssue{},
	}
}
