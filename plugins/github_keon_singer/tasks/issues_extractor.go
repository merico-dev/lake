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
	"github.com/apache/incubator-devlake/helpers/pluginhelper/tap"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/github_keon_singer/models/generated"
)

var _ core.SubTaskEntryPoint = ExtractIssues

func ExtractIssues(taskCtx core.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*GithubSingerTaskData)
	extractor, err := tap.NewTapExtractor(
		&tap.ExtractorArgs[generated.Issues]{
			Ctx:          taskCtx,
			TapProvider:  data.Options.TapProvider,
			ConnectionId: data.Options.ConnectionId,
			Extract: func(resData *generated.Issues) ([]interface{}, errors.Error) {
				// TODO decode some db models from api result
				return nil, nil
			},
			StreamName: "issues",
		},
	)
	if err != nil {
		return err
	}
	return extractor.Execute()
}

var ExtractIssuesMeta = core.SubTaskMeta{
	Name:             "ExtractIssues",
	EntryPoint:       ExtractIssues,
	EnabledByDefault: true,
	Description:      "Extract singer-tap Github issues",
}
