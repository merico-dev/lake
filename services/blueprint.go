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
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"strings"

	"github.com/apache/incubator-devlake/logger"
	"github.com/apache/incubator-devlake/models"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/go-playground/validator/v10"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// BlueprintQuery FIXME ...
type BlueprintQuery struct {
	Enable   *bool `form:"enable,omitempty"`
	Page     int   `form:"page"`
	PageSize int   `form:"pageSize"`
}

var (
	blueprintLog = logger.Global.Nested("blueprint")
	vld          = validator.New()
)

// CreateApiBlueprint accepts a ApiBlueprint instance and insert it to database
func CreateApiBlueprint(apiBlueprint *models.ApiBlueprint) error {
	err := validateApiBlueprint(apiBlueprint)
	if err != nil {
		return err
	}
	blueprint, err := parseApiBlueprint(apiBlueprint)
	if err != nil {
		return err
	}
	blueprint, err = encryptBlueprint(blueprint)
	if err != nil {
		return err
	}
	err = CreateBlueprint(blueprint)
	if err != nil {
		return err
	}
	err = ReloadBlueprints(cronManager)
	if err != nil {
		return errors.Internal.Wrap(err, "error reloading blueprints")
	}
	return nil
}

// GetApiBlueprints returns a paginated list of Blueprints based on `query`
func GetApiBlueprints(query *BlueprintQuery) ([]*models.ApiBlueprint, int64, error) {
	blueprintDOs, count, err := GetBlueprintDOs(query)
	if err != nil {
		return nil, 0, err
	}
	apiBlueprints := make([]*models.ApiBlueprint, 0)
	for _, blueprint := range blueprintDOs {
		blueprint, err = decryptBlueprint(blueprint)
		if err != nil {
			return nil, 0, err
		}
		apiBlueprint, err := parseBlueprint(blueprint)
		if err != nil {
			return nil, 0, err
		}
		apiBlueprints = append(apiBlueprints, apiBlueprint)
	}
	return apiBlueprints, count, nil
}

// GetApiBlueprint returns the detail of a given ApiBlueprint ID
func GetApiBlueprint(blueprintId uint64) (*models.ApiBlueprint, error) {
	blueprint, err := GetBlueprint(blueprintId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound.New("blueprint not found", errors.AsUserMessage())
		}
		return nil, errors.Internal.Wrap(err, "error getting the task from database", errors.AsUserMessage())
	}
	blueprint, err = decryptBlueprint(blueprint)
	if err != nil {
		return nil, err
	}
	apiBlueprint, err := parseBlueprint(blueprint)
	return apiBlueprint, nil
}

func validateApiBlueprint(apiBlueprint *models.ApiBlueprint) error {
	// validation
	err := vld.Struct(apiBlueprint)
	if err != nil {
		return err
	}
	if strings.ToLower(apiBlueprint.CronConfig) == "manual" {
		apiBlueprint.IsManual = true
	}

	if !apiBlueprint.IsManual {
		_, err = cron.ParseStandard(apiBlueprint.CronConfig)
		if err != nil {
			return errors.Default.Wrap(err, "invalid cronConfig")
		}
	}
	if apiBlueprint.Mode == models.BLUEPRINT_MODE_ADVANCED {
		plan := make(core.PipelinePlan, 0)
		err = json.Unmarshal(apiBlueprint.Plan, &plan)

		if err != nil {
			return errors.Default.Wrap(err, "invalid plan")
		}
		// tasks should not be empty
		if len(plan) == 0 || len(plan[0]) == 0 {
			return errors.Default.New("empty plan")
		}
	} else if apiBlueprint.Mode == models.BLUEPRINT_MODE_NORMAL {
		apiBlueprint.Plan, err = GeneratePlanJson(apiBlueprint.Settings)
		if err != nil {
			return errors.Default.Wrap(err, "invalid plan")
		}
	}

	return nil
}

// PatchApiBlueprint FIXME ...
func PatchApiBlueprint(id uint64, body map[string]interface{}) (*models.ApiBlueprint, error) {
	// load record from db
	apiBlueprint, err := GetApiBlueprint(id)
	if err != nil {
		return nil, err
	}
	originMode := apiBlueprint.Mode
	err = helper.DecodeMapStruct(body, apiBlueprint)

	if err != nil {
		return nil, err
	}
	// make sure mode is not being update
	if originMode != apiBlueprint.Mode {
		return nil, errors.Default.New("mode is not updatable")
	}
	// validation
	err = validateApiBlueprint(apiBlueprint)
	if err != nil {
		return nil, err
	}

	// save
	err = save(apiBlueprint)
	if err != nil {
		return nil, errors.Internal.Wrap(err, "error saving blueprint")
	}

	// reload schedule
	err = ReloadBlueprints(cronManager)
	if err != nil {
		return nil, errors.Internal.Wrap(err, "error reloading blueprints")
	}
	// done
	return apiBlueprint, nil
}

// DeleteApiBlueprint FIXME ...
func DeleteApiBlueprint(id uint64) error {
	err := DeleteBlueprint(id)
	if err != nil {
		return errors.Internal.Wrap(err, fmt.Sprintf("error deleting blueprint %d", id))
	}
	err = ReloadBlueprints(cronManager)
	if err != nil {
		return errors.Internal.Wrap(err, "error reloading blueprints")
	}
	return nil
}

// ReloadBlueprints FIXME ...
func ReloadBlueprints(c *cron.Cron) error {
	blueprintDOs := make([]*models.Blueprint, 0)
	err := db.Model(&models.Blueprint{}).
		Where("enable = ? AND is_manual = ?", true, false).
		Find(&blueprintDOs).Error
	if err != nil {
		panic(err)
	}
	for _, e := range c.Entries() {
		c.Remove(e.ID)
	}
	c.Stop()
	for _, pp := range blueprintDOs {
		pp, err = decryptBlueprint(pp)
		if err != nil {
			return err
		}
		apiBlueprint, err := parseBlueprint(pp)
		plan, err := apiBlueprint.UnmarshalPlan()
		if err != nil {
			blueprintLog.Error(err, "created cron job failed")
			return err
		}
		_, err = c.AddFunc(apiBlueprint.CronConfig, func() {
			pipeline, err := createPipelineByBlueprint(apiBlueprint.ID, apiBlueprint.Name, plan)
			if err != nil {
				blueprintLog.Error(err, "run cron job failed")
			} else {
				blueprintLog.Info("Run new cron job successfully, pipeline id: %d", pipeline.ID)
			}
		})
		if err != nil {
			blueprintLog.Error(err, "created cron job failed")
			return err
		}
	}
	if len(blueprintDOs) > 0 {
		c.Start()
	}
	log.Info("total %d blueprints were scheduled", len(blueprintDOs))
	return nil
}

func createPipelineByBlueprint(blueprintId uint64, name string, plan core.PipelinePlan) (*models.ApiPipeline, error) {
	newPipeline := models.NewPipeline{}
	newPipeline.Plan = plan
	newPipeline.Name = name
	newPipeline.BlueprintId = blueprintId
	pipeline, err := CreatePipeline(&newPipeline)
	// Return all created tasks to the User
	if err != nil {
		blueprintLog.Error(err, "created cron job failed")
		return nil, err
	}
	return pipeline, err
}

// GeneratePlanJson generates pipeline plan by version
func GeneratePlanJson(settings json.RawMessage) (json.RawMessage, error) {
	bpSettings := new(models.BlueprintSettings)
	err := json.Unmarshal(settings, bpSettings)

	if err != nil {
		return nil, fmt.Errorf("settings:%s:%s", string(settings), err.Error())
	}
	var plan interface{}
	switch bpSettings.Version {
	case "1.0.0":
		plan, err = GeneratePlanJsonV100(bpSettings)
	default:
		return nil, errors.Default.New(fmt.Sprintf("unknown version of blueprint settings: %s", bpSettings.Version))
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(plan)
}

// GeneratePlanJsonV100 generates pipeline plan according v1.0.0 definition
func GeneratePlanJsonV100(settings *models.BlueprintSettings) (core.PipelinePlan, error) {
	connections := make([]*core.BlueprintConnectionV100, 0)
	err := json.Unmarshal(settings.Connections, &connections)
	if err != nil {
		return nil, err
	}

	plans := make([]core.PipelinePlan, len(connections))
	for i, connection := range connections {
		if len(connection.Scope) == 0 {
			return nil, errors.Default.New(fmt.Sprintf("connections[%d].scope is empty", i))
		}
		plugin, err := core.GetPlugin(connection.Plugin)
		if err != nil {
			return nil, err
		}
		if pluginBp, ok := plugin.(core.PluginBlueprintV100); ok {
			plans[i], err = pluginBp.MakePipelinePlan(connection.ConnectionId, connection.Scope)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.Default.New(fmt.Sprintf("plugin %s does not support blueprint protocol version 1.0.0", connection.Plugin))
		}
	}

	mergedPipelinePlan := MergePipelinePlans(plans...)
	return FormatPipelinePlans(settings.BeforePlan, mergedPipelinePlan, settings.AfterPlan)
}

// FormatPipelinePlans merges multiple pipelines and append before and after pipeline
func FormatPipelinePlans(beforePlanJson json.RawMessage, mainPlan core.PipelinePlan, afterPlanJson json.RawMessage) (core.PipelinePlan, error) {
	newPipelinePlan := core.PipelinePlan{}
	if beforePlanJson != nil {
		beforePipelinePlan := core.PipelinePlan{}
		err := json.Unmarshal(beforePlanJson, &beforePipelinePlan)
		if err != nil {
			return nil, err
		}
		newPipelinePlan = append(newPipelinePlan, beforePipelinePlan...)
	}

	newPipelinePlan = append(newPipelinePlan, mainPlan...)

	if afterPlanJson != nil {
		afterPipelinePlan := core.PipelinePlan{}
		err := json.Unmarshal(afterPlanJson, &afterPipelinePlan)
		if err != nil {
			return nil, err
		}
		newPipelinePlan = append(newPipelinePlan, afterPipelinePlan...)
	}
	return newPipelinePlan, nil
}

// MergePipelinePlans merges multiple pipelines into one unified pipeline
func MergePipelinePlans(plans ...core.PipelinePlan) core.PipelinePlan {
	merged := make(core.PipelinePlan, 0)
	// iterate all pipelineTasks and try to merge them into `merged`
	for _, plan := range plans {
		// add all stages from plan to merged
		for index, stage := range plan {
			if index >= len(merged) {
				merged = append(merged, nil)
			}
			// add all tasks from plan to target respectively
			merged[index] = append(merged[index], stage...)
		}
	}
	return merged
}

// TriggerBlueprint triggers blueprint immediately
func TriggerBlueprint(id uint64) (*models.ApiPipeline, error) {
	// load record from db
	apiBlueprint, err := GetApiBlueprint(id)
	if err != nil {
		return nil, err
	}
	plan, err := apiBlueprint.UnmarshalPlan()
	if err != nil {
		return nil, err
	}
	pipeline, err := createPipelineByBlueprint(apiBlueprint.ID, apiBlueprint.Name, plan)
	// done
	return pipeline, err
}
func save(apiBlueprint *models.ApiBlueprint) error {
	blueprint, err := parseApiBlueprint(apiBlueprint)
	if err != nil {
		return nil
	}
	blueprint, err = encryptBlueprint(blueprint)
	if err != nil {
		return err
	}
	return db.Save(blueprint).Error
}
