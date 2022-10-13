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
	"github.com/apache/incubator-devlake/plugins/helper"
)

type TapdWorkitemType struct {
	ConnectionId   uint64          `gorm:"primaryKey;type:BIGINT  NOT NULL"`
	Id             uint64          `gorm:"primaryKey;type:BIGINT  NOT NULL" json:"id,string"`
	WorkspaceId    uint64          `json:"workspace_id,string"`
	EntityType     string          `gorm:"type:varchar(255)" json:"entity_type"`
	Name           string          `gorm:"type:varchar(255)" json:"name"`
	EnglishName    string          `gorm:"type:varchar(255)" json:"english_name"`
	Status         string          `gorm:"type:varchar(255)" json:"status"`
	Color          string          `gorm:"type:varchar(255)" json:"color"`
	WorkflowID     uint64          `json:"workflow_id"`
	Icon           string          `json:"icon"`
	IconSmall      string          `json:"icon_small"`
	Creator        string          `gorm:"type:varchar(255)" json:"creator"`
	Created        *helper.CSTTime `json:"created"`
	ModifiedBy     string          `gorm:"type:varchar(255)" json:"modified_by"`
	Modified       *helper.CSTTime `json:"modified"`
	IconViper      string          `json:"icon_viper"`
	IconSmallViper string          `json:"icon_small_viper"`
	archived.NoPKModel
}

func (TapdWorkitemType) TableName() string {
	return "_tool_tapd_workitem_types"
}
