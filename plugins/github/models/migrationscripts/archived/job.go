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
	"github.com/apache/incubator-devlake/models/migrationscripts/archived"
	"gorm.io/datatypes"
	"time"
)

type GithubJob struct {
	archived.NoPKModel
	ConnectionId  uint64         `gorm:"primaryKey"`
	RepoId        int            `gorm:"primaryKey"`
	ID            int            `json:"id" gorm:"primaryKey;autoIncrement:false"`
	RunID         int            `json:"run_id"`
	RunURL        string         `json:"run_url" gorm:"type:varchar(255)"`
	NodeID        string         `json:"node_id" gorm:"type:varchar(255)"`
	HeadSha       string         `json:"head_sha" gorm:"type:varchar(255)"`
	URL           string         `json:"url" gorm:"type:varchar(255)"`
	HTMLURL       string         `json:"html_url" gorm:"type:varchar(255)"`
	Status        string         `json:"status" gorm:"type:varchar(255)"`
	Conclusion    string         `json:"conclusion" gorm:"type:varchar(255)"`
	StartedAt     *time.Time     `json:"started_at"`
	CompletedAt   *time.Time     `json:"completed_at"`
	Name          string         `json:"name" gorm:"type:varchar(255)"`
	Steps         datatypes.JSON `json:"steps"`
	CheckRunURL   string         `json:"check_run_url" gorm:"type:varchar(255)"`
	Labels        datatypes.JSON `json:"labels"`
	RunnerID      int            `json:"runner_id"`
	RunnerName    string         `json:"runner_name" gorm:"type:varchar(255)"`
	RunnerGroupID int            `json:"runner_group_id"`
	Type          string         `json:"type" gorm:"type:varchar(255)"`
}

func (GithubJob) TableName() string {
	return "_tool_github_jobs"
}
