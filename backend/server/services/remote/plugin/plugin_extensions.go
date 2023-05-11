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
	"encoding/json"
	"fmt"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/server/services/remote/bridge"
	"github.com/apache/incubator-devlake/server/services/remote/models"
)

type (
	remoteMetricPlugin struct {
		*remotePluginImpl
	}
	remoteDatasourcePlugin struct {
		*remotePluginImpl
	}
)

func (p remoteMetricPlugin) MakeMetricPluginPipelinePlanV200(projectName string, options json.RawMessage) (plugin.PipelinePlan, errors.Error) {
	return nil, errors.Internal.New("Remote metric plugins not supported")
}

func (p remoteDatasourcePlugin) MakeDataSourcePipelinePlanV200(connectionId uint64, bpScopes []*plugin.BlueprintScopeV200, syncPolicy plugin.BlueprintSyncPolicy) (plugin.PipelinePlan, []plugin.Scope, errors.Error) {
	connection := p.connectionTabler.New()
	err := connectionHelper.FirstById(connection, connectionId)
	if err != nil {
		return nil, nil, err
	}

	db := basicRes.GetDal()
	var toolScopeTxRulePairs = make([]interface{}, len(bpScopes))
	for i, bpScope := range bpScopes {
		wrappedToolScope := p.scopeTabler.New()
		err = api.CallDB(db.First, wrappedToolScope, dal.Where("id = ?", bpScope.Id))
		if err != nil {
			return nil, nil, errors.Default.Wrap(err, fmt.Sprintf("error getting scope %s", bpScope.Name))
		}
		toolScope := models.ScopeModel{}
		err := wrappedToolScope.To(&toolScope)
		if err != nil {
			return nil, nil, err
		}
		txRule, err := p.getTxRule(db, toolScope)
		if err != nil {
			return nil, nil, err
		}
		toolScopeTxRulePairs[i] = []interface{}{wrappedToolScope.Unwrap(), txRule}
	}

	entities := bpScopes[0].Entities

	plan_data := models.PipelineData{}
	err = p.invoker.Call("make-pipeline", bridge.DefaultContext, toolScopeTxRulePairs, entities, connection.Unwrap()).Get(&plan_data)
	if err != nil {
		return nil, nil, err
	}

	var scopes = make([]plugin.Scope, len(plan_data.Scopes))
	for i, dynamicScope := range plan_data.Scopes {
		scope, err := dynamicScope.Load()
		if err != nil {
			return nil, nil, err
		}
		scopes[i] = scope
	}

	return plan_data.Plan, scopes, nil
}

var _ models.RemotePlugin = (*remoteMetricPlugin)(nil)
var _ plugin.MetricPluginBlueprintV200 = (*remoteMetricPlugin)(nil)
var _ models.RemotePlugin = (*remoteDatasourcePlugin)(nil)
var _ plugin.DataSourcePluginBlueprintV200 = (*remoteDatasourcePlugin)(nil)
