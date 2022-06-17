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
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/apache/incubator-devlake/plugins/gitee/models"

	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/mitchellh/mapstructure"

	"github.com/apache/incubator-devlake/plugins/core"
)

/*
POST /plugins/gitee/test
*/
func TestConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	var connection models.TestConnectionRequest
	err := mapstructure.Decode(input.Body, &connection)
	if err != nil {
		return nil, err
	}
	err = vld.Struct(connection)
	if err != nil {
		return nil, err
	}
	// test connection
	apiClient, err := helper.NewApiClient(
		connection.Endpoint,
		nil,
		3*time.Second,
		connection.Proxy,
		nil,
	)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("access_token", connection.Token)

	res, err := apiClient.Get("user", query, nil)
	if err != nil {
		return nil, err
	}
	resBody := &models.ApiUserResponse{}
	err = helper.UnmarshalResponse(res, resBody)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	return nil, nil
}

/*
POST /plugins/gitee/connections
*/
func PostConnections(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	connection := &models.GiteeConnection{}
	err := connectionHelper.Create(connection, input)
	if err != nil {
		return nil, err
	}
	return &core.ApiResourceOutput{Body: connection, Status: http.StatusOK}, nil
}

/*
PATCH /plugins/gitee/connections/:connectionId
*/
func PatchConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	connection := &models.GiteeConnection{}
	err := connectionHelper.Patch(connection, input)
	if err != nil {
		return nil, err
	}
	return &core.ApiResourceOutput{Body: connection, Status: http.StatusOK}, nil
}

/*
DELETE /plugins/gitee/connections/:connectionId
*/
func DeleteConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	connection := &models.GiteeConnection{}
	err := connectionHelper.First(connection, input.Params)
	if err != nil {
		return nil, err
	}
	err = connectionHelper.Delete(connection)
	return &core.ApiResourceOutput{Body: connection}, err
}

/*
GET /plugins/gitee/connections
*/
func ListConnections(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	var connections []models.GiteeConnection
	err := connectionHelper.List(&connections)
	if err != nil {
		return nil, err
	}

	return &core.ApiResourceOutput{Body: connections}, nil
}

/*
GET /plugins/gitee/connections/:connectionId
*/
func GetConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	connection := &models.GiteeConnection{}
	err := connectionHelper.First(connection, input.Params)
	if err != nil {
		return nil, err
	}
	return &core.ApiResourceOutput{Body: connection}, err
}
