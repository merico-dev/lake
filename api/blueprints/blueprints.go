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
	"github.com/apache/incubator-devlake/errors"
	"net/http"
	"strconv"

	"github.com/apache/incubator-devlake/api/shared"
	"github.com/apache/incubator-devlake/models"
	"github.com/apache/incubator-devlake/services"
	"github.com/gin-gonic/gin"
)

// @Summary post blueprints
// @Description post blueprints
// @Tags framework/blueprints
// @Accept application/json
// @Param blueprint body models.ApiBlueprint true "json"
// @Success 200  {object} models.ApiBlueprint
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /blueprints [post]
func Post(c *gin.Context) {
	blueprint := &models.ApiBlueprint{}

	err := c.ShouldBind(blueprint)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad request body format", errors.AsUserMessage()))
		return
	}

	err = services.CreateApiBlueprint(blueprint)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error creating blueprint", errors.AsUserMessage()))
		return
	}

	shared.ApiOutputSuccess(c, blueprint, http.StatusCreated)
}

// @Summary get blueprints
// @Description get blueprints
// @Tags framework/blueprints
// @Accept application/json
// @Success 200  {object} gin.H
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /blueprints [get]
func Index(c *gin.Context) {
	var query services.BlueprintQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad request body format", errors.AsUserMessage()))
		return
	}
	blueprints, count, err := services.GetApiBlueprints(&query)
	if err != nil {
		shared.ApiOutputAbort(c, errors.Default.Wrap(err, "error getting blueprints", errors.AsUserMessage()))
		return
	}
	shared.ApiOutputSuccess(c, gin.H{"blueprints": blueprints, "count": count}, http.StatusOK)
}

// @Summary get blueprints
// @Description get blueprints
// @Tags framework/blueprints
// @Accept application/json
// @Param blueprintId path int true "blueprint id"
// @Success 200  {object} models.ApiBlueprint
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /blueprints/{blueprintId} [get]
func Get(c *gin.Context) {
	blueprintId := c.Param("blueprintId")
	id, err := strconv.ParseUint(blueprintId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad blueprintId format supplied", errors.AsUserMessage()))
		return
	}
	blueprint, err := services.GetApiBlueprint(id)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error getting blueprint", errors.AsUserMessage()))
		return
	}
	shared.ApiOutputSuccess(c, blueprint, http.StatusOK)
}

// @Summary delete blueprints
// @Description Delete BluePrints
// @Tags framework/blueprints
// @Param blueprintId path string true "blueprintId"
// @Success 200
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /blueprints/{blueprintId} [delete]
func Delete(c *gin.Context) {
	pipelineId := c.Param("blueprintId")
	id, err := strconv.ParseUint(pipelineId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad blueprintID format supplied", errors.AsUserMessage()))
		return
	}
	err = services.DeleteApiBlueprint(id)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error deleting blueprint", errors.AsUserMessage()))
	}
}

/*
func Put(c *gin.Context) {
	blueprintId := c.Param("blueprintId")
	id, err := strconv.ParseUint(blueprintId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusBadRequest)
		return
	}
	editBlueprint := &models.EditBlueprint{}
	err = c.MustBindWith(editBlueprint, binding.JSON)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	editBlueprint.BlueprintId = id
	blueprint, err := services.ModifyBlueprint(editBlueprint)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusBadRequest)
		return
	}
	shared.ApiOutputSuccess(c, blueprint, http.StatusOK)
}
*/

// @Summary patch blueprints
// @Description patch blueprints
// @Tags framework/blueprints
// @Accept application/json
// @Param blueprintId path string true "blueprintId"
// @Success 200  {object} models.ApiBlueprint
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /blueprints/{blueprintId} [Patch]
func Patch(c *gin.Context) {
	blueprintId := c.Param("blueprintId")
	id, err := strconv.ParseUint(blueprintId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad pipeline ID format supplied", errors.AsUserMessage()))
		return
	}
	var body map[string]interface{}
	err = c.ShouldBind(&body)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad request body format", errors.AsUserMessage()))
		return
	}
	blueprint, err := services.PatchApiBlueprint(id, body)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error patching the blueprint", errors.AsUserMessage()))
		return
	}
	shared.ApiOutputSuccess(c, blueprint, http.StatusOK)
}

// @Summary trigger blueprint
// @Description trigger a blueprint immediately
// @Tags framework/blueprints
// @Accept application/json
// @Param blueprintId path string true "blueprintId"
// @Success 200  {object} models.ApiPipeline
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /blueprints/{blueprintId}/trigger [Post]
func Trigger(c *gin.Context) {
	blueprintId := c.Param("blueprintId")
	id, err := strconv.ParseUint(blueprintId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad blueprintID format supplied", errors.AsUserMessage()))
		return
	}
	pipeline, err := services.TriggerBlueprint(id)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error triggering blueprint", errors.AsUserMessage()))
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
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /blueprints/{blueprintId}/pipelines [get]
func GetBlueprintPipelines(c *gin.Context) {
	blueprintId := c.Param("blueprintId")
	id, err := strconv.ParseUint(blueprintId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, errors.BadInput.Wrap(err, "bad blueprintID at supplied", errors.AsUserMessage()))
		return
	}
	var query services.PipelineQuery
	query.BlueprintId = id
	pipelines, count, err := services.GetPipelines(&query)
	if err != nil {
		shared.ApiOutputError(c, errors.Default.Wrap(err, "error getting pipelines", errors.AsUserMessage()))
		return
	}
	shared.ApiOutputSuccess(c, shared.ResponsePipelines{Pipelines: pipelines, Count: count}, http.StatusOK)
}
