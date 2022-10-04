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
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/singer"
	"github.com/apache/incubator-devlake/plugins/github_singer/models/generated"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var _ core.SubTaskEntryPoint = ExtractPr

func ExtractPr(taskCtx core.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*GithubSingerTaskData)
	extractor := helper.NewSingerApiExtractor(
		&helper.SingerExtractorArgs[generated.PullRequests]{
			Ctx:          taskCtx,
			SingerConfig: data.Config,
			ConnectionId: data.Options.ConnectionId,
			Extract: func(resData *generated.PullRequests) ([]interface{}, errors.Error) {
				//extractedModels := make([]interface{}, 0)
				//println(resData.Data)
				//println(resData.Input)
				// TODO decode some db models from api result
				// extractedModels = append(extractedModels, &models.XXXXXX)
				//return extractedModels, nil
				return nil, nil
			},
			TapType:              "github_pull_request",
			TapClass:             "TAP_GITHUB",
			StreamPropertiesFile: "github.json",
			TapSchemaSetter: func(stream *singer.Stream) bool {
				ret := true
				if stream.Stream == "pull_requests" {
					for _, meta := range stream.Metadata {
						innerMeta := meta["metadata"].(map[string]any)
						innerMeta["selected"] = true
					}
				} else {
					ret = false
				}
				return ret
			},
		},
	)
	return extractor.Execute()
}

var ExtractPrMeta = core.SubTaskMeta{
	Name:             "ExtractPr",
	EntryPoint:       ExtractPr,
	EnabledByDefault: true,
	Description:      "Extract raw data into tool layer table github_singer_pr",
}
