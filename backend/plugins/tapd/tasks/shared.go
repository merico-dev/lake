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

package tasks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/core/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/tapd/models"
)

type Page struct {
	Data Data `json:"data"`
}
type Data struct {
	Count int `json:"count"`
}

var priorityMap = map[string]string{
	"1": "Nice To Have",
	"2": "Low",
	"3": "Middle",
	"4": "High",
}

var accountIdGen *didgen.DomainIdGenerator
var workspaceIdGen *didgen.DomainIdGenerator
var iterIdGen *didgen.DomainIdGenerator

func getAccountIdGen() *didgen.DomainIdGenerator {
	if accountIdGen == nil {
		accountIdGen = didgen.NewDomainIdGenerator(&models.TapdAccount{})
	}
	return accountIdGen
}

func getWorkspaceIdGen() *didgen.DomainIdGenerator {
	if workspaceIdGen == nil {
		workspaceIdGen = didgen.NewDomainIdGenerator(&models.TapdWorkspace{})
	}
	return workspaceIdGen
}

func getIterIdGen() *didgen.DomainIdGenerator {
	if iterIdGen == nil {
		iterIdGen = didgen.NewDomainIdGenerator(&models.TapdIteration{})
	}
	return iterIdGen
}

// res will not be used
func GetTotalPagesFromResponse(r *http.Response, args *api.ApiCollectorArgs) (int, errors.Error) {
	data := args.Ctx.GetData().(*TapdTaskData)
	apiClient, err := NewTapdApiClient(args.Ctx.TaskContext(), data.Connection)
	if err != nil {
		return 0, err
	}
	query := url.Values{}
	query.Set("workspace_id", fmt.Sprintf("%v", data.Options.WorkspaceId))
	res, err := apiClient.Get(fmt.Sprintf("%s/count", r.Request.URL.Path), query, nil)
	if err != nil {
		return 0, err
	}
	var page Page
	err = api.UnmarshalResponse(res, &page)

	count := page.Data.Count
	totalPage := count/args.PageSize + 1

	return totalPage, err
}

// parseIterationChangelog function is used to parse the iteration changelog
func parseIterationChangelog(taskCtx plugin.SubTaskContext, old string, new string) (iterationFromId uint64, iterationToId uint64, err errors.Error) {
	data := taskCtx.GetData().(*TapdTaskData)
	db := taskCtx.GetDal()

	// Find the iteration with the old name
	iterationFrom := &models.TapdIteration{}
	clauses := []dal.Clause{
		dal.From(&models.TapdIteration{}),
		dal.Where("connection_id = ? and workspace_id = ? and name = ?",
			data.Options.ConnectionId, data.Options.WorkspaceId, old),
	}
	err = db.First(iterationFrom, clauses...)
	if err != nil && !db.IsErrorNotFound(err) {
		return 0, 0, err
	}

	// Find the iteration with the new name
	iterationTo := &models.TapdIteration{}
	clauses = []dal.Clause{
		dal.From(&models.TapdIteration{}),
		dal.Where("connection_id = ? and workspace_id = ? and name = ?",
			data.Options.ConnectionId, data.Options.WorkspaceId, new),
	}
	err = db.First(iterationTo, clauses...)
	if err != nil && !db.IsErrorNotFound(err) {
		return 0, 0, err
	}

	return iterationFrom.Id, iterationTo.Id, nil
}

// GetRawMessageDirectFromResponse extracts the raw message from an HTTP response
func GetRawMessageDirectFromResponse(res *http.Response) ([]json.RawMessage, errors.Error) {
	// Read the response body
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, errors.Convert(err)
	}
	// Return the response body as a slice of json.RawMessage
	return []json.RawMessage{body}, nil
}

func GetRawMessageArrayFromResponse(res *http.Response) ([]json.RawMessage, errors.Error) {
	var data struct {
		Data []json.RawMessage `json:"data"`
	}
	err := api.UnmarshalResponse(res, &data)
	return data.Data, err
}

type TapdApiParams models.TapdApiParams

// CreateRawDataSubTaskArgs creates a new instance of api.RawDataSubTaskArgs based on the provided
// task context, raw table name, and a flag to determine if the company ID should be used.
// It returns a pointer to the created api.RawDataSubTaskArgs and a pointer to the filtered TapdTaskData.
func CreateRawDataSubTaskArgs(taskCtx plugin.SubTaskContext, rawTable string) (*api.RawDataSubTaskArgs, *TapdTaskData) {
	// Retrieve task data from the provided task context and cast it to TapdTaskData
	data := taskCtx.GetData().(*TapdTaskData)
	// Create a filtered copy of the original data
	filteredData := *data
	filteredData.Options = &TapdOptions{}
	*filteredData.Options = *data.Options
	// Set up TapdApiParams based on the original data
	var params = TapdApiParams{
		ConnectionId: data.Options.ConnectionId,
		WorkspaceId:  data.Options.WorkspaceId,
	}
	// Create the RawDataSubTaskArgs with the task context, params, and raw table name
	rawDataSubTaskArgs := &api.RawDataSubTaskArgs{
		Ctx:    taskCtx,
		Params: params,
		Table:  rawTable,
	}
	// Return the created RawDataSubTaskArgs and the filtered TapdTaskData
	return rawDataSubTaskArgs, &filteredData
}

// getTapdTypeMappings retrieves story types from _tool_tapd_workitem_types and maps them to
// typeMapping. It takes TapdTaskData, a Dal interface, and a system string as arguments.
// It returns a map of type ID to type name and an error, if any.
func getTapdTypeMappings(data *TapdTaskData, db dal.Dal, system string) (map[uint64]string, errors.Error) {
	typeIdMapping := make(map[uint64]string)
	issueTypes := make([]models.TapdWorkitemType, 0)
	// Create clauses for querying the database
	clauses := []dal.Clause{
		dal.From(&models.TapdWorkitemType{}),
		dal.Where("connection_id = ? and workspace_id = ? and entity_type = ?",
			data.Options.ConnectionId, data.Options.WorkspaceId, system),
	}
	// Query the database for issue types
	err := db.All(&issueTypes, clauses...)
	if err != nil {
		return nil, err
	}
	// Map the retrieved issue types
	for _, issueType := range issueTypes {
		typeIdMapping[issueType.Id] = issueType.Name
	}
	return typeIdMapping, nil
}

// getStdTypeMappings creates a map of user type to standard type based on the provided TapdTaskData.
// It returns the created map.
func getStdTypeMappings(data *TapdTaskData) map[string]string {
	stdTypeMappings := make(map[string]string)
	if data.Options.ScopeConfig == nil {
		return stdTypeMappings
	}
	mapping := data.Options.ScopeConfig.TypeMappings
	// Map user types to standard types
	for userType, stdType := range mapping {
		stdTypeMappings[userType] = strings.ToUpper(stdType)
	}
	return stdTypeMappings
}

// getStatusMapping creates a map of original status values to standard status values
// based on the provided TapdTaskData. It returns the created map.
func getStatusMapping(data *TapdTaskData) map[string]string {
	stdStatusMappings := make(map[string]string)
	if data.Options.ScopeConfig == nil {
		return stdStatusMappings
	}
	mapping := data.Options.ScopeConfig.StatusMappings
	// Map original status values to standard status values
	for userStatus, stdStatus := range mapping {
		stdStatusMappings[userStatus] = strings.ToUpper(stdStatus)
	}
	return stdStatusMappings
}

// getDefaultStdStatusMapping retrieves default standard status mappings for the given TapdTaskData and status list.
// It takes TapdTaskData, a Dal interface, and a statusList of type S (models.TapdStatus).
// It returns a map of English to Chinese status names, a function to get standard status from status key, and an error, if any.
func getDefaultStdStatusMapping[S models.TapdStatus](data *TapdTaskData, db dal.Dal, statusList []S) (map[string]string, func(statusKey string) string, errors.Error) {
	// Create clauses for querying the database
	clauses := []dal.Clause{
		dal.Where("connection_id = ? and workspace_id = ?", data.Options.ConnectionId, data.Options.WorkspaceId),
	}
	// Query the database for status list
	err := db.All(&statusList, clauses...)
	if err != nil {
		return nil, nil, err
	}

	// Create status language and last step maps
	statusLanguageMap := make(map[string]string, len(statusList))
	statusLastStepMap := make(map[string]bool, len(statusList))

	// Populate status maps
	for _, v := range statusList {
		statusLanguageMap[v.GetEnglish()] = v.GetChinese()
		statusLastStepMap[v.GetChinese()] = v.GetIsLastStep()
	}

	// Define function to get standard status from status key
	getStdStatus := func(statusKey string) string {
		if statusLastStepMap[statusKey] {
			return ticket.DONE
		} else if statusKey == "草稿" {
			return ticket.TODO
		} else {
			return ticket.IN_PROGRESS
		}
	}
	return statusLanguageMap, getStdStatus, nil
}

// unicodeToZh converts a string containing Unicode escape sequences to a Chinese string.
// It returns the converted string and an error if the conversion fails.
func unicodeToZh(s string) (string, error) {
	str, err := strconv.Unquote(fmt.Sprintf(`"%v"`, strings.Replace(s, `"`, `\"`, -1)))
	if err != nil {
		return "", err
	}
	return str, nil
}

// convertUnicode converts the ValueAfterParsed and ValueBeforeParsed fields of a struct to Chinese text.
// It takes a pointer to a struct and returns an error if the conversion fails.
func convertUnicode(p interface{}) errors.Error {
	var err errors.Error
	pType := reflect.TypeOf(p)
	if pType.Kind() != reflect.Ptr {
		panic("expected a pointer to a struct")
	}
	pValue := reflect.ValueOf(p).Elem()
	if pValue.Kind() != reflect.Struct {
		panic("expected a pointer to a struct")
	}
	after, err := errors.Convert01(unicodeToZh(pValue.FieldByName("ValueAfterParsed").String()))
	if err != nil {
		return err
	}
	before, err := errors.Convert01(unicodeToZh(pValue.FieldByName("ValueBeforeParsed").String()))
	if err != nil {
		return err
	}
	if after == "--" {
		after = ""
	}
	if before == "--" {
		before = ""
	}
	// Set ValueAfterParsed and ValueBeforeParsed fields
	valueAfterField := pValue.FieldByName("ValueAfterParsed")
	valueAfterField.SetString(after)
	valueBeforeField := pValue.FieldByName("ValueBeforeParsed")
	valueBeforeField.SetString(before)
	return nil
}

// replaceSemicolonWithComma replaces all semicolons with commas in the given string
// and trims any trailing commas. It returns the modified string.
func replaceSemicolonWithComma(str string) string {
	res := strings.ReplaceAll(str, ";", ",")
	return strings.TrimRight(res, ",")
}

// generateDomainAccountIdForUsers generates domain account IDs for a list of users.
// The input 'param' is a string with format "user1,user2,user3".
// The function takes a string containing a list of users separated by commas and a connectionId.
// It returns a string containing the generated domain account IDs for each user, separated by commas.
func generateDomainAccountIdForUsers(param string, connectionId uint64) string {
	if param == "" {
		return ""
	}
	param = replaceSemicolonWithComma(param)
	users := strings.Split(param, ",")
	var res []string
	for _, user := range users {
		res = append(res, getAccountIdGen().Generate(connectionId, user))
	}
	return strings.Join(res, ",")
}

// extractStatus extracts the status from the given blob and returns a map of status names to status values.
func extractStatus(blob []byte) (map[string]string, errors.Error) {
	var statusRes struct {
		Data interface{} `json:"data"`
	}
	err := errors.Convert(json.Unmarshal(blob, &statusRes))
	if err != nil {
		return nil, err
	}
	data, ok := statusRes.Data.(map[string]interface{})
	if !ok {
		return nil, nil
	}
	results := make(map[string]string)
	for k, v := range data {
		if value, ok := v.(string); ok {
			results[k] = value
		}
	}
	return results, nil
}
