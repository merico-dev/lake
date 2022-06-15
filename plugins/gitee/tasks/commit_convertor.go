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

package tasks

import (
	"reflect"

	"github.com/apache/incubator-devlake/models/domainlayer/code"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	giteeModels "github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ConvertCommitsMeta = core.SubTaskMeta{
	Name:             "convertApiCommits",
	EntryPoint:       ConvertCommits,
	EnabledByDefault: true,
	Description:      "Convert tool layer table gitee_commits into  domain layer table commits",
}

func ConvertCommits(taskCtx core.SubTaskContext) error {

	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_COMMIT_TABLE)
	db := taskCtx.GetDb()
	repoId := data.Repo.GiteeId

	// select all commits belongs to the project
	cursor, err := db.Table("_tool_gitee_commits gc").
		Joins(`left join _tool_gitee_repo_commits gpc on (
			gpc.commit_sha = gc.sha
		)`).
		Select("gc.*").
		Where("gpc.repo_id = ?", repoId).
		Rows()
	if err != nil {
		return err
	}
	defer cursor.Close()

	// TODO: adopt batch indate operation
	userDidGen := didgen.NewDomainIdGenerator(&models.GiteeUser{})
	repoDidGen := didgen.NewDomainIdGenerator(&giteeModels.GiteeRepo{})
	domainRepoId := repoDidGen.Generate(repoId)

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		InputRowType:       reflect.TypeOf(models.GiteeCommit{}),
		Input:              cursor,

		Convert: func(inputRow interface{}) ([]interface{}, error) {
			giteeCommit := inputRow.(*models.GiteeCommit)

			// convert commit
			commit := &code.Commit{}
			commit.Sha = giteeCommit.Sha
			commit.Message = giteeCommit.Message
			commit.Additions = giteeCommit.Additions
			commit.Deletions = giteeCommit.Deletions
			commit.AuthorId = userDidGen.Generate(giteeCommit.AuthorId)
			commit.AuthorName = giteeCommit.AuthorName
			commit.AuthorEmail = giteeCommit.AuthorEmail
			commit.AuthoredDate = giteeCommit.AuthoredDate
			commit.CommitterName = giteeCommit.CommitterName
			commit.CommitterEmail = giteeCommit.CommitterEmail
			commit.CommittedDate = giteeCommit.CommittedDate
			commit.CommitterId = userDidGen.Generate(giteeCommit.CommitterId)

			// convert repo / commits relationship
			repoCommit := &code.RepoCommit{
				RepoId:    domainRepoId,
				CommitSha: giteeCommit.Sha,
			}

			return []interface{}{
				commit,
				repoCommit,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
