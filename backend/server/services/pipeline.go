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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/core/utils"
	"github.com/apache/incubator-devlake/helpers/dbhelper"
	"github.com/apache/incubator-devlake/impls/logruslog"
	"github.com/google/uuid"
	v11 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"golang.org/x/sync/semaphore"
)

var notificationService *NotificationService
var temporalClient client.Client
var globalPipelineLog = logruslog.Global.Nested("pipeline service")

// PipelineQuery is a query for GetPipelines
type PipelineQuery struct {
	Pagination
	Status      string `form:"status"`
	Pending     int    `form:"pending"`
	BlueprintId uint64 `uri:"blueprintId" form:"blueprint_id"`
	Label       string `form:"label"`
}

func pipelineServiceInit() {
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
		errMsg := "The process was terminated unexpectedly"
		err := db.UpdateColumns(
			&models.Pipeline{},
			[]dal.DalSet{
				{ColumnName: "status", Value: models.TASK_FAILED},
				{ColumnName: "message", Value: errMsg},
			},
			dal.Where("status = ?", models.TASK_RUNNING),
		)
		if err != nil {
			panic(err)
		}
		err = db.UpdateColumns(
			&models.Task{},
			[]dal.DalSet{
				{ColumnName: "status", Value: models.TASK_FAILED},
				{ColumnName: "message", Value: errMsg},
			},
			dal.Where("status = ?", models.TASK_RUNNING),
		)
		if err != nil {
			panic(err)
		}
	}

	err := ReloadBlueprints(cronManager)
	if err != nil {
		panic(err)
	}

	var pipelineMaxParallel = cfg.GetInt64("PIPELINE_MAX_PARALLEL")
	if pipelineMaxParallel < 0 {
		panic(errors.BadInput.New(`PIPELINE_MAX_PARALLEL should be a positive integer`))
	}
	if pipelineMaxParallel == 0 {
		globalPipelineLog.Warn(nil, `pipelineMaxParallel=0 means pipeline will be run No Limit`)
		pipelineMaxParallel = 10000
	}
	// run pipeline with independent goroutine
	//go RunPipelineInQueue(pipelineMaxParallel)
}

// CreatePipeline and return the model
func CreatePipeline(newPipeline *models.NewPipeline) (*models.Pipeline, errors.Error) {
	pipeline, err := CreateDbPipeline(newPipeline)
	if err != nil {
		return nil, errors.Convert(err)
	}
	return pipeline, nil
}

// GetPipelines by query
func GetPipelines(query *PipelineQuery) ([]*models.Pipeline, int64, errors.Error) {
	pipelines, i, err := GetDbPipelines(query)
	if err != nil {
		return nil, 0, errors.Convert(err)
	}
	for _, p := range pipelines {
		err = fillPipelineDetail(p)
		if err != nil {
			return nil, 0, err
		}
	}
	return pipelines, i, nil
}

// GetPipeline by id
func GetPipeline(pipelineId uint64) (*models.Pipeline, errors.Error) {
	dbPipeline, err := GetDbPipeline(pipelineId)
	if err != nil {
		return nil, err
	}
	err = fillPipelineDetail(dbPipeline)
	if err != nil {
		return nil, err
	}
	return dbPipeline, nil
}

// GetPipelineLogsArchivePath creates an archive for the logs of this pipeline and returns its file path
func GetPipelineLogsArchivePath(pipeline *models.Pipeline) (string, errors.Error) {
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

func dequeuePipeline(runningParallelLabels []string) (pipeline *models.Pipeline, err errors.Error) {
	txHelper := dbhelper.NewTxHelper(basicRes, &err)
	defer txHelper.End()
	tx := txHelper.Begin()
	// mysql read lock, not sure if it works for postgresql
	errors.Must(tx.LockTables(dal.LockTables{
		{Table: "_devlake_pipelines", Exclusive: false},
		{Table: "_devlake_pipeline_labels", Exclusive: false},
	}))
	// prepare query to find an appropriate pipeline to execute
	pipeline = &models.Pipeline{}
	err = tx.First(pipeline,
		dal.Where("status IN ?", []string{models.TASK_CREATED, models.TASK_RERUN}),
		dal.Join(
			`left join _devlake_pipeline_labels ON
				_devlake_pipeline_labels.pipeline_id = _devlake_pipelines.id AND
				_devlake_pipeline_labels.name LIKE 'parallel/%' AND
				_devlake_pipeline_labels.name in ?`,
			runningParallelLabels,
		),
		dal.Groupby("id"),
		dal.Having("count(_devlake_pipeline_labels.name)=0"),
		dal.Select("id"),
		dal.Orderby("id ASC"),
		dal.Limit(1),
	)
	if err == nil {
		// mark the pipeline running, now we want a write lock
		errors.Must(tx.LockTables(dal.LockTables{{Table: "_devlake_pipelines", Exclusive: true}}))
		err = tx.UpdateColumns(&models.Pipeline{}, []dal.DalSet{
			{ColumnName: "status", Value: models.TASK_RUNNING},
			{ColumnName: "message", Value: ""},
			{ColumnName: "began_at", Value: time.Now()},
		}, dal.Where("id = ?", pipeline.ID))
		if err != nil {
			panic(err)
		}
		return
	}
	if tx.IsErrorNotFound(err) {
		pipeline = nil
		err = nil
	} else {
		// log unexpected err
		globalPipelineLog.Error(err, "dequeue failed")
	}

	return
}

// RunPipelineInQueue query pipeline from db and run it in a queue
func RunPipelineInQueue(pipelineMaxParallel int64) {
	sema := semaphore.NewWeighted(pipelineMaxParallel)
	runningParallelLabels := []string{}
	var runningParallelLabelLock sync.Mutex
	var err error
	for {
		// start goroutine when sema lock ready and pipeline exist.
		// to avoid read old pipeline, acquire lock before read exist pipeline
		errors.Must(sema.Acquire(context.TODO(), 1))
		globalPipelineLog.Info("get lock and wait next pipeline")
		var dbPipeline *models.Pipeline
		for {
			dbPipeline, err = dequeuePipeline(runningParallelLabels)
			if err == nil && dbPipeline != nil {
				break
			}
			time.Sleep(time.Second)
		}

		err = fillPipelineDetail(dbPipeline)
		if err != nil {
			panic(err)
		}
		// add pipelineParallelLabels to runningParallelLabels
		var pipelineParallelLabels []string
		for _, dbLabel := range dbPipeline.Labels {
			if strings.HasPrefix(dbLabel, `parallel/`) {
				pipelineParallelLabels = append(pipelineParallelLabels, dbLabel)
			}
		}
		runningParallelLabelLock.Lock()
		runningParallelLabels = append(runningParallelLabels, pipelineParallelLabels...)
		runningParallelLabelLock.Unlock()

		go func(pipelineId uint64, parallelLabels []string) {
			defer sema.Release(1)
			defer func() {
				runningParallelLabelLock.Lock()
				runningParallelLabels = utils.SliceRemove(runningParallelLabels, parallelLabels...)
				runningParallelLabelLock.Unlock()
				globalPipelineLog.Info("finish pipeline #%d, now runningParallelLabels is %s", pipelineId, runningParallelLabels)
			}()
			globalPipelineLog.Info("run pipeline, %d, now running runningParallelLabels are %s", pipelineId, runningParallelLabels)
			err = runPipeline(pipelineId)
			if err != nil {
				globalPipelineLog.Error(err, "failed to run pipeline %d", pipelineId)
			}
		}(dbPipeline.ID, pipelineParallelLabels)
	}
}

func watchTemporalPipelines() {
	ticker := time.NewTicker(3 * time.Second)
	dc := converter.GetDefaultDataConverter()
	go func() {
		// run forever
		for range ticker.C {
			// load all running pipeline from database
			runningDbPipelines := make([]models.Pipeline, 0)
			err := db.All(&runningDbPipelines, dal.Where("status = ?", models.TASK_RUNNING))
			if err != nil {
				panic(err)
			}
			// progressDetails will be only used in this goroutine now
			// So it needn't lock and unlock now
			progressDetails := make(map[uint64]*models.TaskProgressDetail)
			// check their status against temporal
			for _, rp := range runningDbPipelines {
				workflowId := getTemporalWorkflowId(rp.ID)
				desc, err := temporalClient.DescribeWorkflowExecution(
					context.Background(),
					workflowId,
					"",
				)
				if err != nil {
					globalPipelineLog.Error(err, "failed to query workflow execution: %v", err)
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
								globalPipelineLog.Error(err, "failed to get next from workflow history iterator: %v", err)
								continue
							}
							rp.Message = fmt.Sprintf("temporal event type: %v", his.GetEventType())
						}
					}
					rp.FinishedAt = desc.WorkflowExecutionInfo.CloseTime
					err = db.UpdateColumns(rp, []dal.DalSet{
						{ColumnName: "status", Value: rp.Status},
						{ColumnName: "message", Value: rp.Message},
						{ColumnName: "finished_at", Value: rp.FinishedAt},
					})
					if err != nil {
						globalPipelineLog.Error(err, "failed to update db: %v", err)
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
					if err := dc.FromPayload(lastPayload, progressDetail); err != nil {
						globalPipelineLog.Error(err, "failed to unmarshal heartbeat payload: %v", err)
						continue
					}
				}
			}
			runningTasks.setAll(progressDetails)
		}
	}()
}

func getTemporalWorkflowId(pipelineId uint64) string {
	return fmt.Sprintf("pipeline #%d", pipelineId)
}

// NotifyExternal FIXME ...
func NotifyExternal(pipelineId uint64) errors.Error {
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
		globalPipelineLog.Error(err, "failed to send notification: %v", err)
		return err
	}
	return nil
}

// CancelPipeline FIXME ...
func CancelPipeline(pipelineId uint64) errors.Error {
	// prevent RunPipelineInQueue from consuming pending pipelines
	pipeline := &models.Pipeline{}
	err := db.First(pipeline, dal.Where("id = ?", pipelineId))
	if err != nil {
		return errors.BadInput.New("pipeline not found")
	}
	if pipeline.Status == models.TASK_CREATED || pipeline.Status == models.TASK_RERUN {
		pipeline.Status = models.TASK_CANCELLED
		err = db.Update(pipeline)
		if err != nil {
			return errors.Default.Wrap(err, "faile to update pipeline")
		}
		// now, with RunPipelineInQueue being block and target pipeline got updated
		// we should update the related tasks as well
		err = db.UpdateColumn(
			&models.Task{},
			"status", models.TASK_CANCELLED,
			dal.Where("pipeline_id = ?", pipelineId),
		)
		if err != nil {
			return errors.Default.Wrap(err, "faile to update pipeline tasks")
		}
		// the target pipeline is pending, no running, no need to perform the actual cancel operation
		return nil
	}
	if temporalClient != nil {
		return errors.Convert(temporalClient.CancelWorkflow(context.Background(), getTemporalWorkflowId(pipelineId), ""))
	}
	pendingTasks, count, err := GetTasks(&TaskQuery{PipelineId: pipelineId, Pending: 1, Pagination: Pagination{PageSize: -1}})
	if err != nil {
		return errors.Convert(err)
	}
	if count == 0 {
		return nil
	}
	for _, pendingTask := range pendingTasks {
		_ = CancelTask(pendingTask.ID)
	}
	return errors.Convert(err)
}

// getPipelineLogsPath gets the logs directory of this pipeline
func getPipelineLogsPath(pipeline *models.Pipeline) (string, errors.Error) {
	pipelineLog := GetPipelineLogger(pipeline)
	path := filepath.Dir(pipelineLog.GetConfig().Path)
	_, err := os.Stat(path)
	if err == nil {
		return path, nil
	}
	if os.IsNotExist(err) {
		return "", errors.NotFound.Wrap(err, fmt.Sprintf("logs for pipeline #%d not found", pipeline.ID))
	}
	return "", errors.Default.Wrap(err, fmt.Sprintf("error validating logs path for pipeline #%d", pipeline.ID))
}

// RerunPipeline would rerun all failed tasks or specified task
func RerunPipeline(pipelineId uint64, task *models.Task) (tasks []*models.Task, err errors.Error) {
	// prevent pipeline executor from doing anything that might jeopardize the integrity
	pipeline := &models.Pipeline{}
	txHelper := dbhelper.NewTxHelper(basicRes, &err)
	tx := txHelper.Begin()
	defer txHelper.End()
	err = txHelper.LockTablesTimeout(2*time.Second, dal.LockTables{{Table: "_devlake_pipelines", Exclusive: true}})
	if err != nil {
		err = errors.BadInput.Wrap(err, "failed to lock pipeline table, is there any pending pipeline or deletion?")
		return
	}

	// load the pipeline
	err = tx.First(pipeline, dal.Where("id = ?", pipelineId))
	if err != nil {
		return nil, err
	}

	// verify the status
	if pipeline.Status == models.TASK_RUNNING {
		return nil, errors.BadInput.New("pipeline is running")
	}
	if pipeline.Status == models.TASK_CREATED || pipeline.Status == models.TASK_RERUN {
		return nil, errors.BadInput.New("pipeline is waiting to run")
	}

	// determine which tasks to rerun
	var failedTasks []*models.Task
	if task != nil {
		if task.PipelineId != pipelineId {
			return nil, errors.BadInput.New("the task ID and pipeline ID doesn't match")
		}
		failedTasks = append(failedTasks, task)
	} else {
		tasks, err := GetTasksWithLastStatus(pipelineId)
		if err != nil {
			return nil, errors.Default.Wrap(err, "error getting tasks")
		}
		for _, t := range tasks {
			if t.Status != models.TASK_COMPLETED {
				failedTasks = append(failedTasks, t)
			}
		}
	}

	// no tasks to rerun
	if len(failedTasks) == 0 {
		return nil, errors.BadInput.New("no tasks to be re-ran")
	}

	// create new tasks
	// TODO: this is better to be wrapped inside a transaction
	rerunTasks := []*models.Task{}
	for _, t := range failedTasks {
		// mark previous task failed
		t.Status = models.TASK_FAILED
		err := tx.UpdateColumn(t, "status", models.TASK_FAILED)
		if err != nil {
			return nil, err
		}
		// create new task
		subtasks, err := t.GetSubTasks()
		if err != nil {
			return nil, err
		}
		options, err := t.GetOptions()
		if err != nil {
			return nil, err
		}
		rerunTask, err := CreateTask(&models.NewTask{
			PipelineTask: &plugin.PipelineTask{
				Plugin:   t.Plugin,
				Subtasks: subtasks,
				Options:  options,
			},
			PipelineId:  t.PipelineId,
			PipelineRow: t.PipelineRow,
			PipelineCol: t.PipelineCol,
			IsRerun:     true,
		})
		if err != nil {
			return nil, err
		}
		// append to result
		rerunTasks = append(rerunTasks, rerunTask)
	}

	// mark pipline rerun
	err = tx.UpdateColumn(&models.Pipeline{},
		"status", models.TASK_RERUN,
		dal.Where("id = ?", pipelineId),
	)
	if err != nil {
		return nil, err
	}
	return rerunTasks, nil
}
