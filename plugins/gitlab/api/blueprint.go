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
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitlab/models"
	"github.com/apache/incubator-devlake/plugins/gitlab/tasks"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/utils"
)

func MakePipelinePlan(subtaskMetas []core.SubTaskMeta, connectionId uint64, scope []*core.BlueprintScopeV100) (core.PipelinePlan, errors.Error) {
	var err errors.Error
	plan := make(core.PipelinePlan, len(scope))
	for i, scopeElem := range scope {
		plan, err = processScope(subtaskMetas, connectionId, scopeElem, i, plan, nil, nil)
		if err != nil {
			return nil, err
		}
	}
	return plan, nil
}

func processScope(subtaskMetas []core.SubTaskMeta, connectionId uint64, scopeElem *core.BlueprintScopeV100, i int, plan core.PipelinePlan, apiRepo *tasks.GitlabApiProject, connection *models.GitlabConnection) (core.PipelinePlan, errors.Error) {
	var err errors.Error
	// handle taskOptions and transformationRules, by dumping them to taskOptions
	transformationRules := make(map[string]interface{})
	if len(scopeElem.Transformation) > 0 {
		err = errors.Convert(json.Unmarshal(scopeElem.Transformation, &transformationRules))
		if err != nil {
			return nil, err
		}
	}
	// refdiff
	if refdiffRules, ok := transformationRules["refdiff"]; ok && refdiffRules != nil {
		// add a new task to next stage
		j := i + 1
		if j == len(plan) {
			plan = append(plan, nil)
		}
		plan[j] = core.PipelineStage{
			{
				Plugin:  "refdiff",
				Options: refdiffRules.(map[string]interface{}),
			},
		}
		// remove it from github transformationRules
		delete(transformationRules, "refdiff")
	}
	// construct task options for github
	options := make(map[string]interface{})
	err = errors.Convert(json.Unmarshal(scopeElem.Options, &options))
	if err != nil {
		return nil, err
	}
	options["connectionId"] = connectionId
	options["transformationRules"] = transformationRules
	// make sure task options is valid
	op, err := tasks.DecodeAndValidateTaskOptions(options)
	if err != nil {
		return nil, err
	}
	// construct subtasks
	subtasks, err := helper.MakePipelinePlanSubtasks(subtaskMetas, scopeElem.Entities)
	if err != nil {
		return nil, err
	}
	stage := plan[i]
	if stage == nil {
		stage = core.PipelineStage{}
	}
	stage = append(stage, &core.PipelineTask{
		Plugin:   "gitlab",
		Subtasks: subtasks,
		Options:  options,
	})
	// collect git data by gitextractor if CODE was requested
	if utils.StringsContains(scopeElem.Entities, core.DOMAIN_TYPE_CODE) {
		// here is the tricky part, we have to obtain the repo id beforehand
		if connection == nil {
			connection = new(models.GitlabConnection)
			err = connectionHelper.FirstById(connection, connectionId)
			if err != nil {
				return nil, err
			}
		}
		token := strings.Split(connection.Token, ",")[0]
		if apiRepo == nil {
			apiRepo = new(tasks.GitlabApiProject)
			err = getApiRepo(connection, token, op, apiRepo)
			if err != nil {
				return nil, err
			}
		}
		cloneUrl, err := errors.Convert01(url.Parse(apiRepo.HttpUrlToRepo))
		if err != nil {
			return nil, err
		}
		cloneUrl.User = url.UserPassword("git", token)
		stage = append(stage, &core.PipelineTask{
			Plugin: "gitextractor",
			Options: map[string]interface{}{
				"url":    cloneUrl.String(),
				"repoId": didgen.NewDomainIdGenerator(&models.GitlabProject{}).Generate(connectionId, apiRepo.GitlabId),
				"proxy":  connection.Proxy,
			},
		})
	}
	// dora
	if doraRules, ok := transformationRules["dora"]; ok && doraRules != nil {
		j := i + 1
		// add a new task to next stage
		if plan[j] != nil {
			j++
		}
		if j == len(plan) {
			plan = append(plan, nil)
		}
		if err != nil {
			return nil, err
		}
		if apiRepo == nil {
			if connection == nil {
				connection = new(models.GitlabConnection)
				err = connectionHelper.FirstById(connection, connectionId)
				if err != nil {
					return nil, err
				}
			}
			token := strings.Split(connection.Token, ",")[0]
			apiRepo = new(tasks.GitlabApiProject)
			err = getApiRepo(connection, token, op, apiRepo)
			if err != nil {
				return nil, err
			}
		}
		doraOption := make(map[string]interface{})
		doraOption["repoId"] = didgen.NewDomainIdGenerator(&models.GitlabProject{}).Generate(connectionId, apiRepo.GitlabId)
		doraOption["tasks"] = []string{"EnrichTaskEnv"}
		doraOption["transformation"] = doraRules
		plan[j] = core.PipelineStage{
			{
				Plugin:  "dora",
				Options: doraOption,
			},
		}
		// remove it from github transformationRules
		delete(transformationRules, "dora")
	}
	plan[i] = stage
	return plan, nil
}

func getApiRepo(connection *models.GitlabConnection, token string, op *tasks.GitlabOptions, apiRepo *tasks.GitlabApiProject) errors.Error {
	apiClient, err := helper.NewApiClient(
		context.TODO(),
		connection.Endpoint,
		map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		},
		10*time.Second,
		connection.Proxy,
		BasicRes,
	)
	if err != nil {
		return err
	}
	res, err := apiClient.Get(fmt.Sprintf("projects/%d", op.ProjectId), nil, nil)
	if err != nil {
		return err
	}
	//defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.HttpStatus(res.StatusCode).New(fmt.Sprintf("unexpected status code when requesting repo detail from %s", res.Request.URL.String()))
	}
	body, err := errors.Convert01(io.ReadAll(res.Body))
	if err != nil {
		return err
	}
	err = errors.Convert(json.Unmarshal(body, apiRepo))
	if err != nil {
		return err
	}
	return nil
}
