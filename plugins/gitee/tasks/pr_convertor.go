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

	"github.com/apache/incubator-devlake/models/domainlayer"
	"github.com/apache/incubator-devlake/models/domainlayer/code"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ConvertPullRequestsMeta = core.SubTaskMeta{
	Name:             "convertPullRequests",
	EntryPoint:       ConvertPullRequests,
	EnabledByDefault: true,
	Description:      "ConvertPullRequests data from Gitee api",
}

func ConvertPullRequests(taskCtx core.SubTaskContext) error {
	db := taskCtx.GetDb()
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_PULL_REQUEST_TABLE)
	repoId := data.Repo.GiteeId

	cursor, err := db.Model(&models.GiteePullRequest{}).Where("repo_id = ?", repoId).Rows()
	if err != nil {
		return err
	}
	defer cursor.Close()

	prIdGen := didgen.NewDomainIdGenerator(&models.GiteePullRequest{})
	repoIdGen := didgen.NewDomainIdGenerator(&models.GiteeRepo{})
	userIdGen := didgen.NewDomainIdGenerator(&models.GiteeUser{})

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		InputRowType:       reflect.TypeOf(models.GiteePullRequest{}),
		Input:              cursor,
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Convert: func(inputRow interface{}) ([]interface{}, error) {
			pr := inputRow.(*models.GiteePullRequest)
			domainPr := &code.PullRequest{
				DomainEntity: domainlayer.DomainEntity{
					Id: prIdGen.Generate(pr.GiteeId),
				},
				BaseRepoId:     repoIdGen.Generate(pr.RepoId),
				Status:         pr.State,
				Title:          pr.Title,
				Url:            pr.Url,
				AuthorId:       userIdGen.Generate(pr.AuthorId),
				AuthorName:     pr.AuthorName,
				Description:    pr.Body,
				CreatedDate:    pr.GiteeCreatedAt,
				MergedDate:     pr.MergedAt,
				ClosedDate:     pr.ClosedAt,
				PullRequestKey: pr.Number,
				Type:           pr.Type,
				Component:      pr.Component,
				MergeCommitSha: pr.MergeCommitSha,
				BaseRef:        pr.BaseRef,
				BaseCommitSha:  pr.BaseCommitSha,
				HeadRef:        pr.HeadRef,
				HeadCommitSha:  pr.HeadCommitSha,
			}
			return []interface{}{
				domainPr,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
