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
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/helpers/migrationhelper"
	"github.com/apache/incubator-devlake/models/migrationscripts/archived"
	"github.com/apache/incubator-devlake/plugins/core"
)

type modifyAllEntities struct{}

type jenkinsJobDag20220729 struct {
	ConnetionId   uint64 `gorm:"primaryKey"`
	UpstreamJob   string `gorm:"primaryKey;type:varchar(255)"`
	DownstreamJob string `gorm:"primaryKey;type:varchar(255)"`
	archived.NoPKModel
}

func (jenkinsJobDag20220729) TableName() string {
	return "_tool_jenkins_job_dags"
}

// JenkinsBuild db entity for jenkins build
type jenkinsBuild20220729 struct {
	TriggeredBy string `gorm:"type:varchar(255)"`
	Type        string `gorm:"index;type:varchar(255)" `
	Class       string `gorm:"index;type:varchar(255)" `
	HasStages   bool
	Building    bool
}

func (jenkinsBuild20220729) TableName() string {
	return "_tool_jenkins_builds"
}

type jenkinsBuildRepo20220729 struct {
	ConnectionId uint64 `gorm:"primaryKey"`
	BuildName    string `gorm:"primaryKey;type:varchar(255)"`
	CommitSha    string `gorm:"primaryKey;type:varchar(255)"`
	Branch       string `gorm:"type:varchar(255)"`
	RepoUrl      string `gorm:"type:varchar(255)"`
	archived.NoPKModel
}

func (jenkinsBuildRepo20220729) TableName() string {
	return "_tool_jenkins_build_repos"
}

type jenkinsStage20200729 struct {
	archived.NoPKModel
	ConnectionId        uint64 `gorm:"primaryKey"`
	ID                  string `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Name                string `json:"name" gorm:"type:varchar(255)"`
	ExecNode            string `json:"execNode" gorm:"type:varchar(255)"`
	Status              string `json:"status" gorm:"type:varchar(255)"`
	StartTimeMillis     int64  `json:"startTimeMillis"`
	DurationMillis      int    `json:"durationMillis"`
	PauseDurationMillis int    `json:"pauseDurationMillis"`
	BuildName           string `gorm:"primaryKey;type:varchar(255)"`
	Type                string `gorm:"index;type:varchar(255)"`
}

func (jenkinsStage20200729) TableName() string {
	return "_tool_jenkins_stages"
}

func (*modifyAllEntities) Up(basicRes core.BasicRes) errors.Error {

	return migrationhelper.AutoMigrateTables(
		basicRes,
		jenkinsBuild20220729{},
		jenkinsJobDag20220729{},
		jenkinsBuildRepo20220729{},
		jenkinsStage20200729{},
	)
}

func (*modifyAllEntities) Version() uint64 {
	return 20220729231237
}

func (*modifyAllEntities) Name() string {
	return "Jenkins modify build and job"
}
