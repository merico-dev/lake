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
	clauses := []dal.Clause{
		dal.Select("users.id As user_id, accounts.id As account_id"),
		dal.From(&crossdomain.User{}),
		dal.Join(`INNER JOIN accounts
				ON users.email = accounts.email
                  OR users.name = accounts.full_name
                  OR users.name = accounts.user_name `),
	}
	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	}
	defer cursor.Close()

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		InputRowType: reflect.TypeOf(input{}),
		Input:        cursor,
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: Params{
				ConnectionId: data.Options.ConnectionId,
			},
			Table: "users",
		},

		Convert: func(inputRow interface{}) ([]interface{}, error) {
			userAccount := inputRow.(*input)
			return []interface{}{
				&crossdomain.UserAccount{
					UserId:    userAccount.UserId,
					AccountId: userAccount.AccountId,
				},
			}, nil
		},
	})
	if err != nil {
		return err
	}
	return converter.Execute()
}
