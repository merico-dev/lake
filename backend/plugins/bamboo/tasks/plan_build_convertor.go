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

var ConvertPlanBuildsMeta = plugin.SubTaskMeta{
	Name:             "convertPlanBuilds",
	EntryPoint:       ConvertPlanBuilds,
	EnabledByDefault: true,
	Description:      "Convert tool layer table bamboo_planBuilds into  domain layer table planBuilds",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_CICD},
}

func ConvertPlanBuilds(taskCtx plugin.SubTaskContext) errors.Error {
	db := taskCtx.GetDal()
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_PLAN_BUILD_TABLE)
	cursor, err := db.Cursor(
		dal.From(&models.BambooPlanBuild{}),
		dal.Where("connection_id = ? and plan_key = ?", data.Options.ConnectionId, data.Options.PlanKey))
	if err != nil {
		return err
	}
	defer cursor.Close()

	planBuildIdGen := didgen.NewDomainIdGenerator(&models.BambooPlanBuild{})
	planIdGen := didgen.NewDomainIdGenerator(&models.BambooPlan{})

	converter, err := api.NewDataConverter(api.DataConverterArgs{
		InputRowType:       reflect.TypeOf(models.BambooPlanBuild{}),
		Input:              cursor,
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Convert: func(inputRow interface{}) ([]interface{}, errors.Error) {
			line := inputRow.(*models.BambooPlanBuild)
			if line.BuildStartedTime == nil {
				return nil, nil
			}
			domainPlanBuild := &devops.CICDPipeline{
				DomainEntity: domainlayer.DomainEntity{Id: planBuildIdGen.Generate(data.Options.ConnectionId, line.PlanBuildKey)},
				Name:         line.PlanName,
				DurationSec:  uint64(line.BuildDurationInSeconds),
				CreatedDate:  *line.BuildStartedTime,
				FinishedDate: line.BuildCompletedDate,
				CicdScopeId:  planIdGen.Generate(data.Options.ConnectionId, data.Options.PlanKey),

				Result: devops.GetResult(&devops.ResultRule{
					Failed:  []string{"Failed"},
					Success: []string{"Successful"},
					Default: "",
				}, line.BuildState),

				Status: devops.GetStatus(&devops.StatusRule{
					Done:    []string{"Finished"},
					Default: devops.IN_PROGRESS,
				}, line.LifeCycleState),
			}

			domainPlanBuild.Type = line.Type
			domainPlanBuild.Environment = line.Environment

			return []interface{}{
				domainPlanBuild,
			}, nil
		},
	})

	if err != nil {
		return err
	}

	return converter.Execute()
}
