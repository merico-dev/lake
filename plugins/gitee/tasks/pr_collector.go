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

	"github.com/apache/incubator-devlake/plugins/gitee/models"

	"github.com/apache/incubator-devlake/plugins/helper"

	"github.com/apache/incubator-devlake/plugins/core"
)

const RAW_PULL_REQUEST_TABLE = "gitee_api_pull_requests"

var CollectApiPullRequestsMeta = core.SubTaskMeta{
	Name:             "collectApiPullRequests",
	EntryPoint:       CollectApiPullRequests,
	EnabledByDefault: true,
	Description:      "Collect PullRequests data from Gitee api",
}

func CollectApiPullRequests(taskCtx core.SubTaskContext) error {
	db := taskCtx.GetDb()
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_PULL_REQUEST_TABLE)
	since := data.Since
	incremental := false
	if since == nil {
		var latestUpdated models.GiteePullRequest
		err := db.Model(&latestUpdated).
			Where("repo_id = ?", data.Repo.GiteeId).
			Order("gitee_updated_at DESC").Limit(1).Find(&latestUpdated).Error
		if err != nil {
			return fmt.Errorf("failed to get latest gitee issue record: %w", err)
		}
		if latestUpdated.GiteeId > 0 {
			since = &latestUpdated.GiteeUpdatedAt
			incremental = true
		}
	}

	collector, err := helper.NewApiCollector(helper.ApiCollectorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		ApiClient:          data.ApiClient,
		PageSize:           100,
		Incremental:        incremental,

		UrlTemplate: "repos/{{ .Params.Owner }}/{{ .Params.Repo }}/pulls",

		Query: func(reqData *helper.RequestData) (url.Values, error) {
			query := url.Values{}
			query.Set("state", "all")
			if since != nil {
				query.Set("since", since.String())
			}
			query.Set("page", fmt.Sprintf("%v", reqData.Pager.Page))
			query.Set("direction", "asc")
			query.Set("per_page", fmt.Sprintf("%v", reqData.Pager.Size))

			return query, nil
		},

		GetTotalPages: GetTotalPagesFromResponse,
		ResponseParser: func(res *http.Response) ([]json.RawMessage, error) {
			var items []json.RawMessage
			err := helper.UnmarshalResponse(res, &items)
			if err != nil {
				return nil, err
			}
			return items, nil
		},
	})

	if err != nil {
		return err
	}

	return collector.Execute()
}
