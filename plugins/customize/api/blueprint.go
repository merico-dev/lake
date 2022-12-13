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
	"encoding/json"
	"github.com/apache/incubator-devlake/errors"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/plugins/jira/tasks"
)

func MakePipelinePlan(subtaskMetas []core.SubTaskMeta, connectionId uint64, scope []*core.BlueprintScopeV100) (core.PipelinePlan, errors.Error) {
	var err error
	plan := make(core.PipelinePlan, len(scope))
	for i, scopeElem := range scope {
		taskOptions := make(map[string]interface{})
		err = json.Unmarshal(scopeElem.Options, &taskOptions)
		if err != nil {
			return nil, errors.Convert(err)
		}
		op, err := tasks.DecodeTaskOptions(taskOptions)
		if err != nil {
			return nil, err
		}
		err = tasks.ValidateTaskOptions(op)
		if err != nil {
			return nil, err
		}
		// subtasks
		subtasks, err := helper.MakePipelinePlanSubtasks(subtaskMetas, scopeElem.Entities)
		if err != nil {
			return nil, err
		}
		plan[i] = core.PipelineStage{
			{
				Plugin:   "customize",
				Subtasks: subtasks,
				Options:  taskOptions,
			},
		}
	}
	return plan, nil
}
