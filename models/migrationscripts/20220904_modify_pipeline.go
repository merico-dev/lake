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
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/models/common"
	"github.com/apache/incubator-devlake/plugins/core"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

type Pipeline0904Temp struct {
	common.Model
	Name          string     `json:"name" gorm:"index"`
	BlueprintId   uint64     `json:"blueprintId"`
	Plan          string     `json:"plan" encrypt:"yes"`
	TotalTasks    int        `json:"totalTasks"`
	FinishedTasks int        `json:"finishedTasks"`
	BeganAt       *time.Time `json:"beganAt"`
	FinishedAt    *time.Time `json:"finishedAt" gorm:"index"`
	Status        string     `json:"status"`
	Message       string     `json:"message"`
	SpentSeconds  int        `json:"spentSeconds"`
	Stage         int        `json:"stage"`
}

func (Pipeline0904Temp) TableName() string {
	return "_devlake_pipeline_0904_tmp"
}

type PipelineOldVersion struct {
	common.Model
	Name          string         `json:"name" gorm:"index"`
	BlueprintId   uint64         `json:"blueprintId"`
	Plan          datatypes.JSON `json:"plan"`
	TotalTasks    int            `json:"totalTasks"`
	FinishedTasks int            `json:"finishedTasks"`
	BeganAt       *time.Time     `json:"beganAt"`
	FinishedAt    *time.Time     `json:"finishedAt" gorm:"index"`
	Status        string         `json:"status"`
	Message       string         `json:"message"`
	SpentSeconds  int            `json:"spentSeconds"`
	Stage         int            `json:"stage"`
}

func (PipelineOldVersion) TableName() string {
	return "_devlake_pipelines"
}

type modifyPipeline struct{}

func (*modifyPipeline) Up(ctx context.Context, db *gorm.DB) error {
	err := db.Migrator().CreateTable(&Pipeline0904Temp{})
	if err != nil {
		return err
	}
	defer db.Migrator().DropTable(&Pipeline0904Temp{})

	var result *gorm.DB
	var pipelineList []PipelineOldVersion
	result = db.Find(&pipelineList)

	if result.Error != nil {
		return result.Error
	}

	for _, v := range pipelineList {
		c := config.GetConfig()
		encKey := c.GetString(core.EncodeKeyEnvStr)
		if encKey == "" {
			return errors.BadInput.New("invalid encKey", errors.AsUserMessage())
		}
		encryptedPlan, err := core.Encrypt(encKey, string(v.Plan))
		if err != nil {
			return err
		}
		newPipeline := &Pipeline0904Temp{
			Name:          v.Name,
			BlueprintId:   v.BlueprintId,
			FinishedTasks: v.FinishedTasks,
			BeganAt:       v.BeganAt,
			FinishedAt:    v.FinishedAt,
			Status:        v.Status,
			Message:       v.Message,
			SpentSeconds:  v.SpentSeconds,
			Stage:         v.Stage,
			Plan:          encryptedPlan,
		}
		err = db.Create(newPipeline).Error
	}

	err = db.Migrator().DropTable(&PipelineOldVersion{})
	if err != nil {
		return err
	}

	err = db.Migrator().RenameTable(Pipeline0904Temp{}, PipelineOldVersion{})
	if err != nil {
		return err
	}

	return nil
}

func (*modifyPipeline) Version() uint64 {
	return 20220904162121
}

func (*modifyPipeline) Name() string {
	return "modifyPipeline"
}
