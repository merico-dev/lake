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
	"github.com/apache/incubator-devlake/models/common"
	"reflect"

	"github.com/apache/incubator-devlake/models/domainlayer/crossdomain"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ConnectUserAccountsExactMeta = core.SubTaskMeta{
	Name:             "connectUserAccountsExact",
	EntryPoint:       ConnectUserAccountsExact,
	EnabledByDefault: true,
	Description:      "associate users and accounts",
	DomainTypes:      []string{core.DOMAIN_TYPE_CROSS},
}

func ConnectUserAccountsExact(taskCtx core.SubTaskContext) error {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*TaskData)
	type input struct {
		UserId    string
		AccountId string
		common.NoPKModel
	}
	var users []crossdomain.User
	err := db.All(&users)
	if err != nil {
		return err
	}
	err = db.Delete(&crossdomain.UserAccount{}, dal.Where("1=1"))
	if err != nil {
		return err
	}
	emails := make(map[string]string)
	names := make(map[string]string)
	for _, user := range users {
		emails[user.Email] = user.Id
		names[user.Name] = user.Id
	}
	clauses := []dal.Clause{
		dal.Select("*"),
		dal.From(&crossdomain.Account{}),
	}
	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	}
	defer cursor.Close()

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		InputRowType: reflect.TypeOf(crossdomain.Account{}),
		Input:        cursor,
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: Params{
				ConnectionId: data.Options.ConnectionId,
			},
			Table: "users",
		},

		Convert: func(inputRow interface{}) ([]interface{}, error) {
			account := inputRow.(*crossdomain.Account)
			if userId, ok := emails[account.Email]; ok {
				return []interface{}{
					&crossdomain.UserAccount{
						UserId:    userId,
						AccountId: account.Id,
					},
				}, nil
			}
			if userId, ok := names[account.FullName]; ok {
				return []interface{}{
					&crossdomain.UserAccount{
						UserId:    userId,
						AccountId: account.Id,
					},
				}, nil
			}
			if userId, ok := names[account.UserName]; ok {
				return []interface{}{
					&crossdomain.UserAccount{
						UserId:    userId,
						AccountId: account.Id,
					},
				}, nil
			}
			return nil, nil
		},
	})
	if err != nil {
		return err
	}
	return converter.Execute()
}
