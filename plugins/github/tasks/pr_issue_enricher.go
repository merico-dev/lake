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
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
	githubModels "github.com/apache/incubator-devlake/plugins/github/models"
	"github.com/apache/incubator-devlake/plugins/helper"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var EnrichPullRequestIssuesMeta = core.SubTaskMeta{
	Name:             "enrichPullRequestIssues",
	EntryPoint:       EnrichPullRequestIssues,
	EnabledByDefault: true,
	Description:      "Create tool layer table github_pull_request_issues from github_pull_reqeusts",
}

func EnrichPullRequestIssues(taskCtx core.SubTaskContext) (err error) {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*GithubTaskData)
	repoId := data.Repo.GithubId

	var prBodyCloseRegex *regexp.Regexp
	prBodyClosePattern := data.Options.PrBodyClosePattern
	//the pattern before the issue number, sometimes, the issue number is #1098, sometimes it is https://xxx/#1098
	prBodyClosePattern = strings.Replace(prBodyClosePattern, "%s", data.Options.Owner, 1)
	prBodyClosePattern = strings.Replace(prBodyClosePattern, "%s", data.Options.Repo, 1)
	if len(prBodyClosePattern) > 0 {
		prBodyCloseRegex = regexp.MustCompile(prBodyClosePattern)
	}
	charPattern := regexp.MustCompile(`[a-zA-Z\s,]+`)
	cursor, err := db.Cursor(dal.From(&githubModels.GithubPullRequest{}), dal.Where("repo_id = ?", repoId))
	if err != nil {
		return err
	}
	defer cursor.Close()
	// iterate all rows

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		InputRowType: reflect.TypeOf(githubModels.GithubPullRequest{}),
		Input:        cursor,
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: GithubApiParams{
				ConnectionId: data.Options.ConnectionId,
				Owner:        data.Options.Owner,
				Repo:         data.Options.Repo,
			},
			Table: RAW_PULL_REQUEST_TABLE,
		},
		Convert: func(inputRow interface{}) ([]interface{}, error) {
			githubPullRequst := inputRow.(*githubModels.GithubPullRequest)
			results := make([]interface{}, 0, 1)

			//find the matched string in body
			issueNumberListStr := ""

			if prBodyCloseRegex != nil {
				issueNumberListStr = prBodyCloseRegex.FindString(githubPullRequst.Body)
			}

			if issueNumberListStr == "" {
				return nil, nil
			}

			issueNumberListStr = charPattern.ReplaceAllString(issueNumberListStr, "#")
			//split the string by '#'
			issueNumberList := strings.Split(issueNumberListStr, "#")

			for _, issueNumberStr := range issueNumberList {
				issue := &githubModels.GithubIssue{}
				issueNumberStr = strings.TrimSpace(issueNumberStr)
				//change the issueNumberStr to int, if cannot be changed, just continue
				issueNumber, numFormatErr := strconv.Atoi(issueNumberStr)
				if numFormatErr != nil {
					continue
				}
				err = db.All(
					issue,
					dal.Where("number = ? and repo_id = ? and connection_id = ?", issueNumber, repoId, data.Options.ConnectionId),
					dal.Limit(1),
				)
				if err != nil {
					return nil, err
				}
				if issue.Number == 0 {
					continue
				}
				githubPullRequstIssue := &githubModels.GithubPullRequestIssue{
					ConnectionId:      data.Options.ConnectionId,
					PullRequestId:     githubPullRequst.GithubId,
					IssueId:           issue.GithubId,
					PullRequestNumber: githubPullRequst.Number,
					IssueNumber:       issue.Number,
				}
				results = append(results, githubPullRequstIssue)
			}
			return results, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
