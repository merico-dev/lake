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
	"fmt"
	"reflect"

	"github.com/apache/incubator-devlake/models/domainlayer"
	"github.com/apache/incubator-devlake/models/domainlayer/code"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ConvertRepoMeta = core.SubTaskMeta{
	Name:             "convertRepo",
	EntryPoint:       ConvertRepo,
	EnabledByDefault: true,
	Description:      "Convert tool layer table gitee_repos into  domain layer table repos and boards",
}

func ConvertRepo(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_REPOSITORIES_TABLE)
	db := taskCtx.GetDb()
	repoId := data.Repo.GiteeId

	cursor, err := db.Model(&models.GiteeRepo{}).
		Where("gitee_id = ?", repoId).
		Rows()
	if err != nil {
		return err
	}
	defer cursor.Close()

	repoIdGen := didgen.NewDomainIdGenerator(&models.GiteeRepo{})

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		InputRowType:       reflect.TypeOf(models.GiteeRepo{}),
		Input:              cursor,
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Convert: func(inputRow interface{}) ([]interface{}, error) {
			repository := inputRow.(*models.GiteeRepo)
			domainRepository := &code.Repo{
				DomainEntity: domainlayer.DomainEntity{
					Id: repoIdGen.Generate(repository.GiteeId),
				},
				Name:        fmt.Sprintf("%s/%s", repository.OwnerLogin, repository.Name),
				Url:         repository.HTMLUrl,
				Description: repository.Description,
				ForkedFrom:  repository.ParentHTMLUrl,
				Language:    repository.Language,
				CreatedDate: repository.CreatedDate,
				UpdatedDate: repository.UpdatedDate,
			}

			domainBoard := &ticket.Board{
				DomainEntity: domainlayer.DomainEntity{
					Id: repoIdGen.Generate(repository.GiteeId),
				},
				Name:        fmt.Sprintf("%s/%s", repository.OwnerLogin, repository.Name),
				Url:         fmt.Sprintf("%s/%s", repository.HTMLUrl, "issues"),
				Description: repository.Description,
				CreatedDate: &repository.CreatedDate,
			}

			return []interface{}{
				domainRepository,
				domainBoard,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
