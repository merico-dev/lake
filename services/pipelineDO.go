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
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/models/common"
	"github.com/apache/incubator-devlake/plugins/core"
	"strings"
	"time"

	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/models"
	v11 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"golang.org/x/sync/semaphore"
	"gorm.io/gorm"
)

func pipelineServiceInitDO() {
	// notification
	var notificationEndpoint = cfg.GetString("NOTIFICATION_ENDPOINT")
	var notificationSecret = cfg.GetString("NOTIFICATION_SECRET")
	if strings.TrimSpace(notificationEndpoint) != "" {
		notificationService = NewNotificationService(notificationEndpoint, notificationSecret)
	}

	// temporal client
	var temporalUrl = cfg.GetString("TEMPORAL_URL")
	if temporalUrl != "" {
		// TODO: logger
		var err error
		temporalClient, err = client.NewClient(client.Options{
			HostPort: temporalUrl,
		})
		if err != nil {
			panic(err)
		}
		watchTemporalPipelines()
	} else {
		// standalone mode: reset pipeline status
		db.Model(&models.PipelineDO{}).Where("status <> ?", models.TASK_COMPLETED).Update("status", models.TASK_FAILED)
		db.Model(&models.Task{}).Where("status <> ?", models.TASK_COMPLETED).Update("status", models.TASK_FAILED)
	}

	err := ReloadBlueprints(cronManager)
	if err != nil {
		panic(err)
	}

	var pipelineMaxParallel = cfg.GetInt64("PIPELINE_MAX_PARALLEL")
	if pipelineMaxParallel < 0 {
		panic(errors.BadInput.New(`PIPELINE_MAX_PARALLEL should be a positive integer`, errors.AsUserMessage()))
	}
	if pipelineMaxParallel == 0 {
		globalPipelineLog.Warn(nil, `pipelineMaxParallel=0 means pipeline will be run No Limit`)
		pipelineMaxParallel = 10000
	}
	// run pipeline with independent goroutine
	go RunPipelineInQueue(pipelineMaxParallel)
}

// CreatePipeline and return the model
func CreatePipelineDO(newPipeline *models.NewPipeline) (*models.PipelineDO, error) {
	// create pipeline object from posted data
	pipelineDO := &models.PipelineDO{
		Name:          newPipeline.Name,
		FinishedTasks: 0,
		Status:        models.TASK_CREATED,
		Message:       "",
		SpentSeconds:  0,
	}
	if newPipeline.BlueprintId != 0 {
		pipelineDO.BlueprintId = newPipeline.BlueprintId
	}
	pipelineDO, err := encryptPipelineDO(pipelineDO)
	if err != nil {
		return nil, err
	}
	// save pipeline to database
	err = db.Create(&pipelineDO).Error
	if err != nil {
		globalPipelineLog.Error(err, "create pipline failed: %w", err)
		return nil, errors.Internal.Wrap(err, "create pipline failed")
	}

	// create tasks accordingly
	for i := range newPipeline.Plan {
		for j := range newPipeline.Plan[i] {
			pipelineTask := newPipeline.Plan[i][j]
			newTask := &models.NewTask{
				PipelineTask: pipelineTask,
				PipelineId:   pipelineDO.ID,
				PipelineRow:  i + 1,
				PipelineCol:  j + 1,
			}
			_, err := CreateTask(newTask)
			if err != nil {
				globalPipelineLog.Error(err, "create task for pipeline failed: %w", err)
				return nil, err
			}
			// sync task state back to pipeline
			pipelineDO.TotalTasks += 1
		}
	}
	if err != nil {
		globalPipelineLog.Error(err, "save tasks for pipeline failed: %w", err)
		return nil, errors.Internal.Wrap(err, "save tasks for pipeline failed")
	}
	if pipelineDO.TotalTasks == 0 {
		return nil, fmt.Errorf("no task to run")
	}

	// update tasks state
	planByte, err := json.Marshal(newPipeline.Plan)
	if err != nil {
		return nil, err
	}
	pipelineDO.Plan = string(planByte)
	pipelineDO, err = encryptPipelineDO(pipelineDO)
	if err != nil {
		return nil, err
	}
	err = db.Model(pipelineDO).Updates(map[string]interface{}{
		"total_tasks": pipelineDO.TotalTasks,
		"plan":        pipelineDO.Plan,
	}).Error
	if err != nil {
		globalPipelineLog.Error(err, "update pipline state failed: %w", err)
		return nil, errors.Internal.Wrap(err, "update pipline state failed")
	}

	return pipelineDO, nil
}

// GetPipelines by query
func GetPipelineDOs(query *PipelineQuery) ([]*models.PipelineDO, int64, error) {
	pipelineDOs := make([]*models.PipelineDO, 0)
	db := db.Model(pipelineDOs).Order("id DESC")
	if query.BlueprintId != 0 {
		db = db.Where("blueprint_id = ?", query.BlueprintId)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Pending > 0 {
		db = db.Where("finished_at is null")
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
	err = db.Find(&pipelineDOs).Error
	if err != nil {
		return nil, count, err
	}
	return pipelineDOs, count, nil
}

// GetPipeline by id
func GetPipelineDO(pipelineId uint64) (*models.PipelineDO, error) {
	pipelineDO := &models.PipelineDO{}
	err := db.First(pipelineDO, pipelineId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound.New("pipeline not found", errors.AsUserMessage())
		}
		return nil, errors.Internal.Wrap(err, "error getting the pipeline from database", errors.AsUserMessage())
	}
	return pipelineDO, nil
}

// RunPipelineInQueue query pipeline from db and run it in a queue
func RunPipelineInQueueDO(pipelineMaxParallel int64) {
	sema := semaphore.NewWeighted(pipelineMaxParallel)
	startedPipelineIds := []uint64{}
	for {
		globalPipelineLog.Info("wait for new pipeline")
		// start goroutine when sema lock ready and pipeline exist.
		// to avoid read old pipeline, acquire lock before read exist pipeline
		err := sema.Acquire(context.TODO(), 1)
		if err != nil {
			panic(err)
		}
		globalPipelineLog.Info("get lock and wait pipeline")
		pipelineDO := &models.PipelineDO{}
		for {
			db.Where("status = ?", models.TASK_CREATED).
				Not(startedPipelineIds).
				Order("id ASC").Limit(1).Find(pipelineDO)
			if pipelineDO.ID != 0 {
				break
			}
			time.Sleep(time.Second)
		}
		startedPipelineIds = append(startedPipelineIds, pipelineDO.ID)
		go func() {
			defer sema.Release(1)
			globalPipelineLog.Info("run pipeline, %d", pipelineDO.ID)
			err = runPipeline(pipelineDO.ID)
			if err != nil {
				globalPipelineLog.Error(err, "failed to run pipeline %d", pipelineDO.ID)
			}
		}()
	}
}

func watchTemporalPipelinesDO() {
	ticker := time.NewTicker(3 * time.Second)
	dc := converter.GetDefaultDataConverter()
	go func() {
		// run forever
		for range ticker.C {
			// load all running pipeline from database
			runningPipelineDOs := make([]models.PipelineDO, 0)
			err := db.Find(&runningPipelineDOs, "status = ?", models.TASK_RUNNING).Error
			if err != nil {
				panic(err)
			}
			progressDetails := make(map[uint64]*models.TaskProgressDetail)
			// check their status against temporal
			for _, rp := range runningPipelineDOs {
				workflowId := getTemporalWorkflowId(rp.ID)
				desc, err := temporalClient.DescribeWorkflowExecution(
					context.Background(),
					workflowId,
					"",
				)
				if err != nil {
					globalPipelineLog.Error(err, "failed to query workflow execution: %w", err)
					continue
				}
				// workflow is terminated by outsider
				s := desc.WorkflowExecutionInfo.Status
				if s != v11.WORKFLOW_EXECUTION_STATUS_RUNNING {
					rp.Status = models.TASK_COMPLETED
					if s != v11.WORKFLOW_EXECUTION_STATUS_COMPLETED {
						rp.Status = models.TASK_FAILED
						// get error message
						hisIter := temporalClient.GetWorkflowHistory(
							context.Background(),
							workflowId,
							"",
							false,
							v11.HISTORY_EVENT_FILTER_TYPE_CLOSE_EVENT,
						)
						for hisIter.HasNext() {
							his, err := hisIter.Next()
							if err != nil {
								globalPipelineLog.Error(err, "failed to get next from workflow history iterator: %w", err)
								continue
							}
							rp.Message = fmt.Sprintf("temporal event type: %v", his.GetEventType())
						}
					}
					rp.FinishedAt = desc.WorkflowExecutionInfo.CloseTime
					err = db.Model(rp).Updates(map[string]interface{}{
						"status":      rp.Status,
						"message":     rp.Message,
						"finished_at": rp.FinishedAt,
					}).Error
					if err != nil {
						globalPipelineLog.Error(err, "failed to update db: %w", err)
					}
					continue
				}

				// check pending activity
				for _, activity := range desc.PendingActivities {
					taskId, err := getTaskIdFromActivityId(activity.ActivityId)
					if err != nil {
						globalPipelineLog.Error(err, "unable to extract task id from activity id `%s`", activity.ActivityId)
						continue
					}
					progressDetail := &models.TaskProgressDetail{}
					progressDetails[taskId] = progressDetail
					heartbeats := activity.GetHeartbeatDetails()
					if heartbeats == nil {
						continue
					}
					payloads := heartbeats.GetPayloads()
					if len(payloads) == 0 {
						return
					}
					lastPayload := payloads[len(payloads)-1]
					err = dc.FromPayload(lastPayload, progressDetail)
					if err != nil {
						globalPipelineLog.Error(err, "failed to unmarshal heartbeat payload: %w", err)
						continue
					}
				}
			}
			runningTasks.setAll(progressDetails)
		}
	}()
}

// parsePipeline
func parsePipeline(pipelineDO *models.PipelineDO) (*models.Pipeline, error) {
	pipeline := models.Pipeline{
		Model: common.Model{
			ID: pipelineDO.ID,
		},
		Name:          pipelineDO.Name,
		BlueprintId:   pipelineDO.BlueprintId,
		Plan:          []byte(pipelineDO.Plan),
		TotalTasks:    pipelineDO.TotalTasks,
		FinishedTasks: pipelineDO.FinishedTasks,
		BeganAt:       pipelineDO.BeganAt,
		FinishedAt:    pipelineDO.FinishedAt,
		Status:        pipelineDO.Status,
		Message:       pipelineDO.Message,
		SpentSeconds:  pipelineDO.SpentSeconds,
		Stage:         pipelineDO.Stage,
	}
	return &pipeline, nil
}

// parsePipelineDO
func parsePipelineDO(pipeline *models.Pipeline) (*models.PipelineDO, error) {
	pipelineDO := models.PipelineDO{
		Model: common.Model{
			ID: pipeline.ID,
		},
		Name:          pipeline.Name,
		BlueprintId:   pipeline.BlueprintId,
		Plan:          string(pipeline.Plan),
		TotalTasks:    pipeline.TotalTasks,
		FinishedTasks: pipeline.FinishedTasks,
		BeganAt:       pipeline.BeganAt,
		FinishedAt:    pipeline.FinishedAt,
		Status:        pipeline.Status,
		Message:       pipeline.Message,
		SpentSeconds:  pipeline.SpentSeconds,
		Stage:         pipeline.Stage,
	}
	return &pipelineDO, nil
}

// encryptPipelineDO
func encryptPipelineDO(pipelineDO *models.PipelineDO) (*models.PipelineDO, error) {
	encKey := config.GetConfig().GetString(core.EncodeKeyEnvStr)
	planEncrypt, err := core.Encrypt(encKey, pipelineDO.Plan)
	if err != nil {
		return nil, err
	}
	pipelineDO.Plan = planEncrypt
	return pipelineDO, nil
}

// encryptPipelineDO
func decryptPipelineDO(pipelineDO *models.PipelineDO) (*models.PipelineDO, error) {
	encKey := config.GetConfig().GetString(core.EncodeKeyEnvStr)
	plan, err := core.Decrypt(encKey, pipelineDO.Plan)
	if err != nil {
		return nil, err
	}
	pipelineDO.Plan = plan
	return pipelineDO, nil
}
