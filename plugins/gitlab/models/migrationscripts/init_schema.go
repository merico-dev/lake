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
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitlab/models/migrationscripts/archived"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InitSchemas struct{}

func (*InitSchemas) Up(ctx context.Context, db *gorm.DB) error {
	err := db.Migrator().DropTable(
		&archived.GitlabProject{},
		&archived.GitlabMergeRequest{},
		&archived.GitlabCommit{},
		&archived.GitlabTag{},
		&archived.GitlabProjectCommit{},
		&archived.GitlabPipeline{},
		&archived.GitlabReviewer{},
		&archived.GitlabMergeRequestNote{},
		&archived.GitlabMergeRequestCommit{},
		&archived.GitlabMergeRequestComment{},
		&archived.GitlabUser{},
		&archived.GitlabConnection{},
		&archived.GitlabIssue{},
		&archived.GitlabIssueLabel{},
		"_raw_gitlab_api_children_on_pipeline",
		"_raw_gitlab_api_commit",
		"_raw_gitlab_api_issues",
		"_raw_gitlab_api_merge_request_commits",
		"_raw_gitlab_api_merge_request_notes",
		"_raw_gitlab_api_merge_requests",
		"_raw_gitlab_api_pipeline",
		"_raw_gitlab_api_project",
		"_raw_gitlab_api_tag",
	)

	if err != nil {
		return err
	}

	err = db.Migrator().AutoMigrate(
		&archived.GitlabProject{},
		&archived.GitlabMergeRequest{},
		&archived.GitlabCommit{},
		&archived.GitlabTag{},
		&archived.GitlabProjectCommit{},
		&archived.GitlabPipeline{},
		&archived.GitlabReviewer{},
		&archived.GitlabMergeRequestNote{},
		&archived.GitlabMergeRequestCommit{},
		&archived.GitlabMergeRequestComment{},
		&archived.GitlabUser{},
		&archived.GitlabConnection{},
		&archived.GitlabIssue{},
		&archived.GitlabIssueLabel{},
	)

	if err != nil {
		return err
	}

	v := config.GetConfig()
	encKey := v.GetString("ENCODE_KEY")
	endPoint := v.GetString("GITLAB_ENDPOINT")
	gitlabAuth := v.GetString("GITLAB_AUTH")

	if encKey == "" || endPoint == "" || gitlabAuth == "" {
		return nil
	} else {
		conn := &archived.GitlabConnection{}
		conn.Name = "init gitlab connection"
		conn.ID = 1
		conn.Endpoint = endPoint
		conn.Token, err = core.Encrypt(encKey, gitlabAuth)
		if err != nil {
			return err
		}
		conn.Proxy = v.GetString("GITLAB_PROXY")
		conn.RateLimit = v.GetInt("GITLAB_API_REQUESTS_PER_HOUR")

		err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(conn).Error

		if err != nil {
			return err
		}
	}

	return nil
}

func (*InitSchemas) Version() uint64 {
	return 20220614231236
}

func (*InitSchemas) Name() string {
	return "Gitlab init schemas"
}
