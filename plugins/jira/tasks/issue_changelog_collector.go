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
	"fmt"
	"net/http"
	"net/url"
	"reflect"

	"github.com/apache/incubator-devlake/errors"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/plugins/jira/tasks/apiv2models"
)

var _ core.SubTaskEntryPoint = CollectIssueChangelogs

const RAW_CHANGELOG_TABLE = "jira_api_issue_changelogs"

var CollectIssueChangelogsMeta = core.SubTaskMeta{
	Name:             "collectIssueChangelogs",
	EntryPoint:       CollectIssueChangelogs,
	EnabledByDefault: true,
	Description:      "collect Jira Issue change logs",
	DomainTypes:      []string{core.DOMAIN_TYPE_TICKET, core.DOMAIN_TYPE_CROSS},
}

func CollectIssueChangelogs(taskCtx core.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*JiraTaskData)
	// if data.JiraServerInfo.DeploymentType == models.DeploymentServer {
	// 	return nil
	// }
	log := taskCtx.GetLogger()
	db := taskCtx.GetDal()

	// query for issue_ids that needed changelog collection
	clauses := []dal.Clause{
		dal.Select("i.issue_id, i.updated AS update_time"),
		dal.From("_tool_jira_board_issues bi"),
		dal.Join("LEFT JOIN _tool_jira_issues i ON (bi.connection_id = i.connection_id AND bi.issue_id = i.issue_id)"),
		dal.Join("LEFT JOIN _tool_jira_issue_changelogs c ON (c.connection_id = i.connection_id AND c.issue_id = i.issue_id)"),
		dal.Where("i.updated > i.created AND bi.connection_id = ?  AND bi.board_id = ? AND i.std_type != ? ", data.Options.ConnectionId, data.Options.BoardId, "Epic"),
		dal.Groupby("i.issue_id, i.updated"),
		dal.Having("i.updated > max(c.issue_updated) OR  (max(c.issue_updated) IS NULL AND COUNT(c.changelog_id) > 0)"),
	}
	// apply time range if any
	since := data.Since
	if since != nil {
		clauses = append(clauses, dal.Where("i.updated > ?", *since))
	}

	if log.IsLevelEnabled(core.LOG_DEBUG) {
		count, err := db.Count(clauses...)
		if err != nil {
			return err
		}
		log.Debug("total number of issues to collect for: %d", count)
	}

	// construct the input iterator
	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	}
	// smaller struct can reduce memory footprint, we should try to avoid using big struct
	iterator, err := helper.NewDalCursorIterator(db, cursor, reflect.TypeOf(apiv2models.Input{}))
	if err != nil {
		return err
	}

	// now, let ApiCollector takes care the rest
	collector, err := helper.NewApiCollector(helper.ApiCollectorArgs{
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: JiraApiParams{
				ConnectionId: data.Options.ConnectionId,
				BoardId:      data.Options.BoardId,
			},
			Table: RAW_CHANGELOG_TABLE,
		},
		ApiClient:     data.ApiClient,
		PageSize:      100,
		Incremental:   since == nil,
		GetTotalPages: GetTotalPagesFromResponse,
		Input:         iterator,
		//UrlTemplate:   "api/3/issue/{{ .Input.IssueId }}/changelog",
		UrlTemplate: "api/2/issue/{{ .Input.IssueId }}?changelog",
		Query: func(reqData *helper.RequestData) (url.Values, errors.Error) {
			query := url.Values{}
			query.Set("startAt", fmt.Sprintf("%v", reqData.Pager.Skip))
			query.Set("maxResults", fmt.Sprintf("%v", reqData.Pager.Size))
			return query, nil
		},
		Concurrency: 10,
		ResponseParser: func(res *http.Response) ([]json.RawMessage, errors.Error) {
			var data struct {
				Values []json.RawMessage
			}
			err := helper.UnmarshalResponse(res, &data)
			if err != nil {
				return nil, err
			}
			return data.Values, nil
		},
		AfterResponse: ignoreHTTPStatus404,
	})

	if err != nil {
		return err
	}

	return collector.Execute()
}
