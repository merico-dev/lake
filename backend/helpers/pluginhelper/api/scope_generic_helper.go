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
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/log"
	"github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/models/domainlayer/domaininfo"
	"github.com/apache/incubator-devlake/core/plugin"
	serviceHelper "github.com/apache/incubator-devlake/helpers/pluginhelper/services"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	tablesCache       []string // these cached vars can probably be moved somewhere more centralized later
	tablesCacheLoader = new(sync.Once)
)

type NoTransformation struct{}

type (
	GenericScopeApiHelper[Conn any, Scope any, Tr any] struct {
		log              log.Logger
		db               dal.Dal
		validator        *validator.Validate
		reflectionParams *ReflectionParameters
		dbHelper         ScopeDatabaseHelper[Conn, Scope, Tr]
		bpManager        *serviceHelper.BlueprintManager
		connHelper       *ConnectionApiHelper
		opts             *ScopeHelperOptions
	}
	ReflectionParameters struct {
		ScopeIdFieldName  string
		ScopeIdColumnName string
		RawScopeParamName string
	}
	ScopeHelperOptions struct {
		GetScopeParamValue func(db dal.Dal, scopeId string) (string, errors.Error)
	}
)

type (
	requestParams struct {
		connectionId uint64
		scopeId      string
		plugin       string
	}
	deleteRequestParams struct {
		requestParams
		deleteDataOnly bool
	}

	getRequestParams struct {
		requestParams
		loadBlueprints bool
	}
)

func NewGenericScopeHelper[Conn any, Scope any, Tr any](
	basicRes context.BasicRes,
	vld *validator.Validate,
	connHelper *ConnectionApiHelper,
	dbHelper ScopeDatabaseHelper[Conn, Scope, Tr],
	params *ReflectionParameters,
	opts *ScopeHelperOptions,
) *GenericScopeApiHelper[Conn, Scope, Tr] {
	if connHelper == nil {
		panic("nil connHelper")
	}
	if params == nil {
		panic("reflection params not provided")
	}
	if opts == nil {
		opts = &ScopeHelperOptions{}
	}
	tablesCacheLoader.Do(func() {
		var err errors.Error
		tablesCache, err = basicRes.GetDal().AllTables()
		if err != nil {
			panic(err)
		}
	})
	return &GenericScopeApiHelper[Conn, Scope, Tr]{
		log:              basicRes.GetLogger(),
		db:               basicRes.GetDal(),
		validator:        vld,
		reflectionParams: params,
		dbHelper:         dbHelper,
		bpManager:        serviceHelper.NewBlueprintManager(basicRes.GetDal()),
		connHelper:       connHelper,
		opts:             opts,
	}
}

func (c *GenericScopeApiHelper[Conn, Scope, Tr]) PutScopes(input *plugin.ApiResourceInput, scopes []*Scope) ([]*ScopeRes[Scope], errors.Error) {
	params := c.extractFromReqParam(input)
	if params.connectionId == 0 {
		return nil, errors.BadInput.New("invalid connectionId")
	}
	err := c.dbHelper.VerifyConnection(params.connectionId)
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("error verifying connection for connection ID %d", params.connectionId))
	}
	err = c.validatePrimaryKeys(scopes)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	for _, scope := range scopes {
		// Set the connection ID, CreatedDate, and UpdatedDate fields
		setScopeFields(scope, params.connectionId, &now, &now)
		err = VerifyScope(scope, c.validator)
		if err != nil {
			return nil, errors.Default.Wrap(err, "error verifying scope")
		}
	}
	// Save the scopes to the database
	if len(scopes) > 0 {
		err = c.dbHelper.SaveScope(scopes)
		if err != nil {
			return nil, errors.Default.Wrap(err, "error saving scope")
		}
	}
	apiScopes, err := c.addTransformationName(scopes...)
	if err != nil {
		return nil, errors.Default.Wrap(err, "error associating transformation to scope")
	}
	b, _ := json.Marshal(apiScopes[0].Scope)
	_ = b
	return apiScopes, nil
}

func (c *GenericScopeApiHelper[Conn, Scope, Tr]) UpdateScope(input *plugin.ApiResourceInput) (*ScopeRes[Scope], errors.Error) {
	params := c.extractFromReqParam(input)
	if params.connectionId == 0 {
		return nil, errors.BadInput.New("invalid connectionId")
	}
	if len(params.scopeId) == 0 {
		return nil, errors.BadInput.New("invalid scopeId")
	}
	err := c.dbHelper.VerifyConnection(params.connectionId)
	if err != nil {
		return nil, err
	}
	scope, err := c.dbHelper.GetScope(params.connectionId, params.scopeId)
	if err != nil {
		return nil, err
	}
	err = DecodeMapStruct(input.Body, &scope, false)
	if err != nil {
		return nil, errors.Default.Wrap(err, "patch scope error")
	}
	err = VerifyScope(&scope, c.validator)
	if err != nil {
		return nil, errors.Default.Wrap(err, "Invalid scope")
	}
	err = c.dbHelper.UpdateScope(params.connectionId, params.scopeId, scope)
	if err != nil {
		return nil, errors.Default.Wrap(err, "error on saving Scope")
	}
	scopeRes, err := c.addTransformationName(&scope)
	if err != nil {
		return nil, err
	}
	return scopeRes[0], nil
}

func (c *GenericScopeApiHelper[Conn, Scope, Tr]) GetScopes(input *plugin.ApiResourceInput) ([]*ScopeRes[Scope], errors.Error) {
	params := c.extractFromGetReqParam(input)
	if params.connectionId == 0 {
		return nil, errors.BadInput.New("invalid path params: \"connectionId\" not set")
	}
	err := c.dbHelper.VerifyConnection(params.connectionId)
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("error verifying connection for connection ID %d", params.connectionId))
	}
	scopes, err := c.dbHelper.ListScopes(input, params.connectionId)
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("error verifying connection for connection ID %d", params.connectionId))
	}
	apiScopes, err := c.addTransformationName(scopes...)
	if err != nil {
		return nil, errors.Default.Wrap(err, "error associating transformations with scopes")
	}
	if params.loadBlueprints {
		scopesById := c.mapByScopeId(apiScopes)
		var scopeIds []string
		for id := range scopesById {
			scopeIds = append(scopeIds, id)
		}
		blueprintMap, err := c.bpManager.GetBlueprintsByScopes(scopeIds...)
		if err != nil {
			return nil, errors.Default.Wrap(err, fmt.Sprintf("error getting blueprints for scopes from connection %d", params.connectionId))
		}
		apiScopes = nil
		for scopeId, scope := range scopesById {
			if bps, ok := blueprintMap[scopeId]; ok {
				scope.Blueprints = bps
				delete(blueprintMap, scopeId)
			}
			apiScopes = append(apiScopes, scope)
		}
		if len(blueprintMap) > 0 {
			var danglingIds []string
			for bpId := range blueprintMap {
				danglingIds = append(danglingIds, bpId)
			}
			c.log.Warn(nil, "The following dangling scopes were found: %v", danglingIds)
		}
	}
	return apiScopes, nil
}

func (c *GenericScopeApiHelper[Conn, Scope, Tr]) GetScope(input *plugin.ApiResourceInput) (*ScopeRes[Scope], errors.Error) {
	params := c.extractFromGetReqParam(input)
	if params == nil || params.connectionId == 0 {
		return nil, errors.BadInput.New("invalid path params: \"connectionId\" not set")
	}
	if len(params.scopeId) == 0 || params.scopeId == "0" {
		return nil, errors.BadInput.New("invalid path params: \"scopeId\" not set/invalid")
	}
	err := c.dbHelper.VerifyConnection(params.connectionId)
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("error verifying connection for connection ID %d", params.connectionId))
	}
	scope, err := c.dbHelper.GetScope(params.connectionId, params.scopeId)
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("error retrieving scope with scope ID %s", params.scopeId))
	}
	apiScopes, err := c.addTransformationName(&scope)
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("error associating transformation with scope %s", params.scopeId))
	}
	scopeRes := apiScopes[0]
	var blueprints []*models.Blueprint
	if params.loadBlueprints {
		blueprintMap, err := c.bpManager.GetBlueprintsByScopes(params.scopeId)
		if err != nil {
			return nil, errors.Default.Wrap(err, fmt.Sprintf("error getting blueprints for scope with scope ID %s", params.scopeId))
		}
		if len(blueprintMap) == 1 {
			blueprints = blueprintMap[params.scopeId]
		}
	}
	scopeRes.Blueprints = blueprints
	return scopeRes, nil
}

func (c *GenericScopeApiHelper[Conn, Scope, Tr]) DeleteScope(input *plugin.ApiResourceInput) ([]*models.Blueprint, errors.Error) {
	params := c.extractFromDeleteReqParam(input)
	if params == nil || params.connectionId == 0 {
		return nil, errors.BadInput.New("invalid path params: \"connectionId\" not set")
	}
	if len(params.scopeId) == 0 || params.scopeId == "0" {
		return nil, errors.BadInput.New("invalid path params: \"scopeId\" not set/invalid")
	}
	err := c.dbHelper.VerifyConnection(params.connectionId)
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("error verifying connection for connection ID %d", params.connectionId))
	}
	db := c.db
	blueprintsMap, err := c.bpManager.GetBlueprintsByScopes(params.scopeId)
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("error retrieving scope with scope ID %s", params.scopeId))
	}
	blueprints := blueprintsMap[params.scopeId]
	// find all tables for this plugin
	tables, err := getPluginTables(params.plugin)
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("error getting database tables managed by plugin %s", params.plugin))
	}
	// delete all the plugin records referencing this scope
	if c.reflectionParams.RawScopeParamName != "" {
		scopeParamValue := params.scopeId
		if c.opts.GetScopeParamValue != nil {
			scopeParamValue, err = c.opts.GetScopeParamValue(c.db, params.scopeId)
			if err != nil {
				return nil, errors.Default.Wrap(err, fmt.Sprintf("error extracting scope parameter name for scope %s", params.scopeId))
			}
		}
		for _, table := range tables {
			err = db.Exec(createDeleteQuery(table, c.reflectionParams.RawScopeParamName, scopeParamValue))
			if err != nil {
				return nil, errors.Default.Wrap(err, fmt.Sprintf("error deleting data bound to scope %s for plugin %s", params.scopeId, params.plugin))
			}
		}
	}
	var impactedBlueprints []*models.Blueprint
	if !params.deleteDataOnly {
		// Delete the scope itself
		err = c.dbHelper.DeleteScope(params.connectionId, params.scopeId)
		if err != nil {
			return nil, errors.Default.Wrap(err, fmt.Sprintf("error deleting scope %s", params.scopeId))
		}
		// update the blueprints (remove scope reference from them)
		for _, blueprint := range blueprints {
			settings, _ := blueprint.UnmarshalSettings()
			var changed bool
			err = settings.UpdateConnections(func(c *plugin.BlueprintConnectionV200) errors.Error {
				var retainedScopes []*plugin.BlueprintScopeV200
				for _, bpScope := range c.Scopes {
					if bpScope.Id == params.scopeId { // we'll be removing this one
						changed = true
					} else {
						retainedScopes = append(retainedScopes, bpScope)
					}
				}
				c.Scopes = retainedScopes
				return nil
			})
			if err != nil {
				return nil, errors.Default.Wrap(err, fmt.Sprintf("error removing scope %s from blueprint %d", params.scopeId, blueprint.ID))
			}
			if changed {
				err = blueprint.UpdateSettings(&settings)
				if err != nil {
					return nil, errors.Default.Wrap(err, fmt.Sprintf("error writing new settings into blueprint %s", blueprint.Name))
				}
				err = c.bpManager.SaveDbBlueprint(blueprint)
				if err != nil {
					return nil, errors.Default.Wrap(err, fmt.Sprintf("error saving the updated blueprint %s", blueprint.Name))
				}
				impactedBlueprints = append(impactedBlueprints, blueprint)
			}
		}
	}
	return impactedBlueprints, nil
}

func (c *GenericScopeApiHelper[Conn, Scope, Tr]) addTransformationName(scopes ...*Scope) ([]*ScopeRes[Scope], errors.Error) {
	var ruleIds []uint64
	for _, scope := range scopes {
		valueRepoRuleId := reflectField(scope, "TransformationRuleId")
		if !valueRepoRuleId.IsValid() {
			break
		}
		ruleId := reflectField(scope, "TransformationRuleId").Uint()
		if ruleId > 0 {
			ruleIds = append(ruleIds, ruleId)
		}
	}
	var rules []*Tr
	var err errors.Error
	if len(ruleIds) > 0 {
		rules, err = c.dbHelper.ListTransformationRules(ruleIds)
		if err != nil {
			return nil, err
		}
	}
	names := make(map[uint64]string)
	for _, rule := range rules {
		// Get the reflect.Value of the i-th struct pointer in the slice
		names[reflectField(rule, "ID").Uint()] = reflectField(rule, "Name").String()
	}
	apiScopes := make([]*ScopeRes[Scope], 0)
	for _, scope := range scopes {
		txRuleField := reflectField(scope, "TransformationRuleId")
		txRuleName := ""
		if txRuleField.IsValid() {
			txRuleName = names[txRuleField.Uint()]
		}
		apiScopes = append(apiScopes, &ScopeRes[Scope]{
			Scope:                  *scope,
			TransformationRuleName: txRuleName,
		})
	}
	return apiScopes, nil
}

func (c *GenericScopeApiHelper[Conn, Scope, Tr]) mapByScopeId(scopes []*ScopeRes[Scope]) map[string]*ScopeRes[Scope] {
	scopeMap := map[string]*ScopeRes[Scope]{}
	for _, scope := range scopes {
		scopeId := fmt.Sprintf("%v", reflectField(scope.Scope, c.reflectionParams.ScopeIdFieldName).Interface())
		scopeMap[scopeId] = scope
	}
	return scopeMap
}

func (c *GenericScopeApiHelper[Conn, Scope, Tr]) extractFromReqParam(input *plugin.ApiResourceInput) *requestParams {
	connectionId, err := strconv.ParseUint(input.Params["connectionId"], 10, 64)
	if err != nil || connectionId == 0 {
		connectionId = 0
	}
	scopeId := input.Params["scopeId"]
	pluginName := input.Params["plugin"]
	return &requestParams{
		connectionId: connectionId,
		scopeId:      scopeId,
		plugin:       pluginName,
	}
}

func (c *GenericScopeApiHelper[Conn, Scope, Tr]) extractFromDeleteReqParam(input *plugin.ApiResourceInput) *deleteRequestParams {
	params := c.extractFromReqParam(input)
	var err errors.Error
	var deleteDataOnly bool
	{
		ddo, ok := input.Query["delete_data_only"]
		if ok {
			deleteDataOnly, err = errors.Convert01(strconv.ParseBool(ddo[0]))
		}
		if err != nil {
			deleteDataOnly = false
		}
	}
	return &deleteRequestParams{
		requestParams:  *params,
		deleteDataOnly: deleteDataOnly,
	}
}

func (c *GenericScopeApiHelper[Conn, Scope, Tr]) extractFromGetReqParam(input *plugin.ApiResourceInput) *getRequestParams {
	params := c.extractFromReqParam(input)
	var err errors.Error
	var loadBlueprints bool
	{
		lbps, ok := input.Query["blueprints"]
		if ok {
			loadBlueprints, err = errors.Convert01(strconv.ParseBool(lbps[0]))
		}
		if err != nil {
			loadBlueprints = false
		}
	}
	return &getRequestParams{
		requestParams:  *params,
		loadBlueprints: loadBlueprints,
	}
}

func setScopeFields(p interface{}, connectionId uint64, createdDate *time.Time, updatedDate *time.Time) {
	pType := reflect.TypeOf(p)
	if pType.Kind() != reflect.Ptr {
		panic("expected a pointer to a struct")
	}
	pValue := reflectValue(p)
	// set connectionId
	connIdField := pValue.FieldByName("ConnectionId")
	connIdField.SetUint(connectionId)

	// TODO might need to change these to CreatedAt and UpdatedAt

	// set CreatedDate
	createdDateField := pValue.FieldByName("CreatedDate")
	if createdDateField.IsValid() && createdDateField.Type().AssignableTo(reflect.TypeOf(createdDate)) {
		createdDateField.Set(reflect.ValueOf(createdDate))
	}

	// set UpdatedDate
	updatedDateField := pValue.FieldByName("UpdatedDate")
	if !updatedDateField.IsValid() || (updatedDate != nil && !updatedDateField.Type().AssignableTo(reflect.TypeOf(updatedDate))) {
		return
	}
	if updatedDate == nil {
		// if updatedDate is nil, set UpdatedDate to be nil
		updatedDateField.Set(reflect.Zero(updatedDateField.Type()))
	} else {
		// if updatedDate is not nil, set UpdatedDate to be the value
		updatedDateFieldValue := reflect.ValueOf(updatedDate)
		updatedDateField.Set(updatedDateFieldValue)
	}
}

// returnPrimaryKeyValue returns a string containing the primary key value(s) of a struct, concatenated with "-" between them.
// This function receives an interface{} type argument p, which can be a pointer to any struct.
// The function uses reflection to iterate through the fields of the struct, and checks if each field is tagged as "primaryKey".
func returnPrimaryKeyValue(p interface{}) string {
	result := ""
	// get the type and value of the input interface using reflection
	t := reflectType(p)
	v := reflectValue(p)
	// iterate over each field in the struct type
	for i := 0; i < t.NumField(); i++ {
		// get the i-th field
		field := t.Field(i)

		// check if the field is marked as "primaryKey" in the struct tag
		if strings.Contains(string(field.Tag), "primaryKey") {
			// if this is the first primaryKey field encountered, set the result to be its value
			if result == "" {
				result = fmt.Sprintf("%v", v.Field(i).Interface())
			} else {
				// if this is not the first primaryKey field, append its value to the result with a "-" separator
				result = fmt.Sprintf("%s-%v", result, v.Field(i).Interface())
			}
		}
	}

	// return the final primary key value as a string
	return result
}

func VerifyScope(scope interface{}, vld *validator.Validate) errors.Error {
	if vld != nil {
		pType := reflect.TypeOf(scope)
		if pType.Kind() != reflect.Ptr {
			panic("expected a pointer to a struct")
		}
		if err := vld.Struct(scope); err != nil {
			return errors.Default.Wrap(err, "error validating target")
		}
	}
	return nil
}

func (c *GenericScopeApiHelper[Conn, Scope, Tr]) validatePrimaryKeys(scopes []*Scope) errors.Error {
	if c.validator == nil {
		return nil
	}
	keeper := make(map[string]struct{})
	for _, scope := range scopes {
		// Ensure that the primary key value is unique
		primaryValueStr := returnPrimaryKeyValue(scope)
		if _, ok := keeper[primaryValueStr]; ok {
			return errors.BadInput.New("duplicate scope was requested")
		} else {
			keeper[primaryValueStr] = struct{}{}
		}
	}
	return nil
}

// Implement MarshalJSON method to flatten all fields
func (sr *ScopeRes[T]) MarshalJSON() ([]byte, error) {
	var flatMap map[string]interface{}
	err := mapstructure.Decode(sr, &flatMap)
	if err != nil {
		return nil, err
	}
	// Encode the flattened map to JSON
	result, err := json.Marshal(flatMap)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createDeleteQuery(tableName string, scopeIdKey string, scopeId string) string {
	column := "_raw_data_params"
	if strings.HasPrefix(tableName, "_raw_") {
		column = "params"
	}
	query := `DELETE FROM ` + tableName + ` WHERE ` + column + ` LIKE '%"` + scopeIdKey + `":"` + scopeId + `"%'`
	return query
}

func getPluginTables(pluginName string) ([]string, errors.Error) {
	var tables []string
	meta, err := plugin.GetPlugin(pluginName)
	if err != nil {
		return nil, err
	}
	if pluginModel, ok := meta.(plugin.PluginModel); !ok {
		return nil, errors.Default.New(fmt.Sprintf("plugin \"%s\" does not implement listing its tables", pluginName))
	} else {
		// collect raw tables
		for _, table := range tablesCache {
			if strings.HasPrefix(table, "_raw_"+pluginName) {
				tables = append(tables, table)
			}
		}
		// collect tool tables
		tablesInfo := pluginModel.GetTablesInfo()
		for _, table := range tablesInfo {
			// we only care about tables with RawOrigin
			ok = hasField(table, "RawDataParams")
			if ok {
				tables = append(tables, table.TableName())
			}
		}
		// collect domain tables
		for _, domainTable := range domaininfo.GetDomainTablesInfo() {
			// we only care about tables with RawOrigin
			ok = hasField(domainTable, "RawDataParams")
			if ok {
				tables = append(tables, domainTable.TableName())
			}
		}
	}
	return tables, nil
}

func reflectField(obj any, fieldName string) reflect.Value {
	return reflectValue(obj).FieldByName(fieldName)
}

func hasField(obj any, fieldName string) bool {
	_, ok := reflectType(obj).FieldByName(fieldName)
	return ok
}

func reflectValue(obj any) reflect.Value {
	val := reflect.ValueOf(obj)
	kind := val.Kind()
	for kind == reflect.Ptr || kind == reflect.Interface {
		val = val.Elem()
		kind = val.Kind()
	}
	return val
}

func reflectType(obj any) reflect.Type {
	typ := reflect.TypeOf(obj)
	kind := typ.Kind()
	for kind == reflect.Ptr {
		typ = typ.Elem()
		kind = typ.Kind()
	}
	return typ
}
