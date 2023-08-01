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
	"github.com/apache/incubator-devlake/core/models/domainlayer/devops"
	"reflect"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/bamboo/models"
)

var ConvertPlanVcsMeta = plugin.SubTaskMeta{
	Name:             "convertPlanVcs",
	EntryPoint:       ConvertPlanVcs,
	EnabledByDefault: true,
	Description:      "Convert tool layer table bamboo_planBuilds into  domain layer table planBuilds",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_CICD},
}

func ConvertPlanVcs(taskCtx plugin.SubTaskContext) errors.Error {
	db := taskCtx.GetDal()
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_PLAN_BUILD_TABLE)
	cursor, err := db.Cursor(
		dal.From(&models.BambooPlanBuildVcsRevision{}),
		dal.Join(`left join _tool_bamboo_plan_builds on _tool_bamboo_plan_builds.plan_build_key = _tool_bamboo_plan_build_commits.plan_build_key`),
		dal.Where("_tool_bamboo_plan_build_commits.connection_id = ? and _tool_bamboo_plan_builds.plan_key = ?", data.Options.ConnectionId, data.Options.PlanKey))
	if err != nil {
		return err
	}
	defer cursor.Close()

	planBuildIdGen := didgen.NewDomainIdGenerator(&models.BambooPlanBuild{})
	repoMap := getRepoMap(data.Options.RepoMap)
	converter, err := api.NewDataConverter(api.DataConverterArgs{
		InputRowType:       reflect.TypeOf(models.BambooPlanBuildVcsRevision{}),
		Input:              cursor,
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Convert: func(inputRow interface{}) ([]interface{}, errors.Error) {
			line := inputRow.(*models.BambooPlanBuildVcsRevision)
			domainPlanVcs := &devops.CiCDPipelineCommit{
				PipelineId: planBuildIdGen.Generate(data.Options.ConnectionId, line.PlanBuildKey),
				CommitSha:  line.VcsRevisionKey,
				RepoUrl:    line.RepositoryName,
			}
			domainPlanVcs.RepoId = repoMap[line.RepositoryId]
			return []interface{}{
				domainPlanVcs,
			}, nil
		},
	})

	if err != nil {
		return err
	}

	return converter.Execute()
}
