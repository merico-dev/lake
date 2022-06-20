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

	"github.com/apache/incubator-devlake/api/shared"

	"github.com/apache/incubator-devlake/models"
	"github.com/apache/incubator-devlake/services"
	"github.com/gin-gonic/gin"
)

// @Summary post blueprints
// @Description post blueprints
// @Tags Blueprints
// @Accept application/json
// @Param blueprint body string true "json"
// @Success 200  {object} models.Blueprint
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /blueprints [post]
func Post(c *gin.Context) {
	blueprint := &models.Blueprint{}

	err := c.ShouldBind(blueprint)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusBadRequest)
		return
	}

	err = services.CreateBlueprint(blueprint)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusBadRequest)
		return
	}

	shared.ApiOutputSuccess(c, blueprint, http.StatusCreated)
}

// @Summary get blueprints
// @Description get blueprints
// @Tags Blueprints
// @Accept application/json
// @Success 200  {object} gin.H
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /blueprints [get]
func Index(c *gin.Context) {
	var query services.BlueprintQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusBadRequest)
		return
	}
	blueprints, count, err := services.GetBlueprints(&query)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	shared.ApiOutputSuccess(c, gin.H{"blueprints": blueprints, "count": count}, http.StatusOK)
}

// @Summary get blueprints
// @Description get blueprints
// @Tags Blueprints
// @Accept application/json
// @Param blueprintId path int true "blueprint id"
// @Success 200  {object} models.Blueprint
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /blueprints/{blueprintId} [get]
func Get(c *gin.Context) {
	blueprintId := c.Param("blueprintId")
	id, err := strconv.ParseUint(blueprintId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusBadRequest)
		return
	}
	blueprint, err := services.GetBlueprint(id)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusBadRequest)
		return
	}
	shared.ApiOutputSuccess(c, blueprint, http.StatusOK)
}

// @Summary delete blueprints
// @Description Delete BluePrints
// @Tags Blueprints
// @Param blueprintId path string true "blueprintId"
// @Success 200
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /blueprints/{blueprintId} [delete]
func Delete(c *gin.Context) {
	pipelineId := c.Param("blueprintId")
	id, err := strconv.ParseUint(pipelineId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusBadRequest)
		return
	}
	err = services.DeleteBlueprint(id)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusBadRequest)
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
// @Tags Blueprints
// @Accept application/json
// @Param blueprintId path string true "blueprintId"
// @Success 200  {object} models.Blueprint
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /blueprints/{blueprintId} [Patch]
func Patch(c *gin.Context) {
	blueprintId := c.Param("blueprintId")
	id, err := strconv.ParseUint(blueprintId, 10, 64)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusBadRequest)
		return
	}
	var body map[string]interface{}
	err = c.ShouldBind(&body)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusBadRequest)
		return
	}
	blueprint, err := services.PatchBlueprint(id, body)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusBadRequest)
		return
	}
	shared.ApiOutputSuccess(c, blueprint, http.StatusOK)
}
