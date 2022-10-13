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

package migrationscripts

import (
	"context"
	"github.com/apache/incubator-devlake/errors"

	"github.com/apache/incubator-devlake/plugins/tapd/models/migrationscripts/archived"
	"gorm.io/gorm"
)

type addInitTables struct{}

func (*addInitTables) Up(ctx context.Context, db *gorm.DB) errors.Error {
	err := db.Migrator().DropTable(
		"_raw_tapd_api_bug_changelogs",
		"_raw_tapd_api_bugs",
		"_raw_tapd_api_bug_commits",
		"_raw_tapd_api_bug_custom_fields",
		"_raw_tapd_api_bug_status",
		"_raw_tapd_api_companies",
		"_raw_tapd_api_iterations",
		"_raw_tapd_api_story_bugs",
		"_raw_tapd_api_story_categories",
		"_raw_tapd_api_story_changelogs",
		"_raw_tapd_api_stories",
		"_raw_tapd_api_story_commits",
		"_raw_tapd_api_story_custom_fields",
		"_raw_tapd_api_story_status",
		"_raw_tapd_api_task_changelogs",
		"_raw_tapd_api_tasks",
		"_raw_tapd_api_task_commits",
		"_raw_tapd_api_task_custom_fields",
		"_raw_tapd_api_users",
		"_raw_tapd_api_worklogs",
		"_raw_tapd_api_workitem_types",
		"_raw_tapd_api_sub_workspaces",
		"_tool_tapd_users",
		&archived.TapdWorkspace{},
		&archived.TapdSubWorkspace{},
		&archived.TapdWorklog{},
		&archived.TapdWorkspaceIteration{},
		&archived.TapdBugChangelog{},
		&archived.TapdBugChangelogItem{},
		&archived.TapdStoryChangelog{},
		&archived.TapdStoryChangelogItem{},
		&archived.TapdTaskChangelog{},
		&archived.TapdTaskChangelogItem{},
		&archived.TapdIssue{},
		&archived.TapdIteration{},
		&archived.TapdConnection{},
		&archived.TapdBug{},
		&archived.TapdStory{},
		&archived.TapdTask{},
		&archived.TapdTaskLabel{},
		&archived.TapdBugLabel{},
		&archived.TapdStoryLabel{},
		&archived.TapdBugStatus{},
		&archived.TapdStoryStatus{},
		&archived.TapdBugCommit{},
		&archived.TapdStoryCommit{},
		&archived.TapdTaskCommit{},
		&archived.TapdWorkSpaceBug{},
		&archived.TapdWorkSpaceStory{},
		&archived.TapdWorkSpaceTask{},
		&archived.TapdIterationBug{},
		&archived.TapdIterationStory{},
		&archived.TapdIterationTask{},
		&archived.TapdStoryCustomFields{},
		&archived.TapdBugCustomFields{},
		&archived.TapdTaskCustomFields{},
		&archived.TapdStoryCategory{},
		&archived.TapdStoryBug{},
		&archived.TapdWorkitemType{},
	)
	if err != nil {
		return errors.Convert(err)
	}

	return errors.Convert(db.Migrator().AutoMigrate(
		&archived.TapdWorkspace{},
		&archived.TapdSubWorkspace{},
		&archived.TapdWorklog{},
		&archived.TapdWorkspaceIteration{},
		&archived.TapdAccount{},
		&archived.TapdBugChangelog{},
		&archived.TapdBugChangelogItem{},
		&archived.TapdStoryChangelog{},
		&archived.TapdStoryChangelogItem{},
		&archived.TapdTaskChangelog{},
		&archived.TapdTaskChangelogItem{},
		&archived.TapdIssue{},
		&archived.TapdIteration{},
		&archived.TapdConnection{},
		&archived.TapdBug{},
		&archived.TapdStory{},
		&archived.TapdTask{},
		&archived.TapdTaskLabel{},
		&archived.TapdBugLabel{},
		&archived.TapdStoryLabel{},
		&archived.TapdBugStatus{},
		&archived.TapdStoryStatus{},
		&archived.TapdBugCommit{},
		&archived.TapdStoryCommit{},
		&archived.TapdTaskCommit{},
		&archived.TapdWorkSpaceBug{},
		&archived.TapdWorkSpaceStory{},
		&archived.TapdWorkSpaceTask{},
		&archived.TapdIterationBug{},
		&archived.TapdIterationStory{},
		&archived.TapdIterationTask{},
		&archived.TapdStoryCustomFields{},
		&archived.TapdBugCustomFields{},
		&archived.TapdTaskCustomFields{},
		&archived.TapdStoryCategory{},
		&archived.TapdStoryBug{},
		&archived.TapdWorkitemType{},
	))
}

func (*addInitTables) Version() uint64 {
	return 20221013201138
}

func (*addInitTables) Name() string {
	return "Tapd init schemas"
}
