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

package archived

import (
	"time"

	"github.com/apache/incubator-devlake/models/migrationscripts/archived"
)

type GithubPullRequestComment struct {
	ConnectionId    uint64 `gorm:"primaryKey"`
	GithubId        int    `gorm:"primaryKey"`
	PullRequestId   int    `gorm:"index"`
	Body            string
	AuthorUsername  string `gorm:"type:varchar(255)"`
	AuthorUserId    int
	GithubCreatedAt time.Time
	GithubUpdatedAt time.Time `gorm:"index"`
	archived.NoPKModel
}

func (GithubPullRequestComment) TableName() string {
	return "_tool_github_pull_request_comments"
}
