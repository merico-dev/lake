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
	"github.com/apache/incubator-devlake/helpers/pluginhelper/services"
	"strings"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/impls/logruslog"
	"github.com/robfig/cron/v3"
)

var (
	blueprintLog = logruslog.Global.Nested("blueprint")
)

// BlueprintQuery is a query for GetBlueprints
type BlueprintQuery struct {
	Pagination
	Enable   *bool  `form:"enable,omitempty"`
	IsManual *bool  `form:"isManual"`
	Label    string `form:"label"`
}

type BlueprintJob struct {
	Blueprint *models.Blueprint
}

func (bj BlueprintJob) Run() {
	blueprint := bj.Blueprint
	pipeline, err := createPipelineByBlueprint(blueprint)
	if err != nil {
		blueprintLog.Error(err, fmt.Sprintf("run cron job failed on blueprint:[%d][%s]", blueprint.ID, blueprint.Name))
	} else {
		blueprintLog.Info("Run new cron job successfully,blueprint id:[%d] pipeline id:[%d]", blueprint.ID, pipeline.ID)
	}
}

// CreateBlueprint accepts a Blueprint instance and insert it to database
func CreateBlueprint(blueprint *models.Blueprint) errors.Error {
	err := validateBlueprintAndMakePlan(blueprint)
	if err != nil {
		return err
	}
	err = bpManager.SaveDbBlueprint(blueprint)
	if err != nil {
		return err
	}
	err = ReloadBlueprints(cronManager)
	if err != nil {
		return errors.Internal.Wrap(err, "error reloading blueprints")
	}
	return nil
}

// GetBlueprints returns a paginated list of Blueprints based on `query`
func GetBlueprints(query *BlueprintQuery) ([]*models.Blueprint, int64, errors.Error) {
	blueprints, count, err := bpManager.GetDbBlueprints(&services.GetBlueprintQuery{
		Enable:      query.Enable,
		IsManual:    query.IsManual,
		Label:       query.Label,
		SkipRecords: query.GetSkip(),
		PageSize:    query.GetPageSize(),
	})
	if err != nil {
		return nil, 0, errors.Convert(err)
	}
	return blueprints, count, nil
}

// GetBlueprint returns the detail of a given Blueprint ID
func GetBlueprint(blueprintId uint64) (*models.Blueprint, errors.Error) {
	blueprint, err := bpManager.GetDbBlueprint(blueprintId)
	if err != nil {
		if db.IsErrorNotFound(err) {
			return nil, errors.NotFound.New("blueprint not found")
		}
		return nil, errors.Internal.Wrap(err, "error getting the blueprint from database")
	}
	return blueprint, nil
}

// GetBlueprintByProjectName returns the detail of a given ProjectName
func GetBlueprintByProjectName(projectName string) (*models.Blueprint, errors.Error) {
	if projectName == "" {
		return nil, errors.Internal.New("can not use the empty projectName to search the unique blueprint")
	}
	blueprint, err := bpManager.GetDbBlueprintByProjectName(projectName)
	if err != nil {
		// Allow specific projectName to fail to find the corresponding blueprint
		if db.IsErrorNotFound(err) {
			return nil, nil
		}
		return nil, errors.Internal.Wrap(err, fmt.Sprintf("error getting the blueprint from database with project %s", projectName))
	}
	return blueprint, nil
}

func validateBlueprintAndMakePlan(blueprint *models.Blueprint) errors.Error {
	if len(blueprint.Settings) == 0 {
		blueprint.Settings = nil
	}
	// validation
	err := vld.Struct(blueprint)
	if err != nil {
		return errors.BadInput.WrapRaw(err)
	}

	// checking if the project exist
	if blueprint.ProjectName != "" {
		_, err := GetProject(blueprint.ProjectName)
		if err != nil {
			return errors.Default.Wrap(err, fmt.Sprintf("invalid projectName: [%s] for the blueprint [%s]", blueprint.ProjectName, blueprint.Name))
		}

		bp, err := GetBlueprintByProjectName(blueprint.ProjectName)
		if err != nil {
			return err
		}
		if bp != nil {
			if bp.ID != blueprint.ID {
				return errors.Default.New(fmt.Sprintf("Each project can only be used by one blueprint. The currently selected projectName: [%s] has been used by blueprint: [id:%d] [name:%s] and cannot be reused.", bp.ProjectName, bp.ID, bp.Name))
			}
		}
	}

	if strings.ToLower(blueprint.CronConfig) == "manual" {
		blueprint.IsManual = true
	}
	if !blueprint.IsManual {
		_, err = cron.ParseStandard(blueprint.CronConfig)
		if err != nil {
			return errors.Default.Wrap(err, "invalid cronConfig")
		}
	}
	if blueprint.Mode == models.BLUEPRINT_MODE_ADVANCED {
		plan := make(plugin.PipelinePlan, 0)
		err = errors.Convert(json.Unmarshal(blueprint.Plan, &plan))
		if err != nil {
			return errors.Default.Wrap(err, "invalid plan")
		}
		// tasks should not be empty
		if len(plan) == 0 || len(plan[0]) == 0 {
			return errors.Default.New("empty plan")
		}
	} else if blueprint.Mode == models.BLUEPRINT_MODE_NORMAL {
		plan, err := MakePlanForBlueprint(blueprint)
		if err != nil {
			return errors.Default.Wrap(err, "make plan for blueprint failed")
		}
		blueprint.Plan, err = errors.Convert01(json.Marshal(plan))
		if err != nil {
			return errors.Default.Wrap(err, "failed to markshal plan")
		}
	}
	return nil
}

func saveBlueprint(blueprint *models.Blueprint) (*models.Blueprint, errors.Error) {
	// validation
	err := validateBlueprintAndMakePlan(blueprint)
	if err != nil {
		return nil, errors.BadInput.WrapRaw(err)
	}
	err = bpManager.SaveDbBlueprint(blueprint)
	if err != nil {
		return nil, err
	}

	// reload schedule
	err = ReloadBlueprints(cronManager)
	if err != nil {
		return nil, errors.Internal.Wrap(err, "error reloading blueprints")
	}
	// done
	return blueprint, nil
}

// PatchBlueprint FIXME ...
func PatchBlueprint(id uint64, body map[string]interface{}) (*models.Blueprint, errors.Error) {
	// load record from db
	blueprint, err := GetBlueprint(id)
	if err != nil {
		return nil, err
	}

	originMode := blueprint.Mode
	err = helper.DecodeMapStruct(body, blueprint, true)
	if err != nil {
		return nil, err
	}
	// make sure mode is not being update
	if originMode != blueprint.Mode {
		return nil, errors.Default.New("mode is not updatable")
	}

	blueprint, err = saveBlueprint(blueprint)
	if err != nil {
		return nil, err
	}

	return blueprint, nil
}

// ReloadBlueprints FIXME ...
func ReloadBlueprints(c *cron.Cron) errors.Error {
	enable := true
	isManual := false
	blueprints, _, err := bpManager.GetDbBlueprints(&services.GetBlueprintQuery{
		Enable:   &enable,
		IsManual: &isManual,
	})
	if err != nil {
		return err
	}
	for _, e := range c.Entries() {
		c.Remove(e.ID)
	}
	c.Stop()
	for _, blueprint := range blueprints {
		if err != nil {
			blueprintLog.Error(err, failToCreateCronJob)
			return err
		}

		blueprintLog.Info("Add blueprint id:[%d] cronConfg[%s] to cron job", blueprint.ID, blueprint.CronConfig)
		blueprintJob := &BlueprintJob{
			Blueprint: blueprint,
		}

		if _, err := c.AddJob(blueprint.CronConfig, blueprintJob); err != nil {
			blueprintLog.Error(err, failToCreateCronJob)
			return errors.Default.Wrap(err, "created cron job failed")
		}
	}
	if len(blueprints) > 0 {
		c.Start()
	}
	logger.Info("total %d blueprints were scheduled", len(blueprints))
	return nil
}

func createPipelineByBlueprint(blueprint *models.Blueprint) (*models.Pipeline, errors.Error) {
	var plan plugin.PipelinePlan
	var err errors.Error
	if blueprint.Mode == models.BLUEPRINT_MODE_NORMAL {
		plan, err = MakePlanForBlueprint(blueprint)
	} else {
		plan, err = blueprint.UnmarshalPlan()
	}
	if err != nil {
		blueprintLog.Error(err, fmt.Sprintf("failed to MakePlanForBlueprint on blueprint:[%d][%s]", blueprint.ID, blueprint.Name))
		return nil, err
	}
	newPipeline := models.NewPipeline{}
	newPipeline.Plan = plan
	newPipeline.Name = blueprint.Name
	newPipeline.BlueprintId = blueprint.ID
	newPipeline.Labels = blueprint.Labels
	newPipeline.SkipOnFail = blueprint.SkipOnFail
	pipeline, err := CreatePipeline(&newPipeline)
	// Return all created tasks to the User
	if err != nil {
		blueprintLog.Error(err, fmt.Sprintf("%s on blueprint:[%d][%s]", failToCreateCronJob, blueprint.ID, blueprint.Name))
		return nil, errors.Convert(err)
	}
	return pipeline, nil
}

// MakePlanForBlueprint generates pipeline plan by version
func MakePlanForBlueprint(blueprint *models.Blueprint) (plugin.PipelinePlan, errors.Error) {
	bpSettings := new(models.BlueprintSettings)
	err := errors.Convert(json.Unmarshal(blueprint.Settings, bpSettings))
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("settings:%s", string(blueprint.Settings)))
	}

	bpSyncPolicy := plugin.BlueprintSyncPolicy{}
	bpSyncPolicy.TimeAfter = bpSettings.TimeAfter

	var plan plugin.PipelinePlan
	switch bpSettings.Version {
	case "1.0.0":
		// Notice: v1 not complete SkipOnFail & TimeAfter
		plan, err = GeneratePlanJsonV100(bpSettings)
	case "2.0.0":
		// load project metric plugins and convert it to a map
		metrics := make(map[string]json.RawMessage)
		projectMetrics := make([]models.ProjectMetricSetting, 0)
		if blueprint.ProjectName != "" {
			err = db.All(&projectMetrics, dal.Where("project_name = ? AND enable = ?", blueprint.ProjectName, true))
			if err != nil {
				return nil, err
			}
			for _, projectMetric := range projectMetrics {
				metrics[projectMetric.PluginName] = json.RawMessage(projectMetric.PluginOption)
			}
		}
		plan, err = GeneratePlanJsonV200(blueprint.ProjectName, bpSyncPolicy, bpSettings, metrics)
	default:
		return nil, errors.Default.New(fmt.Sprintf("unknown version of blueprint settings: %s", bpSettings.Version))
	}
	if err != nil {
		return nil, err
	}
	return WrapPipelinePlans(bpSettings.BeforePlan, plan, bpSettings.AfterPlan)
}

// WrapPipelinePlans merges multiple pipelines and append before and after pipeline
func WrapPipelinePlans(beforePlanJson json.RawMessage, mainPlan plugin.PipelinePlan, afterPlanJson json.RawMessage) (plugin.PipelinePlan, errors.Error) {
	beforePipelinePlan := plugin.PipelinePlan{}
	afterPipelinePlan := plugin.PipelinePlan{}

	if beforePlanJson != nil {
		err := errors.Convert(json.Unmarshal(beforePlanJson, &beforePipelinePlan))
		if err != nil {
			return nil, err
		}
	}
	if afterPlanJson != nil {
		err := errors.Convert(json.Unmarshal(afterPlanJson, &afterPipelinePlan))
		if err != nil {
			return nil, err
		}
	}

	return SequencializePipelinePlans(beforePipelinePlan, mainPlan, afterPipelinePlan), nil
}

// ParallelizePipelinePlans merges multiple pipelines into one unified plan
// by assuming they can be executed in parallel
func ParallelizePipelinePlans(plans ...plugin.PipelinePlan) plugin.PipelinePlan {
	merged := make(plugin.PipelinePlan, 0)
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

// SequencializePipelinePlans merges multiple pipelines into one unified plan
// by assuming they must be executed in sequencial order
func SequencializePipelinePlans(plans ...plugin.PipelinePlan) plugin.PipelinePlan {
	merged := make(plugin.PipelinePlan, 0)
	// iterate all pipelineTasks and try to merge them into `merged`
	for _, plan := range plans {
		merged = append(merged, plan...)
	}
	return merged
}

// TriggerBlueprint triggers blueprint immediately
func TriggerBlueprint(id uint64) (*models.Pipeline, errors.Error) {
	// load record from db
	blueprint, err := GetBlueprint(id)
	if err != nil {
		return nil, err
	}
	pipeline, err := createPipelineByBlueprint(blueprint)
	// done
	return pipeline, err
}
