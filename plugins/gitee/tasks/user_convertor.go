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

	"github.com/apache/incubator-devlake/models/domainlayer"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/models/domainlayer/user"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ConvertUsersMeta = core.SubTaskMeta{
	Name:             "convertUsers",
	EntryPoint:       ConvertUsers,
	EnabledByDefault: true,
	Description:      "Convert tool layer table gitee_users into  domain layer table users",
}

func ConvertUsers(taskCtx core.SubTaskContext) error {
	db := taskCtx.GetDb()
	rawDataSubTaskArgs, _ := CreateRawDataSubTaskArgs(taskCtx, RAW_COMMIT_TABLE)

	cursor, err := db.Model(&models.GiteeUser{}).
		Rows()
	if err != nil {
		return err
	}
	defer cursor.Close()

	userIdGen := didgen.NewDomainIdGenerator(&models.GiteeUser{})

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		InputRowType:       reflect.TypeOf(models.GiteeUser{}),
		Input:              cursor,
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Convert: func(inputRow interface{}) ([]interface{}, error) {
			GiteeUser := inputRow.(*models.GiteeUser)
			domainUser := &user.User{
				DomainEntity: domainlayer.DomainEntity{Id: userIdGen.Generate(GiteeUser.Id)},
				Name:         GiteeUser.Login,
				AvatarUrl:    GiteeUser.AvatarUrl,
			}
			return []interface{}{
				domainUser,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
