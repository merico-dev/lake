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
)

// @Description CronConfig
type BlueprintDO struct {
	Name   string `json:"name" validate:"required"`
	Mode   string `json:"mode" gorm:"varchar(20)" validate:"required,oneof=NORMAL ADVANCED"`
	Plan   string `json:"plan" encrypt:"yes"`
	Enable bool   `json:"enable"`
	//please check this https://crontab.guru/ for detail
	CronConfig   string `json:"cronConfig" format:"* * * * *" example:"0 0 * * 1"`
	IsManual     bool   `json:"isManual"`
	Settings     string `json:"settings" encrypt:"yes" swaggertype:"array,string" example:"please check api: /blueprints/<PLUGIN_NAME>/blueprint-setting"`
	common.Model `swaggerignore:"true"`
}

func (BlueprintDO) TableName() string {
	return "_devlake_blueprints"
}
