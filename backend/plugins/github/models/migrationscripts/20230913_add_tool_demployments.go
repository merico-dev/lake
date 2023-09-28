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
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/migrationscripts/archived"
	"time"
)

type GithubDeployment struct {
	archived.NoPKModel `json:"-" mapstructure:"-"`
	ConnectionId       uint64    `json:"connection_id" gorm:"primaryKey"`
	GithubId           int       `json:"github_id" gorm:"type:varchar(255)"`
	Id                 string    `json:"id" gorm:"type:varchar(255);primaryKey"`
	DatabaseId         uint      `json:"database_id"`
	CommitOid          string    `json:"commit_oid" gorm:"type:varchar(255)"`
	Description        string    `json:"description" gorm:"type:varchar(255)"`
	Environment        string    `json:"environment" gorm:"type:varchar(255)"`
	State              string    `json:"state" gorm:"type:varchar(255)"`
	LatestStatusState  string    `json:"latest_status_state" gorm:"type:varchar(255)"`
	LatestUpdatedDate  time.Time `json:"latest_status_update_date"`
	RepositoryID       string    `json:"repository_id" gorm:"type:varchar(255)"`
	RepositoryName     string    `json:"repository_name" gorm:"type:varchar(255)"`
	RepositoryUrl      string    `json:"repository_url" gorm:"type:varchar(255)"`
	RefName            string    `json:"ref_name" gorm:"type:varchar(255)"`
	Payload            string    `json:"payload" gorm:"type:text"`
	CreatedDate        time.Time `json:"created_at"`
	UpdatedDate        time.Time `json:"updated_at"`
}

func (GithubDeployment) TableName() string {
	return "_tool_github_deployments"
}

type addDeploymentTable struct {
}

func (*addDeploymentTable) Up(basicRes context.BasicRes) errors.Error {
	return basicRes.GetDal().AutoMigrate(&GithubDeployment{})
}

func (*addDeploymentTable) Version() uint64 {
	return 20230913170100
}
func (*addDeploymentTable) Name() string {
	return "add github deployment table"
}
