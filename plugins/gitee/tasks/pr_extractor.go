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
	"encoding/json"
	"regexp"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ExtractApiPullRequestsMeta = core.SubTaskMeta{
	Name:             "extractApiPullRequests",
	EntryPoint:       ExtractApiPullRequests,
	EnabledByDefault: true,
	Description:      "Extract raw PullRequests data into tool layer table gitee_pull_requests",
}

type GiteeApiPullResponse struct {
	GiteeId int `json:"id"`
	Number  int
	State   string
	Title   string
	Body    json.RawMessage
	HtmlUrl string `json:"html_url"`
	Labels  []struct {
		Name string `json:"name"`
	} `json:"labels"`
	Assignee *struct {
		Id    int
		Login string
		Name  string
	}
	User *struct {
		Id    int
		Login string
		Name  string
	}
	ClosedAt       *helper.Iso8601Time `json:"closed_at"`
	MergedAt       *helper.Iso8601Time `json:"merged_at"`
	GiteeCreatedAt helper.Iso8601Time  `json:"created_at"`
	GiteeUpdatedAt helper.Iso8601Time  `json:"updated_at"`
	MergeCommitSha string              `json:"merge_commit_sha"`
	Head           struct {
		Ref string
		Sha string
	}
	Base struct {
		Ref  string
		Sha  string
		Repo struct {
			Id      int
			Name    string
			Url     string
			HtmlUrl string
			SshUrl  string `json:"ssh_url"`
		}
	}
}

func ExtractApiPullRequests(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_PULL_REQUEST_TABLE)
	config := data.Options.Config
	var labelTypeRegex *regexp.Regexp
	var labelComponentRegex *regexp.Regexp
	var prType = config.PrType
	if prType == "" {
		prType = taskCtx.GetConfig("GITEE_PR_TYPE")
	}
	var prComponent = config.PrComponent
	if prComponent == "" {
		prComponent = taskCtx.GetConfig("GITEE_PR_COMPONENT")
	}
	if len(prType) > 0 {
		labelTypeRegex = regexp.MustCompile(prType)
	}
	if len(prComponent) > 0 {
		labelComponentRegex = regexp.MustCompile(prComponent)
	}

	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Extract: func(row *helper.RawData) ([]interface{}, error) {
			pullResponse := &GiteeApiPullResponse{}
			err := json.Unmarshal(row.Data, pullResponse)
			if err != nil {
				return nil, err
			}

			// need to extract 2 kinds of entities here
			results := make([]interface{}, 0, 1)
			if pullResponse.GiteeId == 0 {
				return nil, nil
			}
			giteePr, err := convertGiteePullRequest(pullResponse, data.Repo.GiteeId)
			if err != nil {
				return nil, err
			}
			for _, label := range pullResponse.Labels {
				results = append(results, &models.GiteePullRequestLabel{
					PullId:    giteePr.GiteeId,
					LabelName: label.Name,
				})
				// if pr.Type has not been set and prType is set in .env, process the below
				if labelTypeRegex != nil {
					groups := labelTypeRegex.FindStringSubmatch(label.Name)
					if len(groups) > 0 {
						giteePr.Type = groups[1]
					}
				}

				// if pr.Component has not been set and prComponent is set in .env, process
				if labelComponentRegex != nil {
					groups := labelComponentRegex.FindStringSubmatch(label.Name)
					if len(groups) > 0 {
						giteePr.Component = groups[1]
					}
				}
			}
			results = append(results, giteePr)

			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}
func convertGiteePullRequest(pull *GiteeApiPullResponse, repoId int) (*models.GiteePullRequest, error) {
	giteePull := &models.GiteePullRequest{
		GiteeId:        pull.GiteeId,
		RepoId:         repoId,
		Number:         pull.Number,
		State:          pull.State,
		Title:          pull.Title,
		Url:            pull.HtmlUrl,
		AuthorName:     pull.User.Login,
		AuthorId:       pull.User.Id,
		GiteeCreatedAt: pull.GiteeCreatedAt.ToTime(),
		GiteeUpdatedAt: pull.GiteeUpdatedAt.ToTime(),
		ClosedAt:       helper.Iso8601TimeToTime(pull.ClosedAt),
		MergedAt:       helper.Iso8601TimeToTime(pull.MergedAt),
		MergeCommitSha: pull.MergeCommitSha,
		Body:           string(pull.Body),
		BaseRef:        pull.Base.Ref,
		BaseCommitSha:  pull.Base.Sha,
		HeadRef:        pull.Head.Ref,
		HeadCommitSha:  pull.Head.Sha,
	}
	return giteePull, nil
}
