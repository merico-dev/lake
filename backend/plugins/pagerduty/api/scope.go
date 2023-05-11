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

package api

import (
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/pagerduty/models"
)

// Needed for Swagger doc gen
type (
	ScopeReq api.ScopeReq[models.Service]
	ScopeRes api.ScopeRes[models.Service]
)

// PutScope create or update pagerduty service
// @Summary create or update pagerduty service
// @Description Create or update pagerduty service
// @Tags plugins/pagerduty
// @Accept application/json
// @Param connectionId path int true "connection ID"
// @Param scope body ScopeReq true "json"
// @Success 200  {object} []models.Service
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router /plugins/pagerduty/connections/{connectionId}/scopes [PUT]
func PutScope(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return scopeHelper.Put(input)
}

// UpdateScope patch to pagerduty service
// @Summary patch to pagerduty service
// @Description patch to pagerduty service
// @Tags plugins/pagerduty
// @Accept application/json
// @Param connectionId path int true "connection ID"
// @Param serviceId path string true "service ID"
// @Param scope body models.Service true "json"
// @Success 200  {object} models.Service
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router /plugins/pagerduty/connections/{connectionId}/scopes/{serviceId} [PATCH]
func UpdateScope(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return scopeHelper.Update(input, "")
}

// GetScopeList get PagerDuty repos
// @Summary get PagerDuty repos
// @Description get PagerDuty repos
// @Tags plugins/pagerduty
// @Param connectionId path int true "connection ID"
// @Param pageSize query int false "page size, default 50"
// @Param page query int false "page size, default 1"
// @Param blueprints query bool false "also return blueprints using these scopes as part of the payload"
// @Success 200  {object} []ScopeRes
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router /plugins/pagerduty/connections/{connectionId}/scopes/ [GET]
func GetScopeList(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return scopeHelper.GetScopeList(input)
}

// GetScope get one PagerDuty service
// @Summary get one PagerDuty service
// @Description get one PagerDuty service
// @Tags plugins/pagerduty
// @Param connectionId path int true "connection ID"
// @Param serviceId path int true "service ID"
// @Param blueprints query bool false "also return blueprints using this scope as part of the payload"
// @Success 200  {object} ScopeRes
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router /plugins/pagerduty/connections/{connectionId}/scopes/{serviceId} [GET]
func GetScope(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return scopeHelper.GetScope(input, "")
}

// DeleteScope delete plugin data associated with the scope and optionally the scope itself
// @Summary delete plugin data associated with the scope and optionally the scope itself
// @Description delete data associated with plugin scope
// @Tags plugins/pagerduty
// @Param connectionId path int true "connection ID"
// @Param serviceId path int true "service ID"
// @Param delete_data_only query bool false "Only delete the scope data, not the scope itself"
// @Success 200  {object} []models.Blueprint "list of blueprints impacted by the deletion"
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router /plugins/pagerduty/connections/{connectionId}/scopes/{serviceId} [DELETE]
func DeleteScope(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return scopeHelper.Delete(input)
}
