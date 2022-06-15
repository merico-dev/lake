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

	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

var vld = validator.New()

/*
POST /plugins/gitee/test
*/
func TestConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	// decode
	var err error
	var connection models.TestConnectionRequest
	err = mapstructure.Decode(input.Body, &connection)
	if err != nil {
		return nil, err
	}
	// validate
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
	query := make(url.Values)
	query["access_token"] = []string{connection.Auth}

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
PATCH /plugins/gitee/connections/:connectionId
*/
func PatchConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	v := config.GetConfig()
	connection := &models.GiteeConnection{}
	err := helper.EncodeStruct(v, connection, "env")
	if err != nil {
		return nil, err
	}
	// update from request and save to .env
	err = helper.DecodeStruct(v, connection, input.Body, "env")
	if err != nil {
		return nil, err
	}
	err = config.WriteConfig(v)
	if err != nil {
		return nil, err
	}
	response := models.GiteeResponse{
		GiteeConnection: *connection,
		Name:            "Gitee",
		ID:              1,
	}
	return &core.ApiResourceOutput{Body: response, Status: http.StatusOK}, nil
}

/*
GET /plugins/gitee/connections
*/
func ListConnections(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	// RETURN ONLY 1 SOURCE (FROM ENV) until multi-connection is developed.
	v := config.GetConfig()
	connection := &models.GiteeConnection{}

	err := helper.EncodeStruct(v, connection, "env")
	if err != nil {
		return nil, err
	}
	response := models.GiteeResponse{
		GiteeConnection: *connection,
		Name:            "Gitee",
		ID:              1,
	}

	return &core.ApiResourceOutput{Body: []models.GiteeResponse{response}}, nil
}

/*
GET /plugins/gitee/connections/:connectionId
*/
func GetConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	//  RETURN ONLY 1 SOURCE FROM ENV (Ignore ID until multi-connection is developed.)
	v := config.GetConfig()
	connection := &models.GiteeConnection{}
	err := helper.EncodeStruct(v, connection, "env")
	if err != nil {
		return nil, err
	}
	response := &models.GiteeResponse{
		GiteeConnection: *connection,
		Name:            "Gitee",
		ID:              1,
	}
	return &core.ApiResourceOutput{Body: response}, nil
}
