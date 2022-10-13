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
	goerror "errors"
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"gorm.io/gorm"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/plugins/tapd/models"
)

type Page struct {
	Data Data `json:"data"`
}
type Data struct {
	Count int `json:"count"`
}

var statusMapping map[string]string

var UserIdGen *didgen.DomainIdGenerator
var WorkspaceIdGen *didgen.DomainIdGenerator
var IssueIdGen *didgen.DomainIdGenerator
var IterIdGen *didgen.DomainIdGenerator

// res will not be used
func GetTotalPagesFromResponse(r *http.Response, args *helper.ApiCollectorArgs) (int, errors.Error) {
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
	err = helper.UnmarshalResponse(res, &page)

	count := page.Data.Count
	totalPage := count/args.PageSize + 1

	return totalPage, err
}

func parseIterationChangelog(taskCtx core.SubTaskContext, old string, new string) (iterationFromId uint64, iterationToId uint64, err errors.Error) {
	data := taskCtx.GetData().(*TapdTaskData)
	db := taskCtx.GetDal()
	iterationFrom := &models.TapdIteration{}
	clauses := []dal.Clause{
		dal.From(&models.TapdIteration{}),
		dal.Where("connection_id = ? and workspace_id = ? and name = ?",
			data.Options.ConnectionId, data.Options.WorkspaceId, old),
	}
	err = db.First(iterationFrom, clauses...)
	if err != nil && !goerror.Is(err, gorm.ErrRecordNotFound) {
		return 0, 0, err
	}

	iterationTo := &models.TapdIteration{}
	clauses = []dal.Clause{
		dal.From(&models.TapdIteration{}),
		dal.Where("connection_id = ? and workspace_id = ? and name = ?",
			data.Options.ConnectionId, data.Options.WorkspaceId, new),
	}
	err = db.First(iterationTo, clauses...)
	if err != nil && !goerror.Is(err, gorm.ErrRecordNotFound) {
		return 0, 0, err
	}
	return iterationFrom.Id, iterationTo.Id, nil
}

func GetRawMessageDirectFromResponse(res *http.Response) ([]json.RawMessage, errors.Error) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, errors.Convert(err)
	}
	return []json.RawMessage{body}, nil
}

func GetRawMessageArrayFromResponse(res *http.Response) ([]json.RawMessage, errors.Error) {
	var data struct {
		Data []json.RawMessage `json:"data"`
	}
	err := helper.UnmarshalResponse(res, &data)
	return data.Data, err
}

type TapdApiParams struct {
	ConnectionId uint64
	CompanyId    uint64
	WorkspaceId  uint64
}

func CreateRawDataSubTaskArgs(taskCtx core.SubTaskContext, rawTable string, useCompanyId bool) (*helper.RawDataSubTaskArgs, *TapdTaskData) {
	data := taskCtx.GetData().(*TapdTaskData)
	filteredData := *data
	filteredData.Options = &TapdOptions{}
	*filteredData.Options = *data.Options
	var params = TapdApiParams{
		ConnectionId: data.Options.ConnectionId,
		WorkspaceId:  data.Options.WorkspaceId,
	}
	if data.Options.CompanyId != 0 && useCompanyId {
		params.CompanyId = data.Options.CompanyId
	} else {
		filteredData.Options.CompanyId = 0
	}
	rawDataSubTaskArgs := &helper.RawDataSubTaskArgs{
		Ctx:    taskCtx,
		Params: params,
		Table:  rawTable,
	}
	return rawDataSubTaskArgs, &filteredData
}

func getTypeMappings(data *TapdTaskData, db dal.Dal, system string) (*typeMappings, errors.Error) {
	typeIdMapping := make(map[uint64]string)
	issueTypes := make([]models.TapdWorkitemType, 0)
	clauses := []dal.Clause{
		dal.From(&models.TapdWorkitemType{}),
		dal.Where("connection_id = ? and workspace_id = ? and entity_type = ?",
			data.Options.ConnectionId, data.Options.WorkspaceId, system),
	}
	err := db.All(&issueTypes, clauses...)
	if err != nil {
		return nil, err
	}
	for _, issueType := range issueTypes {
		typeIdMapping[issueType.Id] = issueType.Name
	}
	stdTypeMappings := make(map[string]string)
	for userType, stdType := range data.Options.TransformationRules.TypeMappings {
		stdTypeMappings[userType] = strings.ToUpper(stdType.StandardType)
	}
	return &typeMappings{
		typeIdMappings:  typeIdMapping,
		stdTypeMappings: stdTypeMappings,
	}, nil
}

func getStatusMapping(data *TapdTaskData) map[string]string {
	if len(statusMapping) != 0 {
		return statusMapping
	}
	statusMapping = make(map[string]string)
	mapping := data.Options.TransformationRules.StatusMappings
	for std, orig := range mapping {
		for _, v := range orig {
			statusMapping[v] = std
		}
	}

	return statusMapping
}
