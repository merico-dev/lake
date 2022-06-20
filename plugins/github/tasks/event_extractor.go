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

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/github/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ExtractApiEventsMeta = core.SubTaskMeta{
	Name:             "extractApiEvents",
	EntryPoint:       ExtractApiEvents,
	EnabledByDefault: true,
	Description:      "Extract raw Events data into tool layer table github_issue_events",
}

type IssueEvent struct {
	GithubId int `json:"id"`
	Event    string
	Actor    struct {
		Login string
	}
	Issue struct {
		Id int
	}
	GithubCreatedAt helper.Iso8601Time `json:"created_at"`
}

func ExtractApiEvents(taskCtx core.SubTaskContext) error {
	data := taskCtx.GetData().(*GithubTaskData)

	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: GithubApiParams{
				ConnectionId: data.Options.ConnectionId,
				Owner:        data.Options.Owner,
				Repo:         data.Options.Repo,
			},
			Table: RAW_EVENTS_TABLE,
		},
		Extract: func(row *helper.RawData) ([]interface{}, error) {
			body := &IssueEvent{}
			err := json.Unmarshal(row.Data, body)
			if err != nil {
				return nil, err
			}
			results := make([]interface{}, 0, 1)
			if body.GithubId == 0 {
				return nil, nil
			}
			githubIssueEvent, err := convertGithubEvent(body, data.Options.ConnectionId)
			if err != nil {
				return nil, err
			}
			results = append(results, githubIssueEvent)

			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}

func convertGithubEvent(event *IssueEvent, connId uint64) (*models.GithubIssueEvent, error) {
	githubEvent := &models.GithubIssueEvent{
		ConnectionId:    connId,
		GithubId:        event.GithubId,
		IssueId:         event.Issue.Id,
		Type:            event.Event,
		AuthorUsername:  event.Actor.Login,
		GithubCreatedAt: event.GithubCreatedAt.ToTime(),
	}
	return githubEvent, nil
}
