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
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type (
	// ScopeApiHelper is used to write the CURD of scopes
	ScopeApiHelper[Conn any, Scope any, Tr any] struct {
		*GenericScopeApiHelper[Conn, Scope, Tr]
	}
	ScopeRes[T any] struct {
		Scope                  T                   `mapstructure:",squash"`
		TransformationRuleName string              `mapstructure:"transformationRuleName,omitempty"`
		Blueprints             []*models.Blueprint `mapstructure:"blueprints,omitempty"`
	}
	ScopeReq[T any] struct {
		Data []*T `json:"data"`
	}
)

// Kept for backward compatibility. Use NewScopeHelper2 instead until we do a mass refactor
func NewScopeHelper[Conn any, Scope any, Tr any](
	basicRes context.BasicRes,
	vld *validator.Validate,
	connHelper *ConnectionApiHelper,
) *ScopeApiHelper[Conn, Scope, Tr] {
	reflectionParams := ReflectionParameters{}
	return NewScopeHelper2[Conn, Scope, Tr](
		basicRes,
		vld,
		connHelper,
		NewScopeDatabaseHelperImpl[Conn, Scope, Tr](basicRes, connHelper, &reflectionParams),
		&reflectionParams,
		nil,
	)
}

// NewScopeHelper creates a ScopeHelper for scopes management
func NewScopeHelper2[Conn any, Scope any, Tr any](
	basicRes context.BasicRes,
	vld *validator.Validate,
	connHelper *ConnectionApiHelper,
	dbHelper ScopeDatabaseHelper[Conn, Scope, Tr],
	params *ReflectionParameters,
	opts *ScopeHelperOptions,
) *ScopeApiHelper[Conn, Scope, Tr] {
	return &ScopeApiHelper[Conn, Scope, Tr]{
		NewGenericScopeHelper[Conn, Scope, Tr](
			basicRes, vld, connHelper, dbHelper, params, opts),
	}
}

// Put saves the given scopes to the database. It expects a slice of struct pointers
// as the scopes argument. It also expects a fieldName argument, which is used to extract
// the connection ID from the input.Params map.
func (c *ScopeApiHelper[Conn, Scope, Tr]) Put(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	var req struct {
		Data []*Scope `json:"data"`
	}
	err := errors.Convert(DecodeMapStruct(input.Body, &req, true))
	if err != nil {
		return nil, errors.BadInput.Wrap(err, "decoding scope error")
	}
	// Extract the connection ID from the input.Params map
	apiScopes, err := c.PutScopes(input, req.Data)
	if err != nil {
		return nil, err
	}
	return &plugin.ApiResourceOutput{Body: apiScopes, Status: http.StatusOK}, nil
}

// TODO remove fieldName param in the future and adjust plugins to use reflection params on init
func (c *ScopeApiHelper[Conn, Scope, Tr]) Update(input *plugin.ApiResourceInput, fieldName string) (*plugin.ApiResourceOutput, errors.Error) {
	if fieldName != "" {
		c.reflectionParams.ScopeIdColumnName = fieldName //for backward compatibility
	}
	apiScope, err := c.GenericScopeApiHelper.UpdateScope(input)
	if err != nil {
		return nil, err
	}
	return &plugin.ApiResourceOutput{Body: apiScope, Status: http.StatusOK}, nil
}

func (c *ScopeApiHelper[Conn, Scope, Tr]) GetScopeList(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	scopes, err := c.GetScopes(input)
	if err != nil {
		return nil, err
	}
	return &plugin.ApiResourceOutput{Body: scopes, Status: http.StatusOK}, nil
}

// TODO remove fieldName param in the future and adjust plugins to use reflection params on init
func (c *ScopeApiHelper[Conn, Scope, Tr]) GetScope(input *plugin.ApiResourceInput, fieldName string) (*plugin.ApiResourceOutput, errors.Error) {
	if fieldName != "" {
		c.reflectionParams.ScopeIdColumnName = fieldName //for backward compatibility
	}
	scope, err := c.GenericScopeApiHelper.GetScope(input)
	if err != nil {
		return nil, err
	}
	return &plugin.ApiResourceOutput{Body: scope, Status: http.StatusOK}, nil
}

func (c *ScopeApiHelper[Conn, Scope, Tr]) Delete(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	bps, err := c.DeleteScope(input)
	if err != nil {
		return nil, err
	}
	return &plugin.ApiResourceOutput{Body: bps, Status: http.StatusOK}, nil
}
