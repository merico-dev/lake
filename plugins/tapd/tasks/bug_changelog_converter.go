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
	"reflect"
	"time"

	"github.com/apache/incubator-devlake/plugins/core/dal"

	"github.com/apache/incubator-devlake/models/common"
	"github.com/apache/incubator-devlake/models/domainlayer"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/plugins/tapd/models"
)

type BugChangelogItemResult struct {
	ConnectionId      uint64     `gorm:"primaryKey;type:BIGINT  NOT NULL"`
	WorkspaceId       uint64     `gorm:"primaryKey;type:BIGINT  NOT NULL"`
	Id                uint64     `gorm:"primaryKey;type:BIGINT  NOT NULL" json:"id"`
	BugId             uint64     `json:"bug_id"`
	Author            string     `json:"author" gorm:"type:varchar(255)"`
	Field             string     `json:"field"`
	OldValue          string     `json:"old_value"`
	NewValue          string     `json:"new_value"`
	Memo              string     `json:"memo"`
	Created           *time.Time `json:"created"`
	ChangelogId       uint64     `gorm:"primaryKey;type:BIGINT  NOT NULL"`
	ValueBeforeParsed string     `json:"value_before"`
	ValueAfterParsed  string     `json:"value_after"`
	IterationIdFrom   uint64
	IterationIdTo     uint64
	common.NoPKModel
}

func ConvertBugChangelog(taskCtx core.SubTaskContext) errors.Error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_BUG_CHANGELOG_TABLE, false)
	logger := taskCtx.GetLogger()
	db := taskCtx.GetDal()
	statusList := make([]models.TapdBugStatus, 0)
	statusLanguageMap, getStdStatus, err := getDefaltStdStatusMapping(data, db, statusList)
	if err != nil {
		return err
	}
	customStatusMap := getStatusMapping(data)
	logger.Info("convert changelog :%d", data.Options.WorkspaceId)
	issueIdGen := didgen.NewDomainIdGenerator(&models.TapdIssue{})
	accountIdGen := didgen.NewDomainIdGenerator(&models.TapdAccount{})
	clIdGen := didgen.NewDomainIdGenerator(&models.TapdBugChangelog{})
	clauses := []dal.Clause{
		dal.Select("tc.created, tc.id, tc.workspace_id, tc.bug_id, tc.author, _tool_tapd_bug_changelog_items.*"),
		dal.From(&models.TapdBugChangelogItem{}),
		dal.Join("left join _tool_tapd_bug_changelogs tc on tc.id = _tool_tapd_bug_changelog_items.changelog_id "),
		dal.Where("tc.connection_id = ? AND tc.workspace_id = ?", data.Options.ConnectionId, data.Options.WorkspaceId),
		dal.Orderby("created DESC"),
	}

	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	}
	defer cursor.Close()

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		InputRowType:       reflect.TypeOf(BugChangelogItemResult{}),
		Input:              cursor,
		Convert: func(inputRow interface{}) ([]interface{}, errors.Error) {
			cl := inputRow.(*BugChangelogItemResult)
			domainCl := &ticket.IssueChangelogs{
				DomainEntity: domainlayer.DomainEntity{
					Id: clIdGen.Generate(data.Options.ConnectionId, cl.Id, cl.Field),
				},
				IssueId:           issueIdGen.Generate(data.Options.ConnectionId, cl.BugId),
				AuthorId:          accountIdGen.Generate(data.Options.ConnectionId, cl.Author),
				AuthorName:        cl.Author,
				FieldId:           cl.Field,
				FieldName:         cl.Field,
				OriginalFromValue: cl.ValueBeforeParsed,
				OriginalToValue:   cl.ValueAfterParsed,
				CreatedDate:       *cl.Created,
			}
			if domainCl.FieldName == "status" {
				domainCl.OriginalFromValue = statusLanguageMap[domainCl.OriginalFromValue]
				domainCl.OriginalToValue = statusLanguageMap[domainCl.OriginalToValue]
				if len(customStatusMap) != 0 {
					domainCl.FromValue = customStatusMap[domainCl.OriginalFromValue]
					domainCl.ToValue = customStatusMap[domainCl.OriginalToValue]
				} else {
					domainCl.FromValue = getStdStatus(domainCl.OriginalFromValue)
					domainCl.ToValue = getStdStatus(domainCl.OriginalToValue)
				}
			}
			return []interface{}{
				domainCl,
			}, nil
		},
	})
	if err != nil {
		logger.Info(err.Error())
		return err
	}

	return converter.Execute()
}

var ConvertBugChangelogMeta = core.SubTaskMeta{
	Name:             "convertBugChangelog",
	EntryPoint:       ConvertBugChangelog,
	EnabledByDefault: true,
	Description:      "convert Tapd bug changelog",
	DomainTypes:      []string{core.DOMAIN_TYPE_TICKET},
}
