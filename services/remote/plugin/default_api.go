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

package plugin

import (
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/models"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
	"net/http"
)

type ConnectionAPI struct {
	connType *models.DynamicTabler
	helper   *helper.ConnectionApiHelper
}

func GetDefaultAPI(connType *models.DynamicTabler, helper *helper.ConnectionApiHelper) map[string]map[string]core.ApiResourceHandler {
	api := &ConnectionAPI{
		connType: connType,
		helper:   helper,
	}
	return map[string]map[string]core.ApiResourceHandler{
		"test": {
			"POST": api.TestConnection,
		},
		"connections": {
			"POST": api.PostConnections,
			"GET":  api.ListConnections,
		},
		"connections/:connectionId": {
			"GET":    api.GetConnection,
			"PATCH":  api.PatchConnection,
			"DELETE": api.DeleteConnection,
		},
	}
}

func (c *ConnectionAPI) TestConnection(_ *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	return nil, errors.Default.New("endpoint not implemented")
}

func (c *ConnectionAPI) PostConnections(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	connection := c.connType.New()
	err := c.helper.Create(connection, input)
	if err != nil {
		return nil, err
	}
	conn := connection.Unwrap()
	return &core.ApiResourceOutput{Body: conn, Status: http.StatusOK}, nil
}

func (c *ConnectionAPI) ListConnections(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	connections := c.connType.NewSlice()
	err := c.helper.List(connections)
	if err != nil {
		return nil, err
	}
	conns := connections.Unwrap()
	return &core.ApiResourceOutput{Body: conns}, nil
}

func (c *ConnectionAPI) GetConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	connection := c.connType.New()
	err := c.helper.First(connection, input.Params)
	if err != nil {
		return nil, err
	}
	conn := connection.Unwrap()
	return &core.ApiResourceOutput{Body: conn}, nil
}

func (c *ConnectionAPI) PatchConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	connection := c.connType.New()
	err := c.helper.Patch(connection, input)
	if err != nil {
		return nil, err
	}
	conn := connection.Unwrap()
	return &core.ApiResourceOutput{Body: conn, Status: http.StatusOK}, nil
}

func (c *ConnectionAPI) DeleteConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	connection := c.connType.New()
	err := c.helper.First(connection, input.Params)
	if err != nil {
		return nil, err
	}
	err = c.helper.Delete(connection)
	conn := connection.Unwrap()
	return &core.ApiResourceOutput{Body: conn}, err
}
