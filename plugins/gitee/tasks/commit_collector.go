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
	"net/url"
	"strconv"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
)

const RAW_COMMIT_TABLE = "gitee_api_commit"

var CollectCommitsMeta = core.SubTaskMeta{
	Name:             "collectApiCommits",
	EntryPoint:       CollectApiCommits,
	EnabledByDefault: true,
	Description:      "Collect commit data from gitee api",
}

func CollectApiCommits(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_COMMIT_TABLE)

	collector, err := helper.NewApiCollector(helper.ApiCollectorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		ApiClient:          data.ApiClient,
		PageSize:           100,
		Incremental:        false,
		UrlTemplate:        "repos/{{ .Params.Owner }}/{{ .Params.Repo }}/commits",
		Query: func(reqData *helper.RequestData) (url.Values, error) {
			query := url.Values{}
			query.Set("access_token", data.Options.Token)
			query.Set("with_stats", "true")
			query.Set("sort", "asc")
			query.Set("page", strconv.Itoa(reqData.Pager.Page))
			query.Set("per_page", strconv.Itoa(reqData.Pager.Size))
			return query, nil
		},
		Concurrency:    20,
		ResponseParser: GetRawMessageFromResponse,
	})

	if err != nil {
		return err
	}

	return collector.Execute()
}
