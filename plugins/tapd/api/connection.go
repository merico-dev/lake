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
	"net/http"
	"time"

	"github.com/apache/incubator-devlake/plugins/tapd/models"

	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/mitchellh/mapstructure"

	"github.com/apache/incubator-devlake/plugins/core"
)

// @Summary test tapd connection
// @Description Test Tapd Connection
// @Tags plugins/tapd
// @Param body body models.TestConnectionRequest true "json body"
// @Success 200  {object} shared.ApiBody "Success"
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /plugins/tapd/test [POST]
func TestConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	// process input
	var connection models.TestConnectionRequest
	err := mapstructure.Decode(input.Body, &connection)
	if err != nil {
		return nil, err
	}
	err = vld.Struct(connection)
	if err != nil {
		return nil, err
	}

	// verify multiple token in parallel
	// PLEASE NOTE: This works because GitHub API Client rotates tokens on each request
	apiClient, err := helper.NewApiClient(
		context.TODO(),
		connection.Endpoint,
		map[string]string{
			"Authorization": fmt.Sprintf("Basic %s", connection.GetEncodedToken()),
		},
		10*time.Second,
		connection.Proxy,
		basicRes,
	)
	if err != nil {
		return nil, fmt.Errorf("verify token failed for %s %w", connection.Username, err)
	}
	res, err := apiClient.Get("/quickstart/testauth", nil, nil)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("verify token failed for %s", connection.Username)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	// output
	return nil, nil
}

// @Summary create tapd connection
// @Description Create Tapd connection
// @Tags plugins/tapd
// @Param body body models.TapdConnection true "json body"
// @Success 200  {object} models.TapdConnection "Success"
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /plugins/tapd/connections [POST]
func PostConnections(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	// create a new connections
	connection := &models.TapdConnection{}

	// update from request and save to database
	//err := refreshAndSaveTapdConnection(tapdConnection, input.Body)
	err := connectionHelper.Create(connection, input)
	if err != nil {
		return nil, err
	}

	return &core.ApiResourceOutput{Body: connection, Status: http.StatusOK}, nil
}

// @Summary patch tapd connection
// @Description Patch Tapd connection
// @Tags plugins/tapd
// @Param body body models.TapdConnection true "json body"
// @Success 200  {object} models.TapdConnection "Success"
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /plugins/tapd/connections/{connectionId} [PATCH]
func PatchConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	connection := &models.TapdConnection{}
	err := connectionHelper.Patch(connection, input)
	if err != nil {
		return nil, err
	}

	return &core.ApiResourceOutput{Body: connection}, nil
}

// @Summary delete a tapd connection
// @Description Delete a Tapd connection
// @Tags plugins/tapd
// @Success 200  {object} models.TapdConnection "Success"
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /plugins/tapd/connections/{connectionId} [DELETE]
func DeleteConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	connection := &models.TapdConnection{}
	err := connectionHelper.First(connection, input.Params)
	if err != nil {
		return nil, err
	}
	err = connectionHelper.Delete(connection)
	return &core.ApiResourceOutput{Body: connection}, err
}

// @Summary get all tapd connections
// @Description Get all Tapd connections
// @Tags plugins/tapd
// @Success 200  {object} models.TapdConnection "Success"
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /plugins/tapd/connections [GET]
func ListConnections(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	var connections []models.TapdConnection
	err := connectionHelper.List(&connections)
	if err != nil {
		return nil, err
	}

	return &core.ApiResourceOutput{Body: connections, Status: http.StatusOK}, nil
}

// @Summary get tapd connection detail
// @Description Get Tapd connection detail
// @Tags plugins/tapd
// @Success 200  {object} models.TapdConnection "Success"
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internel Error"
// @Router /plugins/tapd/connections/{connectionId} [GET]
func GetConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	connection := &models.TapdConnection{}
	err := connectionHelper.First(connection, input.Params)
	if err != nil {
		return nil, err
	}
	return &core.ApiResourceOutput{Body: connection}, nil
}
