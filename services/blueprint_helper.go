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

package services

import (
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/models"
	"github.com/apache/incubator-devlake/models/common"
	"github.com/apache/incubator-devlake/plugins/core"
	"gorm.io/gorm"
)

// CreateApiBlueprint accepts a ApiBlueprint instance and insert it to database
func CreateBlueprint(blueprintDO *models.Blueprint) error {
	err := db.Create(&blueprintDO).Error
	if err != nil {
		return err
	}
	return nil
}

// GetApiBlueprints returns a paginated list of Blueprints based on `query`
func GetBlueprintDOs(query *BlueprintQuery) ([]*models.Blueprint, int64, error) {
	blueprintDOs := make([]*models.Blueprint, 0)
	db := db.Model(blueprintDOs).Order("id DESC")
	if query.Enable != nil {
		db = db.Where("enable = ?", *query.Enable)
	}

	var count int64
	err := db.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	if query.Page > 0 && query.PageSize > 0 {
		offset := query.PageSize * (query.Page - 1)
		db = db.Limit(query.PageSize).Offset(offset)
	}
	err = db.Find(&blueprintDOs).Error
	if err != nil {
		return nil, 0, err
	}

	return blueprintDOs, count, nil
}

// GetApiBlueprint returns the detail of a given ApiBlueprint ID
func GetBlueprint(blueprintDOId uint64) (*models.Blueprint, error) {
	blueprintDO := &models.Blueprint{}
	err := db.First(blueprintDO, blueprintDOId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return blueprintDO, nil
}

// DeleteApiBlueprint FIXME ...
func DeleteBlueprint(id uint64) error {
	err := db.Delete(&models.Blueprint{}, "id = ?", id).Error
	if err != nil {
		return err
	}
	return nil
}

// parseBlueprint
func parseBlueprint(blueprintDO *models.Blueprint) (*models.ApiBlueprint, error) {
	blueprint := models.ApiBlueprint{
		Name:       blueprintDO.Name,
		Mode:       blueprintDO.Mode,
		Plan:       []byte(blueprintDO.Plan),
		Enable:     blueprintDO.Enable,
		CronConfig: blueprintDO.CronConfig,
		IsManual:   blueprintDO.IsManual,
		Settings:   []byte(blueprintDO.Settings),
		Model: common.Model{
			ID: blueprintDO.ID,
		},
	}
	return &blueprint, nil
}

// parseApiBlueprint
func parseApiBlueprint(blueprint *models.ApiBlueprint) (*models.Blueprint, error) {
	blueprintDO := models.Blueprint{
		Name:       blueprint.Name,
		Mode:       blueprint.Mode,
		Plan:       string(blueprint.Plan),
		Enable:     blueprint.Enable,
		CronConfig: blueprint.CronConfig,
		IsManual:   blueprint.IsManual,
		Settings:   string(blueprint.Settings),
		Model: common.Model{
			ID: blueprint.ID,
		},
	}
	return &blueprintDO, nil
}

// encryptBlueprint
func encryptBlueprint(blueprintDO *models.Blueprint) (*models.Blueprint, error) {
	encKey := config.GetConfig().GetString(core.EncodeKeyEnvStr)
	planEncrypt, err := core.Encrypt(encKey, blueprintDO.Plan)
	if err != nil {
		return nil, err
	}
	blueprintDO.Plan = planEncrypt
	settingsEncrypt, err := core.Encrypt(encKey, blueprintDO.Settings)
	blueprintDO.Settings = settingsEncrypt
	if err != nil {
		return nil, err
	}
	return blueprintDO, nil
}

// decryptBlueprint
func decryptBlueprint(blueprintDO *models.Blueprint) (*models.Blueprint, error) {
	encKey := config.GetConfig().GetString(core.EncodeKeyEnvStr)
	plan, err := core.Decrypt(encKey, blueprintDO.Plan)
	if err != nil {
		return nil, err
	}
	blueprintDO.Plan = plan
	settings, err := core.Decrypt(encKey, blueprintDO.Settings)
	blueprintDO.Settings = settings
	if err != nil {
		return nil, err
	}
	return blueprintDO, nil
}
