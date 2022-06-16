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

package models

import (
	"github.com/apache/incubator-devlake/models/common"
	"gorm.io/datatypes"
)

const BLUEPRINT_CRON_MANUAL = "MANUAL"
const BLUEPRINT_MODE_NORMAL = "NORMAL"
const BLUEPRINT_MODE_ADVANCED = "ADVANCED"

type Blueprint struct {
	Name       string         `json:"name" validate:"required"`
	Mode       string         `json:"mode" gorm:"varchar(20)" validate:"required,oneof=NORMAL ADVANCED"`
	Tasks      datatypes.JSON `json:"tasks"`
	Enable     bool           `json:"enable"`
	CronConfig string         `json:"cronConfig" validate:"required"`
	common.Model
}

func (Blueprint) TableName() string {
	return "_devlake_blueprints"
}
