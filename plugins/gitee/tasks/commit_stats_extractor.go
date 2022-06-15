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

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ExtractApiCommitStatsMeta = core.SubTaskMeta{
	Name:             "extractApiCommitStats",
	EntryPoint:       ExtractApiCommitStats,
	EnabledByDefault: false,
	Description:      "Extract raw commit stats data into tool layer table gitee_commit_stats",
}

type ApiSingleCommitResponse struct {
	Sha   string
	Stats struct {
		id        string
		Additions int
		Deletions int
		total     int
	}
	Commit struct {
		Committer struct {
			Name  string
			Email string
			Date  helper.Iso8601Time
		}
	}
}

func ExtractApiCommitStats(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, _ := CreateRawDataSubTaskArgs(taskCtx, RAW_COMMIT_STATS_TABLE)

	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Extract: func(row *helper.RawData) ([]interface{}, error) {
			body := &ApiSingleCommitResponse{}
			err := json.Unmarshal(row.Data, body)
			if err != nil {
				return nil, err
			}
			if body.Sha == "" {
				return nil, nil
			}

			db := taskCtx.GetDb()
			commit := &models.GiteeCommit{}
			err = db.Model(commit).Where("sha = ?", body.Sha).Limit(1).Find(commit).Error
			if err != nil {
				return nil, err
			}

			commit.Additions = body.Stats.Additions
			commit.Deletions = body.Stats.Deletions

			commitStat := &models.GiteeCommitStat{
				Additions:     body.Stats.Additions,
				Deletions:     body.Stats.Deletions,
				CommittedDate: body.Commit.Committer.Date.ToTime(),
				Sha:           body.Sha,
			}

			results := make([]interface{}, 0, 2)

			results = append(results, commit)
			results = append(results, commitStat)

			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}
