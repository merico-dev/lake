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
	"fmt"
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/server/services/remote/models"
	"reflect"
	"time"
)

type ScopeDatabaseHelperImpl struct {
	pa         *pluginAPI
	db         dal.Dal
	params     *api.ReflectionParameters
	connHelper *api.ConnectionApiHelper
}

func NewScopeDatabaseHelperImpl(pa *pluginAPI, basicRes context.BasicRes, params *api.ReflectionParameters) *ScopeDatabaseHelperImpl {
	return &ScopeDatabaseHelperImpl{
		pa:         pa,
		db:         basicRes.GetDal(),
		params:     params,
		connHelper: connectionHelper,
	}
}

func (s *ScopeDatabaseHelperImpl) VerifyConnection(connectionId uint64) errors.Error {
	conn := s.pa.connType.New()
	err := s.connHelper.FirstById(conn, connectionId)
	if err != nil {
		if s.db.IsErrorNotFound(err) {
			return errors.BadInput.New("Invalid Connection Id")
		}
		return err
	}
	return nil
}

func (s *ScopeDatabaseHelperImpl) SaveScope(scopes []*any) errors.Error {
	now := time.Now()
	return s.save(scopes, &now, &now)
}

func (s *ScopeDatabaseHelperImpl) UpdateScope(connectionId uint64, scopeId string, scope any) errors.Error {
	// Update API on Gorm doesn't work with dynamic models. Need to do delete + create instead, unfortunately.
	if err := s.DeleteScope(connectionId, scopeId); err != nil {
		if !s.db.IsErrorNotFound(err) {
			return err
		}
	}
	now := time.Now()
	return s.save([]*any{&scope}, nil, &now)
}

func (s *ScopeDatabaseHelperImpl) GetScope(connectionId uint64, scopeId string) (any, errors.Error) {
	query := dal.Where(fmt.Sprintf("connection_id = ? AND %s = ?", s.params.ScopeIdColumnName), connectionId, scopeId)
	scope := s.pa.scopeType.New()
	err := api.CallDB(s.db.First, scope, query)
	if err != nil {
		return nil, errors.Default.Wrap(err, "could not get scope")
	}
	return scope.Unwrap(), nil
}

func (s *ScopeDatabaseHelperImpl) ListScopes(input *plugin.ApiResourceInput, connectionId uint64) ([]*any, errors.Error) {
	limit, offset := api.GetLimitOffset(input.Query, "pageSize", "page")
	scopes := s.pa.scopeType.NewSlice()
	err := api.CallDB(s.db.All, scopes, dal.Where("connection_id = ?", connectionId), dal.Limit(limit), dal.Offset(offset))
	if err != nil {
		return nil, err
	}
	var result []*any
	for _, scope := range scopes.UnwrapSlice() {
		scope := scope
		result = append(result, &scope)
	}
	return result, nil
}

func (s *ScopeDatabaseHelperImpl) DeleteScope(connectionId uint64, scopeId string) errors.Error {
	rawScope := s.pa.scopeType.New()
	return api.CallDB(s.db.Delete, rawScope, dal.Where("connection_id = ? AND id = ?", connectionId, scopeId))
}

func (s *ScopeDatabaseHelperImpl) GetTransformationRule(ruleId uint64) (any, errors.Error) {
	rule := s.pa.txRuleType.New()
	err := api.CallDB(s.db.First, rule, dal.Where("id = ?", ruleId))
	if err != nil {
		return rule, err
	}
	return rule.Unwrap(), nil
}

func (s *ScopeDatabaseHelperImpl) ListTransformationRules(ruleIds []uint64) ([]*any, errors.Error) {
	rules := s.pa.txRuleType.NewSlice()
	err := api.CallDB(s.db.All, rules, dal.Where("id IN (?)", ruleIds))
	if err != nil {
		return nil, err
	}
	var result []*any
	for _, rule := range rules.UnwrapSlice() {
		rule := rule
		result = append(result, &rule)
	}
	return result, nil
}

func (s *ScopeDatabaseHelperImpl) save(scopes []*any, createdAt *time.Time, updatedAt *time.Time) errors.Error {
	var targets []map[string]any
	for _, x := range scopes {
		ifc := reflect.ValueOf(*x).Elem().Interface()
		m, err := models.ToDatabaseMap(s.pa.scopeType.TableName(), ifc, createdAt, updatedAt)
		if err != nil {
			return err
		}
		targets = append(targets, m)
	}
	err := api.CallDB(s.db.Create, &targets, dal.From(s.pa.scopeType.TableName()))
	if err != nil {
		return errors.Default.Wrap(err, "could not save scope")
	}
	return nil
}

var _ api.ScopeDatabaseHelper[any, any, any] = &ScopeDatabaseHelperImpl{}
