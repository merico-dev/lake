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
	"reflect"

	"github.com/apache/incubator-devlake/models/domainlayer"
	"github.com/apache/incubator-devlake/models/domainlayer/code"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitlab/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ConvertApiNotesMeta = core.SubTaskMeta{
	Name:             "convertApiNotes",
	EntryPoint:       ConvertApiNotes,
	EnabledByDefault: true,
	Description:      "Update domain layer Note according to GitlabMergeRequestNote",
}

func ConvertApiNotes(taskCtx core.SubTaskContext) error {

	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_PROJECT_TABLE)
	db := taskCtx.GetDal()
	clauses := []dal.Clause{
		dal.From(&models.GitlabMergeRequestNote{}),
		dal.Join(`left join _tool_gitlab_merge_requests 
			on _tool_gitlab_merge_requests.gitlab_id = 
			_tool_gitlab_merge_request_notes.merge_request_id`),
		dal.Where(`_tool_gitlab_merge_requests.project_id = ? 
			and _tool_gitlab_merge_request_notes.connection_id = ? `,
			data.Options.ProjectId, data.Options.ConnectionId),
	}

	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	}
	defer cursor.Close()

	domainIdGeneratorNote := didgen.NewDomainIdGenerator(&models.GitlabMergeRequestNote{})
	prIdGen := didgen.NewDomainIdGenerator(&models.GitlabMergeRequest{})
	userIdGen := didgen.NewDomainIdGenerator(&models.GitlabUser{})

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		InputRowType:       reflect.TypeOf(models.GitlabMergeRequestNote{}),
		Input:              cursor,

		Convert: func(inputRow interface{}) ([]interface{}, error) {
			gitlabNotes := inputRow.(*models.GitlabMergeRequestNote)
			domainNote := &code.Note{
				DomainEntity: domainlayer.DomainEntity{
					Id: domainIdGeneratorNote.Generate(data.Options.ConnectionId, gitlabNotes.GitlabId),
				},
				PrId:        prIdGen.Generate(data.Options.ConnectionId, gitlabNotes.MergeRequestId),
				Type:        gitlabNotes.NoteableType,
				Author:      userIdGen.Generate(data.Options.ConnectionId, gitlabNotes.AuthorUsername),
				Body:        gitlabNotes.Body,
				Resolvable:  gitlabNotes.Resolvable,
				IsSystem:    gitlabNotes.IsSystem,
				CreatedDate: gitlabNotes.GitlabCreatedAt,
			}
			return []interface{}{
				domainNote,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
