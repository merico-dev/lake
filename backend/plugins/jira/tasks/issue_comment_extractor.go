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
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/jira/models"
	"github.com/apache/incubator-devlake/plugins/jira/tasks/apiv2models"
)

var _ plugin.SubTaskEntryPoint = ExtractIssueComments

var ExtractIssueCommentsMeta = plugin.SubTaskMeta{
	Name:             "extractIssueComments",
	EntryPoint:       ExtractIssueComments,
	EnabledByDefault: true,
	Description:      "extract Jira Issue comments",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET, plugin.DOMAIN_TYPE_CROSS},
}

func ExtractIssueComments(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*JiraTaskData)
	if data.JiraServerInfo.DeploymentType == models.DeploymentServer {
		return nil
	}
	connectionId := data.Options.ConnectionId
	extractor, err := api.NewApiExtractor(api.ApiExtractorArgs{
		RawDataSubTaskArgs: api.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: JiraApiParams{
				ConnectionId: data.Options.ConnectionId,
				BoardId:      data.Options.BoardId,
			},
			Table: RAW_ISSUE_COMMENT_TABLE,
		},
		Extract: func(row *api.RawData) ([]interface{}, errors.Error) {
			// process input
			var input apiv2models.Input
			err := errors.Convert(json.Unmarshal(row.Input, &input))
			if err != nil {
				return nil, err
			}
			var comment apiv2models.Comment
			err = errors.Convert(json.Unmarshal(row.Data, &comment))
			if err != nil {
				return nil, err
			}
			// prepare output
			var result []interface{}
			c := comment.ToToolLayer(connectionId, input.IssueId, &input.UpdateTime)
			// this is crucial for incremental update
			c.IssueUpdated = &input.UpdateTime
			// collect comment
			result = append(result, c)
			return result, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}
