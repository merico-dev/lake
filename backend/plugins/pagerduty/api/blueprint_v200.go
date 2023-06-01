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
	"fmt"
	"time"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/domainlayer"
	"github.com/apache/incubator-devlake/core/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/core/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/core/utils"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/pagerduty/models"
	"github.com/apache/incubator-devlake/plugins/pagerduty/tasks"
	"github.com/go-playground/validator/v10"
)

func MakeDataSourcePipelinePlanV200(subtaskMetas []plugin.SubTaskMeta, connectionId uint64, bpScopes []*plugin.BlueprintScopeV200, syncPolicy *plugin.BlueprintSyncPolicy,
) (plugin.PipelinePlan, []plugin.Scope, errors.Error) {
	connHelper := api.NewConnectionHelper(basicRes, validator.New())
	// get the connection info for url
	connection := &models.PagerDutyConnection{}
	err := connHelper.FirstById(connection, connectionId)
	if err != nil {
		return nil, nil, err
	}

	plan := make(plugin.PipelinePlan, len(bpScopes))
	plan, err = makeDataSourcePipelinePlanV200(subtaskMetas, plan, bpScopes, connection, syncPolicy)
	if err != nil {
		return nil, nil, err
	}
	scopes, err := makeScopesV200(bpScopes, connection)
	if err != nil {
		return nil, nil, err
	}

	return plan, scopes, nil
}

func makeDataSourcePipelinePlanV200(
	subtaskMetas []plugin.SubTaskMeta,
	plan plugin.PipelinePlan,
	bpScopes []*plugin.BlueprintScopeV200,
	connection *models.PagerDutyConnection,
	syncPolicy *plugin.BlueprintSyncPolicy,
) (plugin.PipelinePlan, errors.Error) {
	var err errors.Error
	for i, bpScope := range bpScopes {
		service := &models.Service{}
		// get repo from db
		err = basicRes.GetDal().First(service, dal.Where(`connection_id = ? AND id = ?`, connection.ID, bpScope.Id))
		if err != nil {
			return nil, errors.Default.Wrap(err, fmt.Sprintf("fail to find service %s", bpScope.Id))
		}
		scopeConfig := &models.PagerdutyScopeConfig{}
		// get scope configs from db
		db := basicRes.GetDal()
		err = db.First(scopeConfig, dal.Where(`id = ?`, service.ScopeConfigId))
		if err != nil {
			if db.IsErrorNotFound(err) {
				scopeConfig.Entities = []string{plugin.DOMAIN_TYPE_TICKET} //default to this
			} else {
				return nil, err
			}
		}
		// construct task options for pagerduty
		op := &tasks.PagerDutyOptions{
			ConnectionId: service.ConnectionId,
			ServiceId:    service.Id,
			ServiceName:  service.Name,
		}
		if syncPolicy.TimeAfter != nil {
			op.TimeAfter = syncPolicy.TimeAfter.Format(time.RFC3339)
		}
		var options map[string]any
		options, err = tasks.EncodeTaskOptions(op)
		if err != nil {
			return nil, err
		}
		var subtasks []string
		subtasks, err = api.MakePipelinePlanSubtasks(subtaskMetas, scopeConfig.Entities)
		if err != nil {
			return nil, err
		}
		stage := []*plugin.PipelineTask{
			{
				Plugin:   "pagerduty",
				Subtasks: subtasks,
				Options:  options,
			},
		}
		plan[i] = stage
	}
	return plan, nil
}

func makeScopesV200(bpScopes []*plugin.BlueprintScopeV200, connection *models.PagerDutyConnection) ([]plugin.Scope, errors.Error) {
	scopes := make([]plugin.Scope, 0)
	for _, bpScope := range bpScopes {
		db := basicRes.GetDal()
		// get service from db
		service := &models.Service{}
		err := db.First(service, dal.Where(`connection_id = ? AND id = ?`, connection.ID, bpScope.Id))
		if err != nil {
			return nil, errors.Default.Wrap(err, fmt.Sprintf("failed to find service: %s", bpScope.Id))
		}
		// get scope configs from db
		scopeConfig := &models.PagerdutyScopeConfig{}
		err = db.First(scopeConfig, dal.Where(`id = ?`, service.ScopeConfigId))
		if err != nil && !db.IsErrorNotFound(err) {
			return nil, err
		}
		// add board to scopes
		if utils.StringsContains(scopeConfig.Entities, plugin.DOMAIN_TYPE_TICKET) {
			scopeTicket := &ticket.Board{
				DomainEntity: domainlayer.DomainEntity{
					Id: didgen.NewDomainIdGenerator(&models.Service{}).Generate(connection.ID, service.Id),
				},
				Name: service.Name,
			}
			scopes = append(scopes, scopeTicket)
		}
	}
	return scopes, nil
}
