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
	"github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/models/domainlayer/domaininfo"
	serviceHelper "github.com/apache/incubator-devlake/helpers/pluginhelper/services"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/log"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"

	"reflect"
)

type NoTransformation struct{}

var (
	tablesCache       []string // these cached vars can probably be moved somewhere more centralized later
	tablesCacheLoader = new(sync.Once)
)

type (
	// ScopeApiHelper is used to write the CURD of scopes
	ScopeApiHelper[Conn any, Scope any, Tr any] struct {
		*GenericScopeHelper[Scope, Tr]
		validator  *validator.Validate
		connHelper *ConnectionApiHelper
	}
	GenericScopeHelper[Scope any, Tr any] struct {
		log                      log.Logger
		db                       dal.Dal
		reflectionParams         *ReflectionParameters
		connectionVerifier       func(connectionId uint64) errors.Error
		scopeGetter              func(connectionId uint64, scopeId string) (Scope, errors.Error)
		scopeLister              func(input *plugin.ApiResourceInput, connectionId uint64) ([]*Scope, errors.Error)
		scopeDeleter             func(connectionId uint64, scopeId string) errors.Error
		transformationRuleGetter func(ruleId uint64) (Tr, errors.Error)
		transformationRuleLister func(ruleIds []uint64) ([]*Tr, errors.Error)
	}
)

type ReflectionParameters struct {
	ScopeIdFieldName  string
	ScopeIdColumnName string
	RawScopeParamName string
}

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

// NewGenericScopeHelper creates a GenericScopeHelper for scopes management (compatible with remote plugins)
func NewGenericScopeHelper[Scope any, Tr any](
	basicRes context.BasicRes,
	params *ReflectionParameters,
	connectionVerifier func(connectionId uint64) errors.Error,
	scopeGetter func(connectionId uint64, scopeId string) (Scope, errors.Error),
	scopeLister func(input *plugin.ApiResourceInput, connectionId uint64) ([]*Scope, errors.Error),
	scopeDeleter func(connectionId uint64, scopeId string) errors.Error,
	transformationRuleGetter func(ruleId uint64) (Tr, errors.Error),
	transformationRuleLister func(ruleIds []uint64) ([]*Tr, errors.Error),
) *GenericScopeHelper[Scope, Tr] {
	tablesCacheLoader.Do(func() {
		var err errors.Error
		tablesCache, err = basicRes.GetDal().AllTables()
		if err != nil {
			panic(err)
		}
	})
	return &GenericScopeHelper[Scope, Tr]{
		log:                      basicRes.GetLogger(),
		db:                       basicRes.GetDal(),
		reflectionParams:         params,
		connectionVerifier:       connectionVerifier,
		scopeGetter:              scopeGetter,
		scopeLister:              scopeLister,
		scopeDeleter:             scopeDeleter,
		transformationRuleGetter: transformationRuleGetter,
		transformationRuleLister: transformationRuleLister,
	}
}

// NewScopeHelper creates a ScopeHelper for scopes management
func NewScopeHelper[Conn any, Scope any, Tr any](
	basicRes context.BasicRes,
	vld *validator.Validate,
	connHelper *ConnectionApiHelper,
	params *ReflectionParameters,
) *ScopeApiHelper[Conn, Scope, Tr] {
	if vld == nil {
		vld = validator.New()
	}
	if connHelper == nil {
		return nil
	}
	if params == nil {
		panic("reflection params not provided")
	}
	return &ScopeApiHelper[Conn, Scope, Tr]{
		GenericScopeHelper: NewGenericScopeHelper[Scope, Tr](
			basicRes,
			params,
			func(connectionId uint64) errors.Error {
				var conn Conn
				err := connHelper.FirstById(&conn, connectionId)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return errors.BadInput.New("Invalid Connection Id")
					}
					return err
				}
				return nil
			},
			func(connectionId uint64, scopeId string) (Scope, errors.Error) {
				query := dal.Where(fmt.Sprintf("connection_id = ? AND %s = ?", params.ScopeIdColumnName), connectionId, scopeId)
				var scope Scope
				err := basicRes.GetDal().First(&scope, query)
				if basicRes.GetDal().IsErrorNotFound(err) {
					return scope, errors.NotFound.New("Scope not found")
				}
				return scope, nil
			},
			func(input *plugin.ApiResourceInput, connectionId uint64) ([]*Scope, errors.Error) {
				limit, offset := GetLimitOffset(input.Query, "pageSize", "page")
				var scopes []*Scope
				err := basicRes.GetDal().All(&scopes, dal.Where("connection_id = ?", connectionId), dal.Limit(limit), dal.Offset(offset))
				return scopes, err
			},
			func(connectionId uint64, scopeId string) errors.Error {
				scope := new(Scope)
				err := basicRes.GetDal().Delete(&scope, dal.Where(fmt.Sprintf("connection_id = ? AND %s = ?", params.ScopeIdFieldName),
					connectionId, scopeId))
				return err
			},
			func(repoRuleId uint64) (Tr, errors.Error) {
				var rule Tr
				err := basicRes.GetDal().First(&rule, dal.Where("id = ?", repoRuleId))
				if err != nil {
					return rule, errors.NotFound.New("transformationRule not found")
				}
				return rule, nil
			},
			func(ruleIds []uint64) ([]*Tr, errors.Error) {
				var rules []*Tr
				err := basicRes.GetDal().All(&rules, dal.Where("id IN (?)", ruleIds))
				return rules, err
			},
		),
		validator:  vld,
		connHelper: connHelper,
	}
}

type ScopeRes[T any] struct {
	Scope                  T                   `mapstructure:",squash"`
	TransformationRuleName string              `mapstructure:"transformationRuleName,omitempty"`
	Blueprints             []*models.Blueprint `mapstructure:"blueprints,omitempty"`
}

type ScopeReq[T any] struct {
	Data []*T `json:"data"`
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
	params := extractFromReqParam(input)
	if params.connectionId == 0 {
		return nil, errors.BadInput.New("invalid connectionId")
	}
	err = c.VerifyConnection(params.connectionId)
	if err != nil {
		return nil, err
	}
	// Create a map to keep track of primary key values
	keeper := make(map[string]struct{})

	// Set the CreatedDate and UpdatedDate fields to the current time for each scope
	now := time.Now()
	for _, v := range req.Data {
		// Ensure that the primary key value is unique
		primaryValueStr := returnPrimaryKeyValue(*v)
		if _, ok := keeper[primaryValueStr]; ok {
			return nil, errors.BadInput.New("duplicated item")
		} else {
			keeper[primaryValueStr] = struct{}{}
		}

		// Set the connection ID, CreatedDate, and UpdatedDate fields
		setScopeFields(v, params.connectionId, &now, &now)

		// Verify that the primary key value is valid
		err = VerifyScope(v, c.validator)
		if err != nil {
			return nil, err
		}
	}
	// Save the scopes to the database
	if req.Data != nil && len(req.Data) > 0 {
		err = c.save(&req.Data)
		if err != nil {
			return nil, err
		}
	}

	apiScopes, err := c.addTransformationName(req.Data)
	if err != nil {
		return nil, err
	}

	return &plugin.ApiResourceOutput{Body: apiScopes, Status: http.StatusOK}, nil
}

func (c *ScopeApiHelper[Conn, Scope, Tr]) Update(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	params := extractFromReqParam(input)
	if params.connectionId == 0 {
		return nil, errors.BadInput.New("invalid connectionId")
	}
	if len(params.scopeId) == 0 {
		return nil, errors.BadInput.New("invalid scopeId")
	}
	err := c.VerifyConnection(params.connectionId)
	if err != nil {
		return &plugin.ApiResourceOutput{Body: nil, Status: http.StatusInternalServerError}, err
	}
	var scope Scope
	err = c.db.First(&scope, dal.Where(fmt.Sprintf("connection_id = ? AND %s = ?", c.reflectionParams.ScopeIdColumnName), params.connectionId, params.scopeId))
	if err != nil {
		return &plugin.ApiResourceOutput{Body: nil, Status: http.StatusInternalServerError}, errors.Default.New("getting Scope error")
	}
	err = DecodeMapStruct(input.Body, &scope, true)
	if err != nil {
		return &plugin.ApiResourceOutput{Body: nil, Status: http.StatusInternalServerError}, errors.Default.Wrap(err, "patch scope error")
	}
	err = VerifyScope(&scope, c.validator)
	if err != nil {
		return &plugin.ApiResourceOutput{Body: nil, Status: http.StatusInternalServerError}, errors.Default.Wrap(err, "Invalid scope")
	}

	err = c.db.Update(scope)
	if err != nil {
		return &plugin.ApiResourceOutput{Body: nil, Status: http.StatusInternalServerError}, errors.Default.Wrap(err, "error on saving Scope")
	}
	valueRepoRuleId := reflect.ValueOf(scope).FieldByName("TransformationRuleId")
	if !valueRepoRuleId.IsValid() {
		return &plugin.ApiResourceOutput{Body: scope, Status: http.StatusOK}, nil
	}
	repoRuleId := reflect.ValueOf(scope).FieldByName("TransformationRuleId").Uint()
	var rule Tr
	if repoRuleId > 0 {
		err = c.db.First(&rule, dal.Where("id = ?", repoRuleId))
		if err != nil {
			return nil, errors.NotFound.New("transformationRule not found")
		}
	}
	scopeRes := &ScopeRes[Scope]{
		Scope:                  scope,
		TransformationRuleName: reflect.ValueOf(rule).FieldByName("Name").String(),
	}

	return &plugin.ApiResourceOutput{Body: scopeRes, Status: http.StatusOK}, nil
}

// GetScopeList returns a list of scopes. It expects a fieldName argument, which is used
// to extract the connection ID from the input.Params map.
func (c *ScopeApiHelper[Conn, Scope, Tr]) GetScopeList(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	scopes, err := c.GenericScopeHelper.GetScopeList(input)
	if err != nil {
		return nil, err
	}
	return &plugin.ApiResourceOutput{Body: scopes, Status: http.StatusOK}, nil
}

func (c *ScopeApiHelper[Conn, Scope, Tr]) GetScope(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	scope, err := c.GenericScopeHelper.GetScope(input)
	if err != nil {
		return nil, err
	}
	return &plugin.ApiResourceOutput{Body: scope, Status: http.StatusOK}, nil
}

func (c *ScopeApiHelper[Conn, Scope, Tr]) Delete(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	err := c.GenericScopeHelper.Delete(input)
	if err != nil {
		return nil, err
	}
	return &plugin.ApiResourceOutput{Body: nil, Status: http.StatusOK}, nil
}

func (c *GenericScopeHelper[Scope, Tr]) GetScopeList(input *plugin.ApiResourceInput) ([]*ScopeRes[Scope], errors.Error) {
	params := extractFromGetReqParam(input)
	if params.connectionId == 0 {
		return nil, errors.BadInput.New("invalid path params: \"connectionId\" not set")
	}
	err := c.connectionVerifier(params.connectionId)
	if err != nil {
		return nil, err
	}
	scopes, err := c.scopeLister(input, params.connectionId)
	if err != nil {
		return nil, err
	}
	apiScopes, err := c.addTransformationName(scopes)
	if err != nil {
		return nil, err
	}
	if params.loadBlueprints {
		scopesById := c.mapByScopeId(apiScopes)
		var scopeIds []string
		for id, _ := range scopesById {
			scopeIds = append(scopeIds, id)
		}
		blueprintMap, err := serviceHelper.NewBlueprintManager(c.db).GetBlueprintsByScopes(scopeIds...)
		if err != nil {
			return nil, errors.Default.Wrap(err, "error getting blueprints for scope")
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
			for k, _ := range blueprintMap {
				danglingIds = append(danglingIds, k)
			}
			c.log.Warn(nil, "The following dangling scopes were found: %v", danglingIds)
		}
	}
	b, err0 := json.Marshal(&apiScopes)
	_ = b
	_ = err0
	return apiScopes, nil
}

func (c *GenericScopeHelper[Scope, Tr]) GetScope(input *plugin.ApiResourceInput) (*ScopeRes[Scope], errors.Error) {
	params := extractFromGetReqParam(input)
	if params == nil || params.connectionId == 0 {
		return nil, errors.BadInput.New("invalid path params: \"connectionId\" not set")
	}
	if len(params.scopeId) == 0 || params.scopeId == "0" {
		return nil, errors.BadInput.New("invalid path params: \"scopeId\" not set/invalid")
	}
	err := c.connectionVerifier(params.connectionId)
	if err != nil {
		return nil, err
	}
	db := c.db
	scope, err := c.scopeGetter(params.connectionId, params.scopeId)
	if err != nil {
		return nil, err
	}
	valueRepoRuleId := reflect.ValueOf(scope).FieldByName("TransformationRuleId")
	transformationRuleName := ""
	if valueRepoRuleId.IsValid() {
		repoRuleId := valueRepoRuleId.Uint()
		var rule any
		if repoRuleId > 0 {
			rule, err = c.transformationRuleGetter(repoRuleId)
			if err != nil {
				return nil, err
			}
		}
		transformationRuleName = reflect.ValueOf(rule).FieldByName("Name").String()
	}
	var blueprints []*models.Blueprint
	if params.loadBlueprints {
		blueprintMap, err := serviceHelper.NewBlueprintManager(db).GetBlueprintsByScopes(params.scopeId)
		if err != nil {
			return nil, errors.Default.Wrap(err, "error getting blueprints for scope")
		}
		if len(blueprintMap) == 1 {
			blueprints = blueprintMap[params.scopeId]
		}
	}
	scopeRes := &ScopeRes[Scope]{
		Scope:                  scope,
		TransformationRuleName: transformationRuleName,
		Blueprints:             blueprints,
	}
	return scopeRes, nil
}

func (c *GenericScopeHelper[Scope, Tr]) Delete(input *plugin.ApiResourceInput) errors.Error {
	params := extractFromDeleteReqParam(input)
	if params == nil || params.connectionId == 0 {
		return errors.BadInput.New("invalid path params: \"connectionId\" not set")
	}
	if len(params.scopeId) == 0 || params.scopeId == "0" {
		return errors.BadInput.New("invalid path params: \"scopeId\" not set/invalid")
	}
	err := c.connectionVerifier(params.connectionId)
	if err != nil {
		return err
	}
	db := c.db
	bpManager := serviceHelper.NewBlueprintManager(db)
	blueprintsMap, err := bpManager.GetBlueprintsByScopes(params.scopeId)
	if err != nil {
		return err
	}
	blueprints := blueprintsMap[params.scopeId]
	// find all tables for this plugin
	tables, err := getPluginTables(params.plugin)
	if err != nil {
		return err
	}
	// delete all the plugin records referencing this scope
	if c.reflectionParams.RawScopeParamName != "" {
		for _, table := range tables {
			err = db.Exec(createDeleteQuery(table, c.reflectionParams.RawScopeParamName, params.scopeId))
			if err != nil {
				return err
			}
		}
	}
	if !params.deleteDataOnly {
		// delete the scope itself
		err = c.scopeDeleter(params.connectionId, params.scopeId)
		if err != nil {
			return err
		}
		// update the blueprints (remove scope reference from them)
		for _, blueprint := range blueprints {
			settings, _ := blueprint.UnmarshalSettings()
			err = settings.UpdateConnections(func(c *plugin.BlueprintConnectionV200) errors.Error {
				var filteredScopes []*plugin.BlueprintScopeV200
				for _, bpScope := range c.Scopes {
					if bpScope.Id != params.scopeId { // keep the ones NOT equal to this scope
						filteredScopes = append(filteredScopes, bpScope)
					}
				}
				c.Scopes = filteredScopes
				return nil
			})
			if err != nil {
				return errors.Default.Wrap(err, fmt.Sprintf("error removing scope %s from blueprint %d", params.scopeId, blueprint.ID))
			}
		}
	}
	return nil
}

func (c *ScopeApiHelper[Conn, Scope, Tr]) VerifyConnection(connId uint64) errors.Error {
	var conn Conn
	err := c.connHelper.FirstById(&conn, connId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.BadInput.New("Invalid Connection Id")
		}
		return err
	}
	return nil
}

func (c *GenericScopeHelper[Scope, Tr]) addTransformationName(scopes []*Scope) ([]*ScopeRes[Scope], errors.Error) {
	var ruleIds []uint64
	for _, scope := range scopes {
		valueRepoRuleId := reflectField(scope, "TransformationRuleId")
		if !valueRepoRuleId.IsValid() {
			break
		}
		ruleId := valueRepoRuleId.Uint()
		if ruleId > 0 {
			ruleIds = append(ruleIds, ruleId)
		}
	}
	var rules []*Tr
	var err errors.Error
	if len(ruleIds) > 0 {
		rules, err = c.transformationRuleLister(ruleIds)
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

func (c *GenericScopeHelper[Scope, Tr]) mapByScopeId(scopes []*ScopeRes[Scope]) map[string]*ScopeRes[Scope] {
	scopeMap := map[string]*ScopeRes[Scope]{}
	for _, scope := range scopes {
		scopeId := fmt.Sprintf("%v", reflectField(scope.Scope, c.reflectionParams.ScopeIdFieldName).Interface())
		scopeMap[scopeId] = scope
	}
	return scopeMap
}

func (c *ScopeApiHelper[Conn, Scope, Tr]) save(scope interface{}) errors.Error {
	err := c.db.CreateOrUpdate(scope)
	if err != nil {
		if c.db.IsDuplicationError(err) {
			return errors.BadInput.New("the scope already exists")
		}
		return err
	}
	return nil
}

func extractFromReqParam(input *plugin.ApiResourceInput) *requestParams {
	connectionId, err := strconv.ParseUint(input.Params["connectionId"], 10, 64)
	if err != nil || connectionId == 0 {
		connectionId = 0
	}
	scopeId, _ := input.Params["scopeId"]
	pluginName := input.Params["plugin"]
	return &requestParams{
		connectionId: connectionId,
		scopeId:      scopeId,
		plugin:       pluginName,
	}
}

func extractFromDeleteReqParam(input *plugin.ApiResourceInput) *deleteRequestParams {
	params := extractFromReqParam(input)
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

func extractFromGetReqParam(input *plugin.ApiResourceInput) *getRequestParams {
	params := extractFromReqParam(input)
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
	pValue := reflect.ValueOf(p).Elem()

	// set connectionId
	connIdField := pValue.FieldByName("ConnectionId")
	connIdField.SetUint(connectionId)

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
	t := reflect.TypeOf(p)
	v := reflect.ValueOf(p)
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
			tables = append(tables, table.TableName())
		}
		// collect domain tables
		for _, domainTable := range domaininfo.GetDomainTablesInfo() {
			// we only care about tables with RawOrigin
			_, ok = reflect.TypeOf(domainTable).Elem().FieldByName("RawDataParams")
			if ok {
				tables = append(tables, domainTable.TableName())
			}
		}
	}
	return tables, nil
}

func reflectField(obj any, fieldName string) reflect.Value {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	return val.FieldByName(fieldName)
}
