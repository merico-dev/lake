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

package plugin

import (
	"github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/server/services/remote/bridge"
)

type pluginAPI struct {
	invoker    bridge.Invoker
	connType   *models.DynamicTabler
	txRuleType *models.DynamicTabler
	scopeType  *models.DynamicTabler
	helper     *api.ConnectionApiHelper
}

func GetDefaultAPI(
	invoker bridge.Invoker,
	connType *models.DynamicTabler,
	txRuleType *models.DynamicTabler,
	scopeType *models.DynamicTabler,
	helper *api.ConnectionApiHelper,
) map[string]map[string]plugin.ApiResourceHandler {
	papi := &pluginAPI{
		invoker:    invoker,
		connType:   connType,
		txRuleType: txRuleType,
		scopeType:  scopeType,
		helper:     helper,
	}

	resources := map[string]map[string]plugin.ApiResourceHandler{
		"test": {
			"POST": papi.TestConnection,
		},
		"connections": {
			"POST": papi.PostConnections,
			"GET":  papi.ListConnections,
		},
		"connections/:connectionId": {
			"GET":    papi.GetConnection,
			"PATCH":  papi.PatchConnection,
			"DELETE": papi.DeleteConnection,
		},
		"connections/:connectionId/scopes": {
			"PUT": papi.PutScope,
			"GET": papi.ListScopes,
		},
		"connections/:connectionId/scopes/:scopeId": {
			"GET":    papi.GetScope,
			"PATCH":  papi.UpdateScope,
			"DELETE": papi.DeleteScope,
		},
		"connections/:connectionId/remote-scopes": {
			"GET": papi.GetRemoteScopes,
		},
		"connections/:connectionId/search-remote-scopes": {
			"GET": papi.SearchRemoteScopes,
		},
	}

	if txRuleType != nil {
		resources["connections/:connectionId/transformation_rules"] = map[string]plugin.ApiResourceHandler{
			"POST": papi.PostTransformationRules,
			"GET":  papi.ListTransformationRules,
		}
		resources["connections/:connectionId/transformation_rules/:id"] = map[string]plugin.ApiResourceHandler{
			"GET":   papi.GetTransformationRule,
			"PATCH": papi.PatchTransformationRule,
		}
	}
	scopeHelper = createScopeHelper(papi)
	return resources
}

func createScopeHelper(pa *pluginAPI) *api.GenericScopeApiHelper[any, any, any] {
	params := &api.ReflectionParameters{
		ScopeIdFieldName:  "Id",
		ScopeIdColumnName: "id",
		RawScopeParamName: "scope_id",
	}
	return api.NewGenericScopeHelper[any, any, any](
		basicRes,
		nil,
		connectionHelper,
		NewScopeDatabaseHelperImpl(pa, basicRes, params),
		params,
		&api.ScopeHelperOptions{},
	)
}
