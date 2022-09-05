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

package pipelines

import (
	"github.com/apache/incubator-devlake/errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/apache/incubator-devlake/api/shared"
	"github.com/apache/incubator-devlake/models"
	"github.com/apache/incubator-devlake/services"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

/*
Create and run a new pipeline
POST /pipelines
{
	"name": "name-of-pipeline",
	"tasks": [
		[ {"plugin": "gitlab", ...}, {"plugin": "jira"} ],
		[ {"plugin": "github", ...}],
	]
}
*/
// @Summary Create and run a new pipeline
// @Description Create and run a new pipeline
// @Tags framework/pipelines
// @Accept application/json
// @Param pipeline body models.NewPipeline true "json"
// @Success 200  {object} models.ApiPipeline
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /pipelines [post]
func Post(c *gin.Context) {
	newPipeline := &models.NewPipeline{}

	err := c.MustBindWith(newPipeline, binding.JSON)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad JSON request body format", errors.AsUserMessage()))
		return
	}

	pipeline, err := services.CreatePipeline(newPipeline)
	// Return all created tasks to the User
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error creating pipeline", errors.AsUserMessage()))
		return
	}
	shared.ApiOutputSuccess(c, pipeline, http.StatusCreated)
}

/*
Get list of pipelines
GET /pipelines?status=TASK_RUNNING&pending=1&page=1&pagesize=10
{
	"pipelines": [
		{"id": 1, "name": "test-pipeline", ...}
	],
	"count": 5
}
*/

// @Summary Get list of pipelines
// @Description GET /pipelines?status=TASK_RUNNING&pending=1&page=1&pagesize=10
// @Tags framework/pipelines
// @Param status query string true "query"
// @Param pending query int true "query"
// @Param page query int true "query"
// @Param pagesize query int true "query"
// @Success 200  {object} shared.ResponsePipelines
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /pipelines [get]
func Index(c *gin.Context) {
	var query services.PipelineQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad request body format", errors.AsUserMessage()))
		return
	}
	pipelines, count, err := services.GetPipelines(&query)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error getting pipelines", errors.AsUserMessage()))
		return
	}
	shared.ApiOutputSuccess(c, shared.ResponsePipelines{Pipelines: pipelines, Count: count}, http.StatusOK)
}

/*
Get detail of a pipeline
GET /pipelines/:pipelineId
{
	"id": 1,
	"name": "test-pipeline",
	...
}
*/
// @Get detail of a pipeline
// @Description GET /pipelines/:pipelineId
// @Description RETURN SAMPLE
// @Description {
// @Description 	"id": 1,
// @Description 	"name": "test-pipeline",
// @Description 	...
// @Description }
// @Tags framework/pipelines
// @Param pipelineId path int true "query"
// @Success 200  {object} models.ApiPipeline
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /pipelines/{pipelineId} [get]
func Get(c *gin.Context) {
	pipelineId := c.Param("pipelineId")
	id, err := strconv.ParseUint(pipelineId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad pipelineID format supplied", errors.AsUserMessage()))
		return
	}
	pipeline, err := services.GetPipeline(id)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error getting pipeline", errors.AsUserMessage()))
		return
	}
	shared.ApiOutputSuccess(c, pipeline, http.StatusOK)
}

/*
Cancel a pending pipeline
DELETE /pipelines/:pipelineId
*/
// @Cancel a pending pipeline
// @Description Cancel a pending pipeline
// @Tags framework/pipelines
// @Param pipelineId path int true "pipeline ID"
// @Success 200
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /pipelines/{pipelineId} [delete]
func Delete(c *gin.Context) {
	pipelineId := c.Param("pipelineId")
	id, err := strconv.ParseUint(pipelineId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad pipelineID format supplied", errors.AsUserMessage()))
		return
	}
	err = services.CancelPipeline(id)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error cancelling pipeline", errors.AsUserMessage()))
		return
	}
	shared.ApiOutputSuccess(c, nil, http.StatusOK)
}

/*
Get download logs of a pipeline
GET /pipelines/:pipelineId/logging.tar.gz
*/
// download logs of a pipeline
// @Description GET /pipelines/:pipelineId/logging.tar.gz
// @Tags framework/pipelines
// @Param pipelineId path int true "query"
// @Success 200  "The archive file"
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 404  {string} errcode.Error "ApiPipeline or Log files not found"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /pipelines/{pipelineId}/logging.tar.gz [get]
func DownloadLogs(c *gin.Context) {
	pipelineId := c.Param("pipelineId")
	id, err := strconv.ParseUint(pipelineId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad pipeline ID format supplied", errors.AsUserMessage()))
		return
	}
	pipeline, err := services.GetPipeline(id)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error getting pipeline", errors.AsUserMessage()))
		return
	}
	archive, err := services.GetPipelineLogsArchivePath(pipeline)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error getting logs for pipeline", errors.AsUserMessage()))
		return
	}
	defer os.Remove(archive)
	c.FileAttachment(archive, filepath.Base(archive))
}
