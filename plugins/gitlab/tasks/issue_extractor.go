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

	"github.com/apache/incubator-devlake/plugins/gitlab/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ExtractApiIssuesMeta = core.SubTaskMeta{
	Name:             "extractApiIssues",
	EntryPoint:       ExtractApiIssues,
	EnabledByDefault: true,
	Description:      "Extract raw Issues data into tool layer table gitlab_issues",
}

type IssuesResponse struct {
	ProjectId int `json:"project_id"`
	Milestone struct {
		Due_date    string
		Project_id  int
		State       string
		Description string
		Iid         int
		Id          int
		Title       string
		CreatedAt   helper.Iso8601Time
		UpdatedAt   helper.Iso8601Time
	}
	Author struct {
		State     string
		WebUrl    string
		AvatarUrl string
		Username  string
		Id        int
		Name      string
	}
	Description string
	State       string
	Iid         int
	Assignees   []struct {
		AvatarUrl string
		WebUrl    string
		State     string
		Username  string
		Id        int
		Name      string
	}
	Assignee *struct {
		AvatarUrl string
		WebUrl    string
		State     string
		Username  string
		Id        int
		Name      string
	}
	Type               string
	Labels             []string `json:"labels"`
	UpVotes            int
	DownVotes          int
	MergeRequestsCount int
	Id                 int `json:"id"`
	Title              string
	GitlabUpdatedAt    helper.Iso8601Time  `json:"updated_at"`
	GitlabCreatedAt    helper.Iso8601Time  `json:"created_at"`
	GitlabClosedAt     *helper.Iso8601Time `json:"closed_at"`
	ClosedBy           struct {
		State     string
		WebUrl    string
		AvatarUrl string
		Username  string
		Id        int
		Name      string
	}
	UserNotesCount int
	DueDate        helper.Iso8601Time
	WebUrl         string `json:"web_url"`
	References     struct {
		Short    string
		Relative string
		Full     string
	}
	TimeStats struct {
		TimeEstimate        int64
		TotalTimeSpent      int64
		HumanTimeEstimate   string
		HumanTotalTimeSpent string
	}
	HasTasks         bool
	TaskStatus       string
	Confidential     bool
	DiscussionLocked bool
	IssueType        string
	Serverity        string
	Links            struct {
		Self       string `json:"url"`
		Notes      string
		AwardEmoji string
		Project    string
	}
	TaskCompletionStatus struct {
		Count          int
		CompletedCount int
	}
}

func ExtractApiIssues(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_ISSUE_TABLE)

	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Extract: func(row *helper.RawData) ([]interface{}, error) {
			body := &IssuesResponse{}
			err := json.Unmarshal(row.Data, body)
			if err != nil {
				return nil, err
			}
			// need to extract 2 kinds of entities here
			if body.ProjectId == 0 {
				return nil, nil
			}
			//If this is not Issue, ignore
			if body.IssueType != "ISSUE" && body.Type != "ISSUE" {
				return nil, nil
			}
			results := make([]interface{}, 0, 2)
			gitlabIssue, err := convertGitlabIssue(body, data.Options.ProjectId)
			if err != nil {
				return nil, err
			}

			for _, label := range body.Labels {
				results = append(results, &models.GitlabIssueLabel{
					IssueId:      gitlabIssue.GitlabId,
					LabelName:    label,
					ConnectionId: data.Options.ConnectionId,
				})

			}
			gitlabIssue.ConnectionId = data.Options.ConnectionId
			results = append(results, gitlabIssue)

			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}
func convertGitlabIssue(issue *IssuesResponse, projectId int) (*models.GitlabIssue, error) {
	gitlabIssue := &models.GitlabIssue{
		GitlabId:        issue.Id,
		ProjectId:       projectId,
		Number:          issue.Iid,
		State:           issue.State,
		Title:           issue.Title,
		Body:            issue.Description,
		Url:             issue.Links.Self,
		ClosedAt:        helper.Iso8601TimeToTime(issue.GitlabClosedAt),
		GitlabCreatedAt: issue.GitlabCreatedAt.ToTime(),
		GitlabUpdatedAt: issue.GitlabUpdatedAt.ToTime(),
		TimeEstimate:    issue.TimeStats.TimeEstimate,
		TotalTimeSpent:  issue.TimeStats.TotalTimeSpent,
	}

	if issue.Assignee != nil {
		gitlabIssue.AssigneeId = issue.Assignee.Id
		gitlabIssue.AssigneeName = issue.Assignee.Username
	}
	if issue.GitlabClosedAt != nil {
		gitlabIssue.LeadTimeMinutes = uint(issue.GitlabClosedAt.ToTime().Sub(issue.GitlabCreatedAt.ToTime()).Minutes())
	}

	return gitlabIssue, nil
}
