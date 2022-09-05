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
	"encoding/json"
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/models/migrationscripts/archived"
	"github.com/apache/incubator-devlake/plugins/core"
	"gorm.io/gorm"
)

type Blueprint0903Temp struct {
	Name   string `json:"name" validate:"required"`
	Mode   string `json:"mode" gorm:"varchar(20)" validate:"required,oneof=NORMAL ADVANCED"`
	Enable bool   `json:"enable"`
	//please check this https://crontab.guru/ for detail
	CronConfig     string `json:"cronConfig" format:"* * * * *" example:"0 0 * * 1"`
	IsManual       bool   `json:"isManual"`
	archived.Model `swaggerignore:"true"`
	Plan           string `json:"plan"`
	Settings       string `json:"settings"`
}

func (Blueprint0903Temp) TableName() string {
	return "_devlake_blueprints_tmp"
}

type BlueprintOldVersion struct {
	Name   string          `json:"name" validate:"required"`
	Mode   string          `json:"mode" gorm:"varchar(20)" validate:"required,oneof=NORMAL ADVANCED"`
	Plan   json.RawMessage `json:"plan"`
	Enable bool            `json:"enable"`
	//please check this https://crontab.guru/ for detail
	CronConfig     string          `json:"cronConfig" format:"* * * * *" example:"0 0 * * 1"`
	IsManual       bool            `json:"isManual"`
	Settings       json.RawMessage `json:"settings" swaggertype:"array,string" example:"please check api: /blueprints/<PLUGIN_NAME>/blueprint-setting"`
	archived.Model `swaggerignore:"true"`
}

func (BlueprintOldVersion) TableName() string {
	return "_devlake_blueprints"
}

type modifyBLueprint struct{}

func (*modifyBLueprint) Up(ctx context.Context, db *gorm.DB) error {
	err := db.Migrator().CreateTable(&Blueprint0903Temp{})
	if err != nil {
		return err
	}
	defer db.Migrator().DropTable(&Blueprint0903Temp{})

	var result *gorm.DB
	var blueprintList []BlueprintOldVersion
	result = db.Find(&blueprintList)

	if result.Error != nil {
		return result.Error
	}

	for _, v := range blueprintList {
		c := config.GetConfig()
		encKey := c.GetString(core.EncodeKeyEnvStr)
		if encKey == "" {
			return errors.BadInput.New("invalid encKey", errors.AsUserMessage())
		}
		encryptedPlan, err := core.Encrypt(encKey, string(v.Plan))
		if err != nil {
			return err
		}
		encryptedSettings, err := core.Encrypt(encKey, string(v.Settings))
		if err != nil {
			return err
		}
		newBlueprint := &Blueprint0903Temp{
			Name:       v.Name,
			Mode:       v.Mode,
			Enable:     v.Enable,
			CronConfig: v.CronConfig,
			IsManual:   v.IsManual,
			Model:      archived.Model{ID: v.ID},
			Plan:       encryptedPlan,
			Settings:   encryptedSettings,
		}
		err = db.Create(newBlueprint).Error
	}

	err = db.Migrator().DropTable(&BlueprintOldVersion{})
	if err != nil {
		return err
	}

	err = db.Migrator().RenameTable(Blueprint0903Temp{}, BlueprintOldVersion{})
	if err != nil {
		return err
	}

	return nil
}

func (*modifyBLueprint) Version() uint64 {
	return 20220904142321
}

func (*modifyBLueprint) Name() string {
	return "modifyBlueprint"
}
