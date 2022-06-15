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
	"github.com/apache/incubator-devlake/plugins/helper"
)

const RAW_COMMIT_TABLE = "gitlab_api_commit"

var CollectApiCommitsMeta = core.SubTaskMeta{
	Name:             "collectApiCommits",
	EntryPoint:       CollectApiCommits,
	EnabledByDefault: true,
	Description:      "Collect commit data from gitlab api",
}

type GitlabApiCommit struct {
	GitlabId       string `json:"id"`
	Title          string
	Message        string
	ProjectId      int
	ShortId        string             `json:"short_id"`
	AuthorName     string             `json:"author_name"`
	AuthorEmail    string             `json:"author_email"`
	AuthoredDate   helper.Iso8601Time `json:"authored_date"`
	CommitterName  string             `json:"committer_name"`
	CommitterEmail string             `json:"committer_email"`
	CommittedDate  helper.Iso8601Time `json:"committed_date"`
	WebUrl         string             `json:"web_url"`
	Stats          struct {
		Additions int
		Deletions int
		Total     int
	}
}

func CollectApiCommits(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_COMMIT_TABLE)

	collector, err := helper.NewApiCollector(helper.ApiCollectorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		ApiClient:          data.ApiClient,
		PageSize:           100,
		Incremental:        false,
		UrlTemplate:        "projects/{{ .Params.ProjectId }}/repository/commits",
		Query:              GetQuery,
		ResponseParser:     GetRawMessageFromResponse,
	})

	if err != nil {
		return err
	}

	return collector.Execute()
}
