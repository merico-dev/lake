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
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	bridge2 "github.com/apache/incubator-devlake/services/remote/bridge"
	"github.com/apache/incubator-devlake/services/remote/models"
)

type (
	remoteMetricPlugin struct {
		*remotePluginImpl
	}
	remoteDatasourcePlugin struct {
		*remotePluginImpl
	}
)

func newMetricPlugin(info *models.PluginInfo, invoker bridge2.Invoker) models.RemotePlugin {
	plugin := newPlugin(info, invoker)
	return &remoteMetricPlugin{plugin}
}

func newDatasourcePlugin(info *models.PluginInfo, invoker bridge2.Invoker) models.RemotePlugin {
	plugin := newPlugin(info, invoker)
	return &remoteDatasourcePlugin{plugin}
}

func (p *remoteMetricPlugin) MakeMetricPluginPipelinePlanV200(projectName string, options json.RawMessage) (core.PipelinePlan, errors.Error) {
	plan := core.PipelinePlan{}
	err := p.invoker.Call("MakeMetricPluginPipelinePlanV200", bridge2.DefaultContext, projectName, options).Get(&plan)
	if err != nil {
		return nil, err
	}
	return plan, err
}

func (p *remoteDatasourcePlugin) MakeDataSourcePipelinePlanV200(connectionId uint64, bpScopes []*core.BlueprintScopeV200) (core.PipelinePlan, []core.Scope, errors.Error) {
	plan := core.PipelinePlan{}
	scopes := []models.PipelineScope{}
	err := p.invoker.Call("MakeDataSourcePluginPipelinePlanV200", bridge2.DefaultContext, connectionId, bpScopes).Get(&plan, &scopes)
	if err != nil {
		return nil, nil, err
	}
	var castedScopes []core.Scope
	for _, scope := range scopes {
		castedScopes = append(castedScopes, &models.WrappedPipelineScope{Scope: scope})
	}
	return plan, castedScopes, nil
}

var _ models.RemotePlugin = (*remoteMetricPlugin)(nil)
var _ core.MetricPluginBlueprintV200 = (*remoteMetricPlugin)(nil)
var _ models.RemotePlugin = (*remoteDatasourcePlugin)(nil)
var _ core.DataSourcePluginBlueprintV200 = (*remoteDatasourcePlugin)(nil)
