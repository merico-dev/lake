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
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
)

type ScopeDatabaseHelper[Conn any, Scope any, Tr any] interface {
	VerifyConnection(connectionId uint64) errors.Error
	SaveScope(scopes []*Scope) errors.Error
	UpdateScope(connectionId uint64, scopeId string, scope Scope) errors.Error
	GetScope(connectionId uint64, scopeId string) (Scope, errors.Error)
	ListScopes(input *plugin.ApiResourceInput, connectionId uint64) ([]*Scope, errors.Error)
	DeleteScope(connectionId uint64, scopeId string) errors.Error
	GetTransformationRule(ruleId uint64) (Tr, errors.Error)
	ListTransformationRules(ruleIds []uint64) ([]*Tr, errors.Error)
}

type ScopeDatabaseHelperImpl[Conn any, Scope any, Tr any] struct {
	ScopeDatabaseHelper[Conn, Scope, Tr]
	db         dal.Dal
	connHelper *ConnectionApiHelper
	params     *ReflectionParameters
}

func NewScopeDatabaseHelperImpl[Conn any, Scope any, Tr any](
	basicRes context.BasicRes, connHelper *ConnectionApiHelper, params *ReflectionParameters) *ScopeDatabaseHelperImpl[Conn, Scope, Tr] {
	return &ScopeDatabaseHelperImpl[Conn, Scope, Tr]{
		db:         basicRes.GetDal(),
		connHelper: connHelper,
		params:     params,
	}
}

func (s *ScopeDatabaseHelperImpl[Conn, Scope, Tr]) VerifyConnection(connectionId uint64) errors.Error {
	var conn Conn
	err := s.connHelper.FirstById(&conn, connectionId)
	if err != nil {
		if s.db.IsErrorNotFound(err) {
			return errors.BadInput.New("Invalid Connection Id")
		}
		return err
	}
	return nil
}

func (s *ScopeDatabaseHelperImpl[Conn, Scope, Tr]) SaveScope(scopes []*Scope) errors.Error {
	err := s.db.CreateOrUpdate(&scopes)
	if err != nil {
		if s.db.IsDuplicationError(err) {
			return errors.BadInput.New("the scope already exists")
		}
		return err
	}
	return nil
}

func (s *ScopeDatabaseHelperImpl[Conn, Scope, Tr]) UpdateScope(connectionId uint64, scopeId string, scope Scope) errors.Error {
	return s.db.Update(&scope)
}

func (s *ScopeDatabaseHelperImpl[Conn, Scope, Tr]) GetScope(connectionId uint64, scopeId string) (Scope, errors.Error) {
	query := dal.Where(fmt.Sprintf("connection_id = ? AND %s = ?", s.params.ScopeIdColumnName), connectionId, scopeId)
	var scope Scope
	err := s.db.First(&scope, query)
	return scope, err
}

func (s *ScopeDatabaseHelperImpl[Conn, Scope, Tr]) ListScopes(input *plugin.ApiResourceInput, connectionId uint64) ([]*Scope, errors.Error) {
	limit, offset := GetLimitOffset(input.Query, "pageSize", "page")
	var scopes []*Scope
	err := s.db.All(&scopes, dal.Where("connection_id = ?", connectionId), dal.Limit(limit), dal.Offset(offset))
	return scopes, err
}

func (s *ScopeDatabaseHelperImpl[Conn, Scope, Tr]) DeleteScope(connectionId uint64, scopeId string) errors.Error {
	scope := new(Scope)
	err := s.db.Delete(&scope, dal.Where(fmt.Sprintf("connection_id = ? AND %s = ?", s.params.ScopeIdColumnName),
		connectionId, scopeId))
	return err
}

func (s *ScopeDatabaseHelperImpl[Conn, Scope, Tr]) GetTransformationRule(ruleId uint64) (Tr, errors.Error) {
	var rule Tr
	err := s.db.First(&rule, dal.Where("id = ?", ruleId))
	return rule, err
}

func (s *ScopeDatabaseHelperImpl[Conn, Scope, Tr]) ListTransformationRules(ruleIds []uint64) ([]*Tr, errors.Error) {
	var rules []*Tr
	err := s.db.All(&rules, dal.Where("id IN (?)", ruleIds))
	return rules, err
}

var _ ScopeDatabaseHelper[any, any, any] = &ScopeDatabaseHelperImpl[any, any, any]{}
