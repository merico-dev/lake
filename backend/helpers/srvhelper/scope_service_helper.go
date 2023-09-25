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

package srvhelper

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/models/domainlayer/domaininfo"
	"github.com/apache/incubator-devlake/core/plugin"
)

type ScopePagination struct {
	Pagination
	ConnectionId uint64 `json:"connectionId" mapstructure:"connectionId" validate:"required"`
	Blueprnts    bool   `json:"blueprints" mapstructure:"blueprints"`
}

type ScopeResItem[S plugin.ToolLayerScope] struct {
	Scope      S                   `json:"scope"`
	Blueprints []*models.Blueprint `json:"blueprints"`
}

type ScopeSrvHelper[C plugin.ToolLayerConnection, S plugin.ToolLayerScope, SC plugin.ToolLayerScopeConfig] struct {
	*ModelSrvHelper[S]
	pluginName    string
	searchColumns []string
}

// NewScopeSrvHelper creates a ScopeDalHelper for scope management
func NewScopeSrvHelper[
	C plugin.ToolLayerConnection,
	S plugin.ToolLayerScope,
	SC plugin.ToolLayerScopeConfig,
](
	basicRes context.BasicRes,
	pluginName string,
	searchColumns []string,
) *ScopeSrvHelper[C, S, SC] {
	return &ScopeSrvHelper[C, S, SC]{
		ModelSrvHelper: NewModelSrvHelper[S](basicRes),
		pluginName:     pluginName,
		searchColumns:  searchColumns,
	}
}

func (self *ScopeSrvHelper[C, S, SC]) Validate(scope *S) errors.Error {
	connectionId := (*scope).ScopeConnectionId()
	connectionCount := errors.Must1(self.db.Count(dal.From(new(SC)), dal.Where("id = ?", connectionId)))
	if connectionCount == 0 {
		return errors.BadInput.New("connectionId is invalid")
	}
	scopeConfigId := (*scope).ScopeScopeConfigId()
	scopeConfigCount := errors.Must1(self.db.Count(dal.From(new(SC)), dal.Where("id = ?", scopeConfigId)))
	if scopeConfigCount == 0 {
		return errors.BadInput.New("scopeConfigId is invalid")
	}
	return nil
}

func (self *ScopeSrvHelper[C, S, SC]) GetPage(pagination *ScopePagination) ([]*ScopeResItem[S], int64, errors.Error) {
	if pagination.ConnectionId < 1 {
		return nil, 0, errors.BadInput.New("connectionId is required")
	}
	scopes, count, err := self.ModelSrvHelper.GetPage(
		&pagination.Pagination,
		dal.Where("connection_id = ?", pagination.ConnectionId),
	)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*ScopeResItem[S], len(scopes))
	if pagination.Blueprnts {
		for i, s := range scopes {
			// load blueprints
			scope := (*s)
			blueprints := self.getAllBlueprinsByScope(scope.ScopeConnectionId(), scope.ScopeId())
			resItem := &ScopeResItem[S]{
				Scope:      scope,
				Blueprints: blueprints,
			}
			data[i] = resItem
		}
	}
	return data, count, nil
}

func (self *ScopeSrvHelper[C, S, SC]) Delete(scope *S, dataOnly bool) (refs *DsRefs, err errors.Error) {
	err = self.ModelSrvHelper.NoRunningPipeline(func(tx dal.Transaction) errors.Error {
		s := (*scope)
		// check referencing blueprints
		if !dataOnly {
			refs, err = toDsRefs(self.getAllBlueprinsByScope(s.ScopeConnectionId(), s.ScopeId()))
			if err != nil {
				return err
			}
			errors.Must(tx.Delete(scope))
		}
		// delete data
		self.deleteScopeData(s, tx)
		return nil
	})
	return
}

func (self *ScopeSrvHelper[C, S, SC]) getAllBlueprinsByScope(connectionId uint64, scopeId string) []*models.Blueprint {
	blueprints := make([]*models.Blueprint, 0)
	errors.Must(self.db.All(
		&blueprints,
		dal.From("_devlake_blueprints bp"),
		dal.Join("JOIN _devlake_blueprint_scopes sc ON sc.blueprint_id = bp.id"),
		dal.Where(
			"mode = ? AND sc.connection_id = ? AND sc.plugin_name = ? AND sc.scope_id = ?",
			"NORMAL",
			connectionId,
			self.pluginName,
			scopeId,
		),
	))
	return blueprints
}

func (self *ScopeSrvHelper[C, S, SC]) deleteScopeData(scope plugin.ToolLayerScope, tx dal.Transaction) {
	rawDataParams := plugin.MarshalScopeParams(scope.ScopeParams())
	generateWhereClause := func(table string) (string, []any) {
		var where string
		var params []interface{}
		if strings.HasPrefix(table, "_raw_") {
			// raw table: should check connection and scope
			where = "params = ?"
			params = []interface{}{rawDataParams}
		} else if strings.HasPrefix(table, "_tool_") {
			// tool layer table: should check connection and scope
			where = "_raw_data_params = ?"
			params = []interface{}{rawDataParams}
		} else {
			// framework tables: should check plugin, connection and scope
			if table == (models.CollectorLatestState{}.TableName()) {
				// diff sync state
				where = "raw_data_table LIKE ? AND raw_data_params = ?"
			} else {
				// domain layer table
				where = "_raw_data_table LIKE ? AND _raw_data_params = ?"
			}
			rawDataTablePrefix := fmt.Sprintf("_raw_%s%%", self.pluginName)
			params = []interface{}{rawDataTablePrefix, rawDataParams}
		}
		return where, params
	}
	tables := errors.Must1(self.getAffectedTables())
	for _, table := range tables {
		where, params := generateWhereClause(table)
		self.log.Info("deleting data from table %s with WHERE \"%s\" and params: \"%v\"", table, where, params)
		sql := fmt.Sprintf("DELETE FROM %s WHERE %s", table, where)
		errors.Must(tx.Exec(sql, params...))
	}
}

func (self *ScopeSrvHelper[C, S, SC]) getAffectedTables() ([]string, errors.Error) {
	var tables []string
	meta, err := plugin.GetPlugin(self.pluginName)
	if err != nil {
		return nil, err
	}
	if pluginModel, ok := meta.(plugin.PluginModel); !ok {
		panic(errors.Default.New(fmt.Sprintf("plugin \"%s\" does not implement listing its tables", self.pluginName)))
	} else {
		// Unfortunately, can't cache the tables because Python creates some tables on a per-demand basis, so such a cache would possibly get outdated.
		// It's a rare scenario in practice, but might as well play it safe and sacrifice some performance here
		var allTables []string
		if allTables, err = self.db.AllTables(); err != nil {
			return nil, err
		}
		// collect raw tables
		for _, table := range allTables {
			if strings.HasPrefix(table, "_raw_"+self.pluginName) {
				tables = append(tables, table)
			}
		}
		// collect tool tables
		toolModels := pluginModel.GetTablesInfo()
		for _, toolModel := range toolModels {
			if !isScopeModel(toolModel) && hasField(toolModel, "RawDataParams") {
				tables = append(tables, toolModel.TableName())
			}
		}
		// collect domain tables
		for _, domainModel := range domaininfo.GetDomainTablesInfo() {
			// we only care about tables with RawOrigin
			ok = hasField(domainModel, "RawDataParams")
			if ok {
				tables = append(tables, domainModel.TableName())
			}
		}
		// additional tables
		tables = append(tables, models.CollectorLatestState{}.TableName())
	}
	self.log.Debug("Discovered %d tables used by plugin \"%s\": %v", len(tables), self.pluginName, tables)
	return tables, nil
}

// TODO: sort out the follow functions
func isScopeModel(obj dal.Tabler) bool {
	_, ok := obj.(plugin.ToolLayerScope)
	return ok
}

func hasField(obj any, fieldName string) bool {
	obj = models.UnwrapObject(obj)
	_, ok := reflectType(obj).FieldByName(fieldName)
	return ok
}

func reflectType(obj any) reflect.Type {
	obj = models.UnwrapObject(obj)
	typ := reflect.TypeOf(obj)
	kind := typ.Kind()
	for kind == reflect.Ptr {
		typ = typ.Elem()
		kind = typ.Kind()
	}
	return typ
}
