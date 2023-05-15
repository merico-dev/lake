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

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/log"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/jira/models"
	"github.com/apache/incubator-devlake/plugins/jira/tasks/apiv2models"
)

var _ plugin.SubTaskEntryPoint = CollectIssueComments

const RAW_ISSUE_COMMENT_TABLE = "jira_api_issue_comments"

var CollectIssueCommentsMeta = plugin.SubTaskMeta{
	Name:             "collectIssueComments",
	EntryPoint:       CollectIssueComments,
	EnabledByDefault: true,
	Description:      "collect Jira issue comments, supports both timeFilter and diffSync.",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET, plugin.DOMAIN_TYPE_CROSS},
}

func CollectIssueComments(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*JiraTaskData)
	if data.JiraServerInfo.DeploymentType == models.DeploymentServer {
		return nil
	}
	logger := taskCtx.GetLogger()
	db := taskCtx.GetDal()

	collectorWithState, err := api.NewStatefulApiCollector(api.RawDataSubTaskArgs{
		Ctx: taskCtx,
		Params: JiraApiParams{
			ConnectionId: data.Options.ConnectionId,
			BoardId:      data.Options.BoardId,
		},
		Table: RAW_ISSUE_COMMENT_TABLE,
	}, data.TimeAfter)
	if err != nil {
		return err
	}

	clauses := []dal.Clause{
		dal.Select("i.issue_id AS issue_id, i.updated AS update_time"),
		dal.From("_tool_jira_board_issues bi"),
		dal.Join("LEFT JOIN _tool_jira_issues i ON (bi.connection_id = i.connection_id AND bi.issue_id = i.issue_id)"),
		dal.Where("bi.connection_id=? and bi.board_id = ? AND i.std_type != ? AND i.comment_total > 100", data.Options.ConnectionId, data.Options.BoardId, "Epic"),
	}
	incremental := collectorWithState.IsIncremental()
	if incremental && collectorWithState.LatestState.LatestSuccessStart != nil {
		clauses = append(
			clauses,
			dal.Where("i.updated > ?", collectorWithState.LatestState.LatestSuccessStart),
		)
	}

	if logger.IsLevelEnabled(log.LOG_DEBUG) {
		count, err := db.Count(clauses...)
		if err != nil {
			return err
		}
		logger.Debug("total number of issues to collect for: %d", count)
	}

	// construct the input iterator
	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	}
	// smaller struct can reduce memory footprint, we should try to avoid using big struct
	iterator, err := api.NewDalCursorIterator(db, cursor, reflect.TypeOf(apiv2models.Input{}))
	if err != nil {
		return err
	}

	// now, let ApiCollector takes care the rest
	err = collectorWithState.InitCollector(api.ApiCollectorArgs{
		ApiClient:     data.ApiClient,
		PageSize:      100,
		Incremental:   incremental,
		GetTotalPages: GetTotalPagesFromResponse,
		Input:         iterator,
		UrlTemplate:   "api/3/issue/{{ .Input.IssueId }}/comment",
		Query: func(reqData *api.RequestData) (url.Values, errors.Error) {
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
			err := api.UnmarshalResponse(res, &data)
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

	return collectorWithState.Execute()
}
