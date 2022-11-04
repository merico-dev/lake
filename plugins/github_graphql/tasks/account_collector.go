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
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"github.com/apache/incubator-devlake/plugins/github/models"
	githubTasks "github.com/apache/incubator-devlake/plugins/github/tasks"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/merico-dev/graphql"
	"reflect"
)

const RAW_ACCOUNTS_TABLE = "github_graphql_accounts"

type GraphqlQueryAccountWrapper struct {
	RateLimit struct {
		Cost int
	}
	Users []GraphqlQueryAccount `graphql:"user(login: $login)" graphql-extend:"true"`
}

type GraphqlQueryAccount struct {
	Login     string
	Id        int `graphql:"databaseId"`
	Name      string
	Company   string
	Email     string
	AvatarUrl string
	HtmlUrl   string `graphql:"url"`
	//Type      string
	Organizations struct {
		Nodes []struct {
			Email      string
			Name       string
			DatabaseId int
			Login      string
		}
	} `graphql:"organizations(first: 10)"`
}

var CollectAccountMeta = core.SubTaskMeta{
	Name:             "CollectAccount",
	EntryPoint:       CollectAccount,
	EnabledByDefault: true,
	Description:      "Collect Account data from GithubGraphql api",
}

type SimpleAccount struct {
	Login string
}

var _ core.SubTaskEntryPoint = CollectAccount

func CollectAccount(taskCtx core.SubTaskContext) errors.Error {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*githubTasks.GithubTaskData)

	cursor, err := db.Cursor(
		dal.Select("login"),
		dal.From(models.GithubRepoAccount{}.TableName()),
		dal.Where("repo_github_id = ? and connection_id=?", data.Repo.GithubId, data.Options.ConnectionId),
	)
	if err != nil {
		return err
	}
	iterator, err := helper.NewDalCursorIterator(db, cursor, reflect.TypeOf(SimpleAccount{}))
	if err != nil {
		return err
	}

	collector, err := helper.NewGraphqlCollector(helper.GraphqlCollectorArgs{
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: githubTasks.GithubApiParams{
				ConnectionId: data.Options.ConnectionId,
				Owner:        data.Options.Owner,
				Repo:         data.Options.Repo,
			},
			Table: RAW_ACCOUNTS_TABLE,
		},
		Input:         iterator,
		InputStep:     100,
		GraphqlClient: data.GraphqlClient,
		BuildQuery: func(reqData *helper.GraphqlRequestData) (interface{}, map[string]interface{}, error) {
			accounts := reqData.Input.([]interface{})
			query := &GraphqlQueryAccountWrapper{}
			users := []map[string]interface{}{}
			for _, iAccount := range accounts {
				account := iAccount.(*SimpleAccount)
				users = append(users, map[string]interface{}{
					`login`: graphql.String(account.Login),
				})
			}
			variables := map[string]interface{}{
				"user": users,
			}
			return query, variables, nil
		},
		ResponseParserWithDataErrors: func(iQuery interface{}, variables map[string]interface{}, dataErrors []graphql.DataError) ([]interface{}, error) {
			for _, dataError := range dataErrors {
				// log and ignore
				taskCtx.GetLogger().Warn(dataError, `query user get error but ignore`)
			}
			query := iQuery.(*GraphqlQueryAccountWrapper)
			accounts := query.Users

			results := make([]interface{}, 0, 1)
			for _, account := range accounts {
				relatedUsers, err := convertAccount(account, data.Options.ConnectionId)
				if err != nil {
					return nil, err
				}
				results = append(results, relatedUsers...)
			}
			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return collector.Execute()
}

func convertAccount(res GraphqlQueryAccount, connId uint64) ([]interface{}, errors.Error) {
	results := make([]interface{}, 0, len(res.Organizations.Nodes)+1)
	githubAccount := &models.GithubAccount{
		ConnectionId: connId,
		Id:           res.Id,
		Login:        res.Login,
		Name:         res.Name,
		Company:      res.Company,
		Email:        res.Email,
		AvatarUrl:    res.AvatarUrl,
		//Url:          res.Url,
		HtmlUrl: res.HtmlUrl,
		Type:    `User`,
	}
	results = append(results, githubAccount)
	for _, apiAccountOrg := range res.Organizations.Nodes {
		githubAccountOrg := &models.GithubAccountOrg{
			ConnectionId: connId,
			AccountId:    res.Id,
			OrgId:        apiAccountOrg.DatabaseId,
			OrgLogin:     apiAccountOrg.Login,
		}
		results = append(results, githubAccountOrg)
	}

	return results, nil
}
