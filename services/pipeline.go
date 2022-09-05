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
	"fmt"
	"github.com/apache/incubator-devlake/logger"
	"github.com/apache/incubator-devlake/models"
	"github.com/apache/incubator-devlake/utils"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"os"
	"path/filepath"
)

var notificationService *NotificationService
var temporalClient client.Client
var globalPipelineLog = logger.Global.Nested("pipeline service")

// PipelineQuery FIXME ...
type PipelineQuery struct {
	Status      string `form:"status"`
	Pending     int    `form:"pending"`
	Page        int    `form:"page"`
	PageSize    int    `form:"pageSize"`
	BlueprintId uint64 `form:"blueprint_id"`
}

func pipelineServiceInit() {
	pipelineServiceInitDO()
}

// CreatePipeline and return the model
func CreatePipeline(newPipeline *models.NewPipeline) (*models.ApiPipeline, error) {
	pipelineDO, err := CreatePipelineDO(newPipeline)
	if err != nil {
		return nil, err
	}
	pipelineDO, err = decryptPipelineDO(pipelineDO)
	if err != nil {
		return nil, err
	}
	pipeline, err := parsePipeline(pipelineDO)
	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

// GetPipelines by query
func GetPipelines(query *PipelineQuery) ([]*models.ApiPipeline, int64, error) {
	pipelineDOs, i, err := GetPipelineDOs(query)
	if err != nil {
		return nil, 0, err
	}
	pipelines := make([]*models.ApiPipeline, 0)
	for _, pipelineDO := range pipelineDOs {
		pipelineDO, err = decryptPipelineDO(pipelineDO)
		if err != nil {
			return nil, 0, err
		}
		pipelineDO, err := parsePipeline(pipelineDO)
		if err != nil {
			return nil, 0, err
		}
		pipelines = append(pipelines, pipelineDO)
	}

	return pipelines, i, nil
}

// GetPipeline by id
func GetPipeline(pipelineId uint64) (*models.ApiPipeline, error) {
	pipelineDO, err := GetPipelineDO(pipelineId)
	if err != nil {
		return nil, err
	}
	pipelineDO, err = decryptPipelineDO(pipelineDO)
	if err != nil {
		return nil, err
	}
	pipeline, err := parsePipeline(pipelineDO)
	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

// GetPipelineLogsArchivePath creates an archive for the logs of this pipeline and returns its file path
func GetPipelineLogsArchivePath(pipeline *models.ApiPipeline) (string, error) {
	logPath, err := getPipelineLogsPath(pipeline)
	if err != nil {
		return "", err
	}
	archive := fmt.Sprintf("%s/%s/logging.tar.gz", os.TempDir(), uuid.New())
	if err = utils.CreateGZipArchive(archive, fmt.Sprintf("%s/*", logPath)); err != nil {
		return "", err
	}
	return archive, err
}

// RunPipelineInQueue query pipeline from db and run it in a queue
func RunPipelineInQueue(pipelineMaxParallel int64) {
	RunPipelineInQueueDO(pipelineMaxParallel)
}

func watchTemporalPipelines() {
	watchTemporalPipelinesDO()
}

func getTemporalWorkflowId(pipelineId uint64) string {
	return fmt.Sprintf("pipeline #%d", pipelineId)
}

// NotifyExternal FIXME ...
func NotifyExternal(pipelineId uint64) error {
	if notificationService == nil {
		return nil
	}
	// send notification to an external web endpoint
	pipeline, err := GetPipeline(pipelineId)
	if err != nil {
		return err
	}
	err = notificationService.PipelineStatusChanged(PipelineNotification{
		PipelineID: pipeline.ID,
		CreatedAt:  pipeline.CreatedAt,
		UpdatedAt:  pipeline.UpdatedAt,
		BeganAt:    pipeline.BeganAt,
		FinishedAt: pipeline.FinishedAt,
		Status:     pipeline.Status,
	})
	if err != nil {
		globalPipelineLog.Error(err, "failed to send notification: %w", err)
		return err
	}
	return nil
}

// CancelPipeline FIXME ...
func CancelPipeline(pipelineId uint64) error {
	if temporalClient != nil {
		return temporalClient.CancelWorkflow(context.Background(), getTemporalWorkflowId(pipelineId), "")
	}
	pendingTasks, count, err := GetTasks(&TaskQuery{PipelineId: pipelineId, Pending: 1, PageSize: -1})
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	for _, pendingTask := range pendingTasks {
		_ = CancelTask(pendingTask.ID)
	}
	return err
}

// getPipelineLogsPath gets the logs directory of this pipeline
func getPipelineLogsPath(pipeline *models.ApiPipeline) (string, error) {
	pipelineLog := getPipelineLogger(pipeline)
	path := filepath.Dir(pipelineLog.GetConfig().Path)
	_, err := os.Stat(path)
	if err == nil {
		return path, nil
	}
	if os.IsNotExist(err) {
		return "", fmt.Errorf("logs for pipeline #%d not found: %v", pipeline.ID, err)
	}
	return "", fmt.Errorf("err validating logs path for pipeline #%d: %v", pipeline.ID, err)
}
