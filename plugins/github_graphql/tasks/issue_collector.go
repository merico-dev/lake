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
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/github/models"
	githubTasks "github.com/apache/incubator-devlake/plugins/github/tasks"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/merico-dev/graphql"
	"strings"
	"time"
)

const RAW_ISSUES_TABLE = "github_graphql_issues"

type GraphqlQueryIssueWrapper struct {
	RateLimit struct {
		Cost int
	}
	Repository struct {
		IssueList struct {
			TotalCount graphql.Int
			Issues     []GraphqlQueryIssue `graphql:"nodes"`
			PageInfo   *helper.GraphqlQueryPageInfo
		} `graphql:"issues(first: $pageSize, after: $skipCursor)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type GraphqlQueryIssue struct {
	DatabaseId   int
	Number       int
	State        string
	StateReason  string
	Title        string
	Body         string
	Author       *GraphqlInlineAccountQuery
	Url          string
	ClosedAt     *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	AssigneeList struct {
		// FIXME now domain layer just support one assignee
		Assignees []GraphqlInlineAccountQuery `graphql:"nodes"`
	} `graphql:"assignees(first: 1)"`
	Milestone *struct {
		Number int `json:"number"`
	} `json:"milestone"`
	Labels struct {
		Nodes []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		}
	} `graphql:"labels(first: 100)"`
}

var CollectIssueMeta = core.SubTaskMeta{
	Name:             "CollectIssue",
	EntryPoint:       CollectIssue,
	EnabledByDefault: true,
	Description:      "Collect Issue data from GithubGraphql api",
}

var _ core.SubTaskEntryPoint = CollectIssue

func CollectIssue(taskCtx core.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*githubTasks.GithubTaskData)
	config := data.Options.TransformationRules
	issueRegexes, err := githubTasks.NewIssueRegexes(config)
	if err != nil {
		return nil
	}

	collector, err := helper.NewGraphqlCollector(helper.GraphqlCollectorArgs{
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			/*
				This struct will be JSONEncoded and stored into database along with raw data itself, to identity minimal
				set of data to be process, for example, we process JiraIssues by Board
			*/
			Params: githubTasks.GithubApiParams{
				ConnectionId: data.Options.ConnectionId,
				Owner:        data.Options.Owner,
				Repo:         data.Options.Repo,
			},
			Table: RAW_ISSUES_TABLE,
		},
		GraphqlClient: data.GraphqlClient,
		PageSize:      100,
		/*
			(Optional) Return query string for request, or you can plug them into UrlTemplate directly
		*/
		BuildQuery: func(reqData *helper.GraphqlRequestData) (interface{}, map[string]interface{}, error) {
			query := &GraphqlQueryIssueWrapper{}
			variables := map[string]interface{}{
				"pageSize":   graphql.Int(reqData.Pager.Size),
				"skipCursor": (*graphql.String)(reqData.Pager.SkipCursor),
				"owner":      graphql.String(data.Options.Owner),
				"name":       graphql.String(data.Options.Repo),
			}
			return query, variables, nil
		},
		GetPageInfo: func(iQuery interface{}, args *helper.GraphqlCollectorArgs) (*helper.GraphqlQueryPageInfo, error) {
			query := iQuery.(*GraphqlQueryIssueWrapper)
			return query.Repository.IssueList.PageInfo, nil
		},
		ResponseParser: func(iQuery interface{}, variables map[string]interface{}) ([]interface{}, error) {
			query := iQuery.(*GraphqlQueryIssueWrapper)
			issues := query.Repository.IssueList.Issues

			results := make([]interface{}, 0, 1)
			for _, issue := range issues {
				githubIssue, err := convertGithubIssue(issue, data.Options.ConnectionId, data.Repo.GithubId)
				if err != nil {
					return nil, err
				}
				githubLabels, err := convertGithubLabels(issueRegexes, issue, githubIssue)
				if err != nil {
					return nil, err
				}
				results = append(results, githubLabels...)
				results = append(results, githubIssue)
				if issue.AssigneeList.Assignees != nil && len(issue.AssigneeList.Assignees) > 0 {
					relatedUser, err := convertGraphqlPreAccount(issue.AssigneeList.Assignees[0], data.Repo.GithubId, data.Options.ConnectionId)
					if err != nil {
						return nil, err
					}
					results = append(results, relatedUser)
				}
				if issue.Author != nil {
					relatedUser, err := convertGraphqlPreAccount(*issue.Author, data.Repo.GithubId, data.Options.ConnectionId)
					if err != nil {
						return nil, err
					}
					results = append(results, relatedUser)
				}
			}
			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return collector.Execute()
}

func convertGithubIssue(issue GraphqlQueryIssue, connectionId uint64, repositoryId int) (*models.GithubIssue, errors.Error) {
	githubIssue := &models.GithubIssue{
		ConnectionId:    connectionId,
		GithubId:        issue.DatabaseId,
		RepoId:          repositoryId,
		Number:          issue.Number,
		State:           issue.State,
		Title:           issue.Title,
		Body:            strings.ReplaceAll(issue.Body, "\x00", `<0x00>`),
		Url:             issue.Url,
		ClosedAt:        issue.ClosedAt,
		GithubCreatedAt: issue.CreatedAt,
		GithubUpdatedAt: issue.UpdatedAt,
	}
	if issue.AssigneeList.Assignees != nil && len(issue.AssigneeList.Assignees) > 0 {
		githubIssue.AssigneeId = issue.AssigneeList.Assignees[0].Id
		githubIssue.AssigneeName = issue.AssigneeList.Assignees[0].Login
	}
	if issue.Author != nil {
		githubIssue.AuthorId = issue.Author.Id
		githubIssue.AuthorName = issue.Author.Login
	}
	if issue.ClosedAt != nil {
		githubIssue.LeadTimeMinutes = uint(issue.ClosedAt.Sub(issue.CreatedAt).Minutes())
	}
	if issue.Milestone != nil {
		githubIssue.MilestoneId = issue.Milestone.Number
	}
	return githubIssue, nil
}

func convertGithubLabels(issueRegexes *githubTasks.IssueRegexes, issue GraphqlQueryIssue, githubIssue *models.GithubIssue) ([]interface{}, errors.Error) {
	var results []interface{}
	for _, label := range issue.Labels.Nodes {
		results = append(results, &models.GithubIssueLabel{
			ConnectionId: githubIssue.ConnectionId,
			IssueId:      githubIssue.GithubId,
			LabelName:    label.Name,
		})
		if issueRegexes.SeverityRegex != nil {
			groups := issueRegexes.SeverityRegex.FindStringSubmatch(label.Name)
			if len(groups) > 0 {
				githubIssue.Severity = groups[1]
			}
		}
		if issueRegexes.ComponentRegex != nil {
			groups := issueRegexes.ComponentRegex.FindStringSubmatch(label.Name)
			if len(groups) > 0 {
				githubIssue.Component = groups[1]
			}
		}
		if issueRegexes.PriorityRegex != nil {
			groups := issueRegexes.PriorityRegex.FindStringSubmatch(label.Name)
			if len(groups) > 0 {
				githubIssue.Priority = groups[1]
			}
		}
		if issueRegexes.TypeBugRegex != nil {
			if ok := issueRegexes.TypeBugRegex.MatchString(label.Name); ok {
				githubIssue.Type = ticket.BUG
			}
		}
		if issueRegexes.TypeRequirementRegex != nil {
			if ok := issueRegexes.TypeRequirementRegex.MatchString(label.Name); ok {
				githubIssue.Type = ticket.REQUIREMENT
			}
		}
		if issueRegexes.TypeIncidentRegex != nil {
			if ok := issueRegexes.TypeIncidentRegex.MatchString(label.Name); ok {
				githubIssue.Type = ticket.INCIDENT
			}
		}
	}
	return results, nil
}
