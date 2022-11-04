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
	"time"
)

type GithubRun struct {
	ConnectionId     uint64     `gorm:"primaryKey"`
	RepoId           int        `gorm:"primaryKey"`
	ID               int64      `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Name             string     `json:"name" gorm:"type:varchar(255)"`
	NodeID           string     `json:"node_id" gorm:"type:varchar(255)"`
	HeadBranch       string     `json:"head_branch" gorm:"type:varchar(255)"`
	HeadSha          string     `json:"head_sha" gorm:"type:varchar(255)"`
	Path             string     `json:"path" gorm:"type:varchar(255)"`
	RunNumber        int        `json:"run_number"`
	Event            string     `json:"event" gorm:"type:varchar(255)"`
	Status           string     `json:"status" gorm:"type:varchar(255)"`
	Conclusion       string     `json:"conclusion" gorm:"type:varchar(255)"`
	WorkflowID       int        `json:"workflow_id"`
	CheckSuiteID     int64      `json:"check_suite_id"`
	CheckSuiteNodeID string     `json:"check_suite_node_id" gorm:"type:varchar(255)"`
	URL              string     `json:"url" gorm:"type:varchar(255)"`
	HTMLURL          string     `json:"html_url" gorm:"type:varchar(255)"`
	GithubCreatedAt  *time.Time `json:"created_at"`
	GithubUpdatedAt  *time.Time `json:"updated_at"`
	RunAttempt       int        `json:"run_attempt"`
	RunStartedAt     *time.Time `json:"run_started_at"`
	JobsURL          string     `json:"jobs_url" gorm:"type:varchar(255)"`
	LogsURL          string     `json:"logs_url" gorm:"type:varchar(255)"`
	CheckSuiteURL    string     `json:"check_suite_url" gorm:"type:varchar(255)"`
	ArtifactsURL     string     `json:"artifacts_url" gorm:"type:varchar(255)"`
	CancelURL        string     `json:"cancel_url" gorm:"type:varchar(255)"`
	RerunURL         string     `json:"rerun_url" gorm:"type:varchar(255)"`
	WorkflowURL      string     `json:"workflow_url" gorm:"type:varchar(255)"`
	Type             string     `json:"type" gorm:"type:varchar(255)"`
	archived.NoPKModel
}

func (GithubRun) TableName() string {
	return "_tool_github_runs"
}
