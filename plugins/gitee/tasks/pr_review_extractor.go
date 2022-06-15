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
	"strings"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ExtractApiPullRequestReviewsMeta = core.SubTaskMeta{
	Name:             "extractApiPullRequestReviews",
	EntryPoint:       ExtractApiPullRequestReviews,
	EnabledByDefault: true,
	Description:      "Extract raw PullRequestReviews data into tool layer table gitee_reviewers",
}

type PullRequestReview struct {
	GiteeId int `json:"id"`
	User    struct {
		Id    int
		Login string
	}
	Body        string
	State       string
	SubmittedAt helper.Iso8601Time `json:"submitted_at"`
}

func ExtractApiPullRequestReviews(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, _ := CreateRawDataSubTaskArgs(taskCtx, RAW_PULL_REQUEST_REVIEW_TABLE)
	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Extract: func(row *helper.RawData) ([]interface{}, error) {
			apiPullRequestReview := &PullRequestReview{}
			if strings.HasPrefix(string(row.Data), "{\"message\": \"Not Found\"") {
				return nil, nil
			}
			err := json.Unmarshal(row.Data, apiPullRequestReview)
			if err != nil {
				return nil, err
			}
			pull := &SimplePr{}
			err = json.Unmarshal(row.Input, pull)
			if err != nil {
				return nil, err
			}
			// need to extract 2 kinds of entities here
			results := make([]interface{}, 0, 1)

			giteeReviewer := &models.GiteeReviewer{
				GiteeId:       apiPullRequestReview.User.Id,
				Login:         apiPullRequestReview.User.Login,
				PullRequestId: pull.GiteeId,
			}
			results = append(results, giteeReviewer)

			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}
