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
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/plugins/zentao/models"
	"github.com/mitchellh/mapstructure"
	"net/http"
)

//TODO Please modify the following code to fit your needs
func TestConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	// process input
	var params models.TestConnectionRequest
	err := mapstructure.Decode(input.Body, &params)
	if err != nil {
		return nil, errors.BadInput.Wrap(err, "could not decode request parameters", errors.AsUserMessage())
	}
	err = vld.Struct(params)
	if err != nil {
		return nil, errors.BadInput.Wrap(err, "could not validate request parameters", errors.AsUserMessage())
	}

	authApiClient, err := helper.NewApiClient(context.TODO(), params.Endpoint, nil, 0, params.Proxy, basicRes)
	if err != nil {
		return nil, err
	}

	// request for access token
	tokenReqBody := &models.ApiAccessTokenRequest{
		Account:  params.Username,
		Password: params.Password,
	}
	tokenRes, err := authApiClient.Post("/tokens", nil, tokenReqBody, nil)
	if err != nil {
		return nil, err
	}
	tokenResBody := &models.ApiAccessTokenResponse{}
	err = helper.UnmarshalResponse(tokenRes, tokenResBody)
	if err != nil {
		return nil, err
	}
	if tokenResBody.Token == "" {
		return nil, errors.Default.New("failed to request access token")
	}

	// output
	return nil, nil
}

//TODO Please modify the folowing code to adapt to your plugin
/*
POST /plugins/Zentao/connections
{
	"name": "Zentao data connection name",
	"endpoint": "Zentao api endpoint, i.e. https://example.com",
	"username": "username, usually should be email address",
	"password": "Zentao api access token"
}
*/
func PostConnections(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	// update from request and save to database
	connection := &models.ZentaoConnection{}
	err := connectionHelper.Create(connection, input)
	if err != nil {
		return nil, err
	}
	return &core.ApiResourceOutput{Body: connection, Status: http.StatusOK}, nil
}

//TODO Please modify the folowing code to adapt to your plugin
/*
PATCH /plugins/Zentao/connections/:connectionId
{
	"name": "Zentao data connection name",
	"endpoint": "Zentao api endpoint, i.e. https://example.com",
	"username": "username, usually should be email address",
	"password": "Zentao api access token"
}
*/
func PatchConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	connection := &models.ZentaoConnection{}
	err := connectionHelper.Patch(connection, input)
	if err != nil {
		return nil, err
	}
	return &core.ApiResourceOutput{Body: connection}, nil
}

/*
DELETE /plugins/Zentao/connections/:connectionId
*/
func DeleteConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	connection := &models.ZentaoConnection{}
	err := connectionHelper.First(connection, input.Params)
	if err != nil {
		return nil, err
	}
	err = connectionHelper.Delete(connection)
	return &core.ApiResourceOutput{Body: connection}, err
}

/*
GET /plugins/Zentao/connections
*/
func ListConnections(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	var connections []models.ZentaoConnection
	err := connectionHelper.List(&connections)
	if err != nil {
		return nil, err
	}
	return &core.ApiResourceOutput{Body: connections, Status: http.StatusOK}, nil
}

//TODO Please modify the folowing code to adapt to your plugin
/*
GET /plugins/Zentao/connections/:connectionId
{
	"name": "Zentao data connection name",
	"endpoint": "Zentao api endpoint, i.e. https://merico.atlassian.net/rest",
	"username": "username, usually should be email address",
	"password": "Zentao api access token"
}
*/
func GetConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	connection := &models.ZentaoConnection{}
	err := connectionHelper.First(connection, input.Params)
	return &core.ApiResourceOutput{Body: connection}, err
}
