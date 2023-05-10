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

package impl

import (
	"fmt"
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/pagerduty/api"
	"github.com/apache/incubator-devlake/plugins/pagerduty/models"
	"github.com/apache/incubator-devlake/plugins/pagerduty/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/pagerduty/tasks"
	"time"
)

// make sure interface is implemented
var _ plugin.PluginMeta = (*PagerDuty)(nil)
var _ plugin.PluginInit = (*PagerDuty)(nil)
var _ plugin.PluginTask = (*PagerDuty)(nil)
var _ plugin.PluginApi = (*PagerDuty)(nil)

var _ plugin.PluginModel = (*PagerDuty)(nil)
var _ plugin.PluginBlueprintV100 = (*PagerDuty)(nil)
var _ plugin.CloseablePluginTask = (*PagerDuty)(nil)

type PagerDuty struct{}

func (p PagerDuty) Description() string {
	return "collect some PagerDuty data"
}

func (p PagerDuty) Init(basicRes context.BasicRes) errors.Error {
	api.Init(basicRes)
	return nil
}

func (p PagerDuty) SubTaskMetas() []plugin.SubTaskMeta {
	return []plugin.SubTaskMeta{
		tasks.CollectIncidentsMeta,
		tasks.ExtractIncidentsMeta,
		tasks.ConvertIncidentsMeta,
	}
}

func (p PagerDuty) GetTablesInfo() []dal.Tabler {
	return []dal.Tabler{
		models.Service{},
		models.Incident{},
		models.User{},
		models.Assignment{},
	}
}

func (p PagerDuty) PrepareTaskData(taskCtx plugin.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	op, err := tasks.DecodeAndValidateTaskOptions(options)
	if err != nil {
		return nil, err
	}
	connectionHelper := helper.NewConnectionHelper(
		taskCtx,
		nil,
	)
	connection := &models.PagerDutyConnection{}
	err = connectionHelper.FirstById(connection, op.ConnectionId)
	if err != nil {
		return nil, errors.Default.Wrap(err, "unable to get Pagerduty connection by the given connection ID")
	}
	var timeAfter *time.Time
	if op.TimeAfter != "" {
		convertedTime, err := errors.Convert01(time.Parse(time.RFC3339, op.TimeAfter))
		if err != nil {
			return nil, errors.BadInput.Wrap(err, fmt.Sprintf("invalid value for `timeAfter`: %s", timeAfter))
		}
		timeAfter = &convertedTime
	}
	client, err := helper.NewApiClient(taskCtx.GetContext(), connection.Endpoint, map[string]string{
		"Authorization": fmt.Sprintf("Token %s", connection.Token),
	}, 0, connection.Proxy, taskCtx)
	if err != nil {
		return nil, err
	}
	asyncClient, err := helper.CreateAsyncApiClient(taskCtx, client, nil)
	if err != nil {
		return nil, err
	}
	return &tasks.PagerDutyTaskData{
		Options:   op,
		TimeAfter: timeAfter,
		Client:    asyncClient,
	}, nil
}

// PkgPath information lost when compiled as plugin(.so)
func (p PagerDuty) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/pagerduty"
}

func (p PagerDuty) MigrationScripts() []plugin.MigrationScript {
	return migrationscripts.All()
}

func (p PagerDuty) ApiResources() map[string]map[string]plugin.ApiResourceHandler {
	return map[string]map[string]plugin.ApiResourceHandler{
		"test": {
			"POST": api.TestConnection,
		},
		"connections": {
			"POST": api.PostConnections,
			"GET":  api.ListConnections,
		},
		"connections/:connectionId": {
			"GET":    api.GetConnection,
			"PATCH":  api.PatchConnection,
			"DELETE": api.DeleteConnection,
		},
		"connections/:connectionId/remote-scopes": {
			"GET": api.RemoteScopes,
		},
		"connections/:connectionId/search-remote-scopes": {
			"GET": api.SearchRemoteScopes,
		},
		"connections/:connectionId/scopes": {
			"GET": api.GetScopeList,
			"PUT": api.PutScope,
		},
		"connections/:connectionId/scopes/:scopeId": {
			"GET":    api.GetScope,
			"PATCH":  api.UpdateScope,
			"DELETE": api.DeleteScope,
		},
		"connections/:connectionId/transformation_rules": {
			"POST": api.CreateTransformationRule,
			"GET":  api.GetTransformationRuleList,
		},
		"connections/:connectionId/transformation_rules/:id": {
			"PATCH": api.UpdateTransformationRule,
			"GET":   api.GetTransformationRule,
		},
	}
}

func (p PagerDuty) MakePipelinePlan(connectionId uint64, scope []*plugin.BlueprintScopeV100) (plugin.PipelinePlan, errors.Error) {
	return api.MakePipelinePlan(p.SubTaskMetas(), connectionId, scope)
}

func (p PagerDuty) MakeDataSourcePipelinePlanV200(connectionId uint64, scopes []*plugin.BlueprintScopeV200, syncPolicy plugin.BlueprintSyncPolicy,
) (plugin.PipelinePlan, []plugin.Scope, errors.Error) {
	return api.MakeDataSourcePipelinePlanV200(p.SubTaskMetas(), connectionId, scopes, &syncPolicy)
}

func (p PagerDuty) Close(taskCtx plugin.TaskContext) errors.Error {
	_, ok := taskCtx.GetData().(*tasks.PagerDutyTaskData)
	if !ok {
		return errors.Default.New(fmt.Sprintf("GetData failed when try to close %+v", taskCtx))
	}
	return nil
}
