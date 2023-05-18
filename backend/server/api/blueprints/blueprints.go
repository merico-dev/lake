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

package blueprints

import (
	"net/http"
	"strconv"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/server/api/shared"
	"github.com/apache/incubator-devlake/server/services"

	"github.com/gin-gonic/gin"
)

type PaginatedBlueprint struct {
	Blueprints []*models.Blueprint `json:"blueprints"`
	Count      int64               `json:"count"`
}

// @Summary post blueprints
// @Description post blueprints
// @Tags framework/blueprints
// @Accept application/json
// @Param blueprint body models.Blueprint true "json"
// @Success 200  {object} models.Blueprint
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router /blueprints [post]
func Post(c *gin.Context) {
	blueprint := &models.Blueprint{}

	err := c.ShouldBind(blueprint)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, shared.BadRequestBody))
		return
	}

	err = services.CreateBlueprint(blueprint)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error creating blueprint"))
		return
	}

	shared.ApiOutputSuccess(c, blueprint, http.StatusCreated)
}

// @Summary get blueprints
// @Description get blueprints
// @Tags framework/blueprints
// @Param enable query bool false "enable"
// @Param isManual query bool false "isManual"
// @Param page query int false "page"
// @Param pageSize query int false "pageSize"
// @Param label query string false "label"
// @Success 200  {object} PaginatedBlueprint
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router /blueprints [get]
func Index(c *gin.Context) {
	var query services.BlueprintQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, shared.BadRequestBody))
		return
	}
	blueprints, count, err := services.GetBlueprints(&query)
	if err != nil {
		shared.ApiOutputAbort(c, errors.Default.Wrap(err, "error getting blueprints"))
		return
	}
	shared.ApiOutputSuccess(c, PaginatedBlueprint{Blueprints: blueprints, Count: count}, http.StatusOK)
}

// @Summary get blueprints
// @Description get blueprints
// @Tags framework/blueprints
// @Accept application/json
// @Param blueprintId path int true "blueprint id"
// @Success 200  {object} models.Blueprint
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router /blueprints/{blueprintId} [get]
func Get(c *gin.Context) {
	blueprintId := c.Param("blueprintId")
	id, err := strconv.ParseUint(blueprintId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad blueprintId format supplied"))
		return
	}
	blueprint, err := services.GetBlueprint(id)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error getting blueprint"))
		return
	}
	shared.ApiOutputSuccess(c, blueprint, http.StatusOK)
}

// // @Summary delete blueprints
// // @Description Delete BluePrints
// // @Tags framework/blueprints
// // @Param blueprintId path string true "blueprintId"
// // @Success 200
// // @Failure 400  {object} shared.ApiBody "Bad Request"
// // @Failure 500  {object} shared.ApiBody "Internal Error"
// // @Router /blueprints/{blueprintId} [delete]
// func Delete(c *gin.Context) {
// 	pipelineId := c.Param("blueprintId")
// 	id, err := strconv.ParseUint(pipelineId, 10, 64)
// 	if err != nil {
// 		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad blueprintID format supplied"))
// 		return
// 	}
// 	err = services.DeleteBlueprint(id)
// 	if err != nil {
// 		shared.ApiOutputError(c, errors.Default.Wrap(err, "error deleting blueprint"))
// 	}
// }

// @Summary patch blueprints
// @Description patch blueprints
// @Tags framework/blueprints
// @Accept application/json
// @Param blueprintId path string true "blueprintId"
// @Success 200  {object} models.Blueprint
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router /blueprints/{blueprintId} [Patch]
func Patch(c *gin.Context) {
	blueprintId := c.Param("blueprintId")
	id, err := strconv.ParseUint(blueprintId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad pipeline ID format supplied"))
		return
	}
	var body map[string]interface{}
	err = c.ShouldBind(&body)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, shared.BadRequestBody))
		return
	}
	blueprint, err := services.PatchBlueprint(id, body)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error patching the blueprint"))
		return
	}
	shared.ApiOutputSuccess(c, blueprint, http.StatusOK)
}

// @Summary trigger blueprint
// @Description trigger a blueprint immediately
// @Tags framework/blueprints
// @Accept application/json
// @Param blueprintId path string true "blueprintId"
// @Success 200  {object} models.Pipeline
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router /blueprints/{blueprintId}/trigger [Post]
func Trigger(c *gin.Context) {
	blueprintId := c.Param("blueprintId")
	id, err := strconv.ParseUint(blueprintId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad blueprintID format supplied"))
		return
	}
	pipeline, err := services.TriggerBlueprint(id)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error triggering blueprint"))
		return
	}
	shared.ApiOutputSuccess(c, pipeline, http.StatusOK)
}

// @Summary get pipelines by blueprint id
// @Description get pipelines by blueprint id
// @Tags framework/blueprints
// @Accept application/json
// @Param blueprintId path int true "blueprint id"
// @Success 200  {object} shared.ResponsePipelines
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router /blueprints/{blueprintId}/pipelines [get]
func GetBlueprintPipelines(c *gin.Context) {
	var query services.PipelineQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, shared.BadRequestBody))
		return
	}
	err = c.ShouldBindUri(&query)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad request URI format"))
		return
	}

	pipelines, count, err := services.GetPipelines(&query)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error getting pipelines"))
		return
	}
	shared.ApiOutputSuccess(c, shared.ResponsePipelines{Pipelines: pipelines, Count: count}, http.StatusOK)
}

// @Summary delete blueprint by id
// @Description delete blueprint by id
// @Tags framework/blueprints
// @Accept application/json
// @Param blueprintId path int true "blueprint id"
// @Success 200
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router /blueprints/{blueprintId} [delete]
func Delete(c *gin.Context) {
	blueprintId := c.Param("blueprintId")
	id, err := strconv.ParseUint(blueprintId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad blueprintId format supplied"))
		return
	}
	deleteBlueprintOnly, err := errors.Convert01(strconv.ParseBool(c.DefaultQuery("delete_blueprint_only", "false")))
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad blueprintId format supplied"))
		return
	}
	err = services.DeleteBlueprint(id, deleteBlueprintOnly)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error deleting blueprint"))
		return
	}
	shared.ApiOutputSuccess(c, nil, http.StatusOK)
}
