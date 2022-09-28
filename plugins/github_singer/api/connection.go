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
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"net/http"
	"time"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/plugins/github_singer/models"
)

//TODO Please modify the following code to fit your needs
func TestConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	// decode
	var err errors.Error
	var connection models.TestConnectionRequest
	if err = helper.Decode(input.Body, &connection, vld); err != nil {
		return nil, err
	}
	// test connection
	apiClient, err := helper.NewApiClient(
		context.TODO(),
		connection.Endpoint,
		map[string]string{
			"Authorization": fmt.Sprintf("Bearer %v", connection.Token),
		},
		3*time.Second,
		connection.Proxy,
		basicRes,
	)
	if err != nil {
		return nil, err
	}

	res, err := apiClient.Get("user", nil, nil)
	if err != nil {
		return nil, err
	}
	resBody := &models.ApiUserResponse{}
	err = helper.UnmarshalResponse(res, resBody)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.HttpStatus(res.StatusCode).New(fmt.Sprintf("unexpected status code: %d", res.StatusCode))
	}
	return nil, nil
}

//TODO Please modify the folowing code to adapt to your plugin
/*
POST /plugins/GithubSinger/connections
{
	"name": "GithubSinger data connection name",
	"endpoint": "GithubSinger api endpoint, i.e. https://example.com",
	"username": "username, usually should be email address",
	"password": "GithubSinger api access token"
}
*/
func PostConnections(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	// update from request and save to database
	connection := &models.GithubSingerConnection{}
	err := connectionHelper.Create(connection, input)
	if err != nil {
		return nil, err
	}
	return &core.ApiResourceOutput{Body: connection, Status: http.StatusOK}, nil
}

//TODO Please modify the folowing code to adapt to your plugin
/*
PATCH /plugins/GithubSinger/connections/:connectionId
{
	"name": "GithubSinger data connection name",
	"endpoint": "GithubSinger api endpoint, i.e. https://example.com",
	"username": "username, usually should be email address",
	"password": "GithubSinger api access token"
}
*/
func PatchConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	connection := &models.GithubSingerConnection{}
	err := connectionHelper.Patch(connection, input)
	if err != nil {
		return nil, err
	}
	return &core.ApiResourceOutput{Body: connection}, nil
}

/*
DELETE /plugins/GithubSinger/connections/:connectionId
*/
func DeleteConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	connection := &models.GithubSingerConnection{}
	err := connectionHelper.First(connection, input.Params)
	if err != nil {
		return nil, err
	}
	err = connectionHelper.Delete(connection)
	return &core.ApiResourceOutput{Body: connection}, err
}

/*
GET /plugins/GithubSinger/connections
*/
func ListConnections(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	var connections []models.GithubSingerConnection
	err := connectionHelper.List(&connections)
	if err != nil {
		return nil, err
	}
	return &core.ApiResourceOutput{Body: connections, Status: http.StatusOK}, nil
}

//TODO Please modify the folowing code to adapt to your plugin
/*
GET /plugins/GithubSinger/connections/:connectionId
{
	"name": "GithubSinger data connection name",
	"endpoint": "GithubSinger api endpoint, i.e. https://merico.atlassian.net/rest",
	"username": "username, usually should be email address",
	"password": "GithubSinger api access token"
}
*/
func GetConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	connection := &models.GithubSingerConnection{}
	err := connectionHelper.First(connection, input.Params)
	return &core.ApiResourceOutput{Body: connection}, err
}