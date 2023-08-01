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

package tasks

import (
	"reflect"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/domainlayer"
	"github.com/apache/incubator-devlake/core/models/domainlayer/devops"
	"github.com/apache/incubator-devlake/core/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/bamboo/models"
)

var ConvertDeployBuildsMeta = plugin.SubTaskMeta{
	Name:             "convertDeployBuilds",
	EntryPoint:       ConvertDeployBuilds,
	EnabledByDefault: true,
	Description:      "Convert tool layer table bamboo_deploy_builds into  domain layer table deployBuilds",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_CICD},
}

func ConvertDeployBuilds(taskCtx plugin.SubTaskContext) errors.Error {
	db := taskCtx.GetDal()
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_JOB_BUILD_TABLE)
	cursor, err := db.Cursor(
		dal.From(&models.BambooDeployBuild{}),
		dal.Where("connection_id = ? and plan_key = ?", data.Options.ConnectionId, data.Options.PlanKey))
	if err != nil {
		return err
	}
	defer cursor.Close()

	deployBuildIdGen := didgen.NewDomainIdGenerator(&models.BambooDeployBuild{})
	planIdGen := didgen.NewDomainIdGenerator(&models.BambooPlan{})

	converter, err := api.NewDataConverter(api.DataConverterArgs{
		InputRowType:       reflect.TypeOf(models.BambooDeployBuild{}),
		Input:              cursor,
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Convert: func(inputRow interface{}) ([]interface{}, errors.Error) {
			deployBuild := inputRow.(*models.BambooDeployBuild)
			if deployBuild.StartedDate == nil {
				return nil, nil
			}
			domainTask := &devops.CICDTask{
				DomainEntity: domainlayer.DomainEntity{
					Id: deployBuildIdGen.Generate(data.Options.ConnectionId, deployBuild.DeployBuildId),
				},
				PipelineId:  deployBuildIdGen.Generate(data.Options.ConnectionId, deployBuild.DeployBuildId),
				CicdScopeId: planIdGen.Generate(data.Options.ConnectionId, data.Options.PlanKey),

				Name: deployBuild.DeploymentVersionName,

				Result: devops.GetResult(&devops.ResultRule{
					Failed:  []string{"Failed"},
					Success: []string{"Successful"},
					Default: "",
				}, deployBuild.DeploymentState),

				Status: devops.GetStatus(&devops.StatusRule{
					Done:    []string{"Finished"},
					Default: devops.IN_PROGRESS,
				}, deployBuild.LifeCycleState),

				StartedDate:  *deployBuild.StartedDate,
				FinishedDate: deployBuild.FinishedDate,
			}

			domainTask.Type = devops.DEPLOYMENT
			domainTask.Environment = deployBuild.Environment
			if data.RegexEnricher.ReturnNameIfMatched(models.ENV_NAME_PATTERN, deployBuild.Environment) != "" {
				domainTask.Environment = devops.PRODUCTION
			}
			if deployBuild.FinishedDate != nil && deployBuild.StartedDate != nil {
				sec := deployBuild.FinishedDate.Sub(*deployBuild.StartedDate).Seconds()
				if sec >= 0 {
					domainTask.DurationSec = uint64(sec)
				}
			}
			pipeline := &devops.CICDPipeline{
				DomainEntity: domainlayer.DomainEntity{
					Id: domainTask.PipelineId,
				},
				Name:         domainTask.Name,
				Result:       domainTask.Result,
				Status:       domainTask.Status,
				DurationSec:  domainTask.DurationSec,
				Type:         domainTask.Type,
				Environment:  domainTask.Environment,
				CreatedDate:  domainTask.StartedDate,
				FinishedDate: domainTask.FinishedDate,
				CicdScopeId:  domainTask.CicdScopeId,
			}

			return []interface{}{
				domainTask,
				pipeline,
			}, nil
		},
	})

	if err != nil {
		return err
	}

	return converter.Execute()
}
