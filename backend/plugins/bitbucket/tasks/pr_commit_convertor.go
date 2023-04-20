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

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/domainlayer/code"
	"github.com/apache/incubator-devlake/core/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"

	bitbucketModels "github.com/apache/incubator-devlake/plugins/bitbucket/models"
)

var ConvertPrCommitsMeta = plugin.SubTaskMeta{
	Name:             "convertPullRequestCommits",
	EntryPoint:       ConvertPullRequestCommits,
	EnabledByDefault: true,
	Description:      "Convert tool layer table bitbucket_pull_request_commits into  domain layer table pull_request_commits",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_CODE_REVIEW},
}

func ConvertPullRequestCommits(taskCtx plugin.SubTaskContext) (err errors.Error) {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_PULL_REQUEST_COMMITS_TABLE)
	db := taskCtx.GetDal()

	pullIdGen := didgen.NewDomainIdGenerator(&bitbucketModels.BitbucketPullRequest{})

	cursor, err := db.Cursor(
		dal.From(&bitbucketModels.BitbucketPrCommit{}),
		dal.Where("connection_id = ? AND repo_id = ?", data.Options.ConnectionId, data.Options.FullName),
	)
	if err != nil {
		return err
	}
	defer cursor.Close()

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		InputRowType:       reflect.TypeOf(bitbucketModels.BitbucketPrCommit{}),
		Input:              cursor,
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Convert: func(inputRow interface{}) ([]interface{}, errors.Error) {
			prCommit := inputRow.(*bitbucketModels.BitbucketPrCommit)
			domainPrCommit := &code.PullRequestCommit{
				CommitSha:           prCommit.CommitSha,
				PullRequestId:       pullIdGen.Generate(prCommit.ConnectionId, prCommit.RepoId, prCommit.PullRequestId),
				CommitAuthoredEmail: prCommit.CommitAuthoredEmail,
				CommitAuthoredName:  prCommit.CommitAuthoredName,
				CommitAuthoredDate:  prCommit.CommitAuthoredDate,
			}
			return []interface{}{
				domainPrCommit,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
