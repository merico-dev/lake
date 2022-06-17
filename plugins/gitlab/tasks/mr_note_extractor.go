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
	"github.com/apache/incubator-devlake/plugins/gitlab/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

type MergeRequestNote struct {
	GitlabId        int    `json:"id"`
	MergeRequestId  int    `json:"noteable_id"`
	MergeRequestIid int    `json:"noteable_iid"`
	NoteableType    string `json:"noteable_type"`
	Body            string
	GitlabCreatedAt helper.Iso8601Time `json:"created_at"`
	Confidential    bool
	Resolvable      bool `json:"resolvable"`
	System          bool `json:"system"`
	Author          struct {
		Username string `json:"username"`
	}
}

var ExtractApiMergeRequestsNotesMeta = core.SubTaskMeta{
	Name:             "extractApiMergeRequestsNotes",
	EntryPoint:       ExtractApiMergeRequestsNotes,
	EnabledByDefault: true,
	Description:      "Extract raw merge requests notes data into tool layer table GitlabMergeRequestNote",
}

func ExtractApiMergeRequestsNotes(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_MERGE_REQUEST_NOTES_TABLE)

	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Extract: func(row *helper.RawData) ([]interface{}, error) {
			mrNote := &MergeRequestNote{}
			err := json.Unmarshal(row.Data, mrNote)
			if err != nil {
				return nil, err
			}

			toolMrNote, err := convertMergeRequestNote(mrNote)
			toolMrNote.ConnectionId = data.Options.ConnectionId
			if err != nil {
				return nil, err
			}
			results := make([]interface{}, 0, 2)
			if !toolMrNote.IsSystem {
				toolMrComment := &models.GitlabMergeRequestComment{
					GitlabId:        toolMrNote.GitlabId,
					MergeRequestId:  toolMrNote.MergeRequestId,
					MergeRequestIid: toolMrNote.MergeRequestIid,
					Body:            toolMrNote.Body,
					AuthorUsername:  toolMrNote.AuthorUsername,
					GitlabCreatedAt: toolMrNote.GitlabCreatedAt,
					Resolvable:      toolMrNote.Resolvable,
					ConnectionId:    data.Options.ConnectionId,
				}
				results = append(results, toolMrComment)

			}
			results = append(results, toolMrNote)

			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}

func convertMergeRequestNote(mrNote *MergeRequestNote) (*models.GitlabMergeRequestNote, error) {
	gitlabMergeRequestNote := &models.GitlabMergeRequestNote{
		GitlabId:        mrNote.GitlabId,
		MergeRequestId:  mrNote.MergeRequestId,
		MergeRequestIid: mrNote.MergeRequestIid,
		NoteableType:    mrNote.NoteableType,
		AuthorUsername:  mrNote.Author.Username,
		Body:            mrNote.Body,
		GitlabCreatedAt: mrNote.GitlabCreatedAt.ToTime(),
		Confidential:    mrNote.Confidential,
		Resolvable:      mrNote.Resolvable,
		IsSystem:        mrNote.System,
	}
	return gitlabMergeRequestNote, nil
}
