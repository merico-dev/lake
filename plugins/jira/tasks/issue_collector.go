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
	goerror "errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/apache/incubator-devlake/errors"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/plugins/jira/models"
	"gorm.io/gorm"
)

const RAW_ISSUE_TABLE = "jira_api_issues"

// this struct should be moved to `jira_api_common.go`
type JiraApiParams struct {
	ConnectionId uint64
	BoardId      uint64
}

var _ core.SubTaskEntryPoint = CollectIssues

var CollectIssuesMeta = core.SubTaskMeta{
	Name:             "collectIssues",
	EntryPoint:       CollectIssues,
	EnabledByDefault: true,
	Description:      "collect Jira issues",
	DomainTypes:      []string{core.DOMAIN_TYPE_TICKET, core.DOMAIN_TYPE_CROSS},
}

func CollectIssues(taskCtx core.SubTaskContext) errors.Error {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*JiraTaskData)

	since := data.Since
	incremental := false
	// user didn't specify a time range to sync, try load from database
	if since == nil {
		var latestUpdated models.JiraIssue
		clauses := []dal.Clause{
			dal.Select("_tool_jira_issues.*"),
			dal.From("_tool_jira_issues"),
			dal.Join("LEFT JOIN _tool_jira_board_issues bi ON (bi.connection_id = _tool_jira_issues.connection_id AND bi.issue_id = _tool_jira_issues.issue_id)"),
			dal.Where("bi.connection_id = ? and bi.board_id = ?", data.Options.ConnectionId, data.Options.BoardId),
			dal.Orderby("_tool_jira_issues.updated DESC"),
		}
		err := db.First(&latestUpdated, clauses...)
		if err != nil && !goerror.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound.Wrap(err, "failed to get latest jira issue record: %w")
		}
		if latestUpdated.IssueId > 0 {
			since = &latestUpdated.Updated
			incremental = true
		}
	}
	// build jql
	// IMPORTANT: we have to keep paginated data in a consistence order to avoid data-missing, if we sort issues by
	//  `updated`, issue will be jumping between pages if it got updated during the collection process
	jql := "ORDER BY created ASC"
	if since != nil {
		// prepend a time range criteria if `since` was specified, either by user or from database
		jql = fmt.Sprintf("updated >= '%v' %v", since.Format("2006/01/02 15:04"), jql)
	}

	collector, err := helper.NewApiCollector(helper.ApiCollectorArgs{
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			/*
				This struct will be JSONEncoded and stored into database along with raw data itself, to identity minimal
				set of data to be process, for example, we process JiraIssues by Board
			*/
			Params: JiraApiParams{
				ConnectionId: data.Options.ConnectionId,
				BoardId:      data.Options.BoardId,
			},
			/*
				Table store raw data
			*/
			Table: RAW_ISSUE_TABLE,
		},
		ApiClient:   data.ApiClient,
		PageSize:    100,
		Incremental: incremental,
		/*
			url may use arbitrary variables from different connection in any order, we need GoTemplate to allow more
			flexible for all kinds of possibility.
			Pager contains information for a particular page, calculated by ApiCollector, and will be passed into
			GoTemplate to generate a url for that page.
			We want to do page-fetching in ApiCollector, because the logic are highly similar, by doing so, we can
			avoid duplicate logic for every tasks, and when we have a better idea like improving performance, we can
			do it in one place
		*/
		UrlTemplate: "agile/1.0/board/{{ .Params.BoardId }}/issue",
		/*
			(Optional) Return query string for request, or you can plug them into UrlTemplate directly
		*/
		Query: func(reqData *helper.RequestData) (url.Values, errors.Error) {
			query := url.Values{}
			query.Set("jql", jql)
			query.Set("startAt", fmt.Sprintf("%v", reqData.Pager.Skip))
			query.Set("maxResults", fmt.Sprintf("%v", reqData.Pager.Size))
			//query.Set("expand", "changelog")
			return query, nil
		},
		/*
			Some api might do pagination by http headers
		*/
		//Header: func(pager *core.Pager) http.Header {
		//},
		/*
			Sometimes, we need to collect data based on previous collected data, like jira changelog, it requires
			issue_id as part of the url.
			We can mimic `stdin` design, to accept a `Input` function which produces a `Iterator`, collector
			should iterate all records, and do data-fetching for each on, either in parallel or sequential order
			UrlTemplate: "api/3/issue/{{ Input.ID }}/changelog"
		*/
		//Input: databaseIssuesIterator,
		/*
			For api endpoint that returns number of total pages, ApiCollector can collect pages in parallel with ease,
			or other techniques are required if this information was missing.
		*/
		GetTotalPages: GetTotalPagesFromResponse,
		Concurrency:   10,
		ResponseParser: func(res *http.Response) ([]json.RawMessage, errors.Error) {
			var data struct {
				Issues []json.RawMessage `json:"issues"`
			}
			blob, err := io.ReadAll(res.Body)
			if err != nil {
				return nil, errors.Convert(err)
			}
			err = json.Unmarshal(blob, &data)
			if err != nil {
				return nil, errors.Convert(err)
			}
			return data.Issues, nil
		},
	})

	if err != nil {
		return err
	}

	return collector.Execute()
}
