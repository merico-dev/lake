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

type GiteeRepo struct {
	GiteeId       int    `gorm:"primaryKey"`
	Name          string `gorm:"type:varchar(255)"`
	HTMLUrl       string `gorm:"type:varchar(255)"`
	Description   string
	OwnerId       int        `json:"ownerId"`
	OwnerLogin    string     `json:"ownerLogin" gorm:"type:varchar(255)"`
	Language      string     `json:"language" gorm:"type:varchar(255)"`
	ParentGiteeId int        `json:"parentId"`
	ParentHTMLUrl string     `json:"parentHtmlUrl"`
	CreatedDate   time.Time  `json:"createdDate"`
	UpdatedDate   *time.Time `json:"updatedDate"`
	archived.NoPKModel
}

func (GiteeRepo) TableName() string {
	return "_tool_gitee_repo"
}
