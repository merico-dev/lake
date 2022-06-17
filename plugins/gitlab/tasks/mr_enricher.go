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
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"github.com/apache/incubator-devlake/plugins/helper"
	"reflect"
	"time"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitlab/models"
)

var EnrichMergeRequestsMeta = core.SubTaskMeta{
	Name:             "enrichMrs",
	EntryPoint:       EnrichMergeRequests,
	EnabledByDefault: true,
	Description:      "Enrich merge requests data from GitlabCommit, GitlabMergeRequestNote and GitlabMergeRequest",
}

func EnrichMergeRequests(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_MERGE_REQUEST_TABLE)

	db := taskCtx.GetDal()
	clauses := []dal.Clause{
		dal.From(&models.GitlabMergeRequest{}),
		dal.Where("project_id=? and connection_id = ?", data.Options.ProjectId, data.Options.ConnectionId),
	}

	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	} // get mrs from theDB
	defer cursor.Close()

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		InputRowType:       reflect.TypeOf(models.GitlabMergeRequest{}),
		Input:              cursor,

		Convert: func(inputRow interface{}) ([]interface{}, error) {
			gitlabMr := inputRow.(*models.GitlabMergeRequest)
			// enrich first_comment_time field
			notes := make([]models.GitlabMergeRequestNote, 0)
			// `system` = 0 is needed since we only care about human comments
			noteClauses := []dal.Clause{
				dal.From(&models.GitlabMergeRequestNote{}),
				dal.Where("merge_request_id = ? AND is_system = ? AND connection_id = ? ",
					gitlabMr.GitlabId, false, data.Options.ConnectionId),
				dal.Orderby("gitlab_created_at asc"),
			}
			err = db.All(&notes, noteClauses...)
			if err != nil {
				return nil, err
			}

			commits := make([]models.GitlabCommit, 0)
			commitClauses := []dal.Clause{
				dal.From(&models.GitlabCommit{}),
				dal.Join(`join _tool_gitlab_merge_request_commits gmrc 
					on gmrc.commit_sha = _tool_gitlab_commits.sha`),
				dal.Where("merge_request_id = ? AND gmrc.connection_id = ?",
					gitlabMr.GitlabId, data.Options.ConnectionId),
				dal.Orderby("authored_date asc"),
			}
			err = db.All(&commits, commitClauses...)
			if err != nil {
				return nil, err
			}

			// calculate reviewRounds from commits and notes
			reviewRounds := getReviewRounds(commits, notes)
			gitlabMr.ReviewRounds = reviewRounds

			if len(notes) > 0 {
				earliestNote, err := findEarliestNote(notes)
				if err != nil {
					return nil, err
				}
				if earliestNote != nil {
					gitlabMr.FirstCommentTime = &earliestNote.GitlabCreatedAt
				}
			}
			return []interface{}{
				gitlabMr,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}

func findEarliestNote(notes []models.GitlabMergeRequestNote) (*models.GitlabMergeRequestNote, error) {
	var earliestNote *models.GitlabMergeRequestNote
	earliestTime := time.Now()
	for i := range notes {
		if !notes[i].Resolvable {
			continue
		}
		noteTime := notes[i].GitlabCreatedAt
		if noteTime.Before(earliestTime) {
			earliestTime = noteTime
			earliestNote = &notes[i]
		}
	}
	return earliestNote, nil
}

func getReviewRounds(commits []models.GitlabCommit, notes []models.GitlabMergeRequestNote) int {
	i := 0
	j := 0
	reviewRounds := 0
	if len(commits) == 0 && len(notes) == 0 {
		return 1
	}
	// state is used to keep track of previous activity
	// 0: init, 1: commit, 2: comment
	// whenever state is switched to comment, we increment reviewRounds by 1
	state := 0 // 0, 1, 2
	for i < len(commits) && j < len(notes) {
		if commits[i].AuthoredDate.Before(notes[j].GitlabCreatedAt) {
			i++
			state = 1
		} else {
			j++
			if state != 2 {
				reviewRounds++
			}
			state = 2
		}
	}
	// There's another implicit round of review in 2 scenarios
	// One: the last state is commit (state == 1)
	// Two: the last state is comment but there're still commits left
	if state == 1 || i < len(commits) {
		reviewRounds++
	}
	return reviewRounds
}
