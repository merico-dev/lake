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
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
)

type GitlabApiParams struct {
	ConnectionId uint64
	ProjectId    int
}

type GitlabInput struct {
	GitlabId int
	Iid      int
}

func GetTotalPagesFromResponse(res *http.Response, args *helper.ApiCollectorArgs) (int, error) {
	total := res.Header.Get("X-Total-Pages")
	if total == "" {
		return 0, nil
	}
	totalInt, err := strconv.Atoi(total)
	if err != nil {
		return 0, err
	}
	return totalInt, nil
}

func GetRawMessageFromResponse(res *http.Response) ([]json.RawMessage, error) {
	rawMessages := []json.RawMessage{}

	if res == nil {
		return nil, fmt.Errorf("res is nil")
	}
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", err, res.Request.URL.String())
	}

	err = json.Unmarshal(resBody, &rawMessages)
	if err != nil {
		return nil, fmt.Errorf("%w %s %s", err, res.Request.URL.String(), string(resBody))
	}

	return rawMessages, nil
}

func GetQuery(reqData *helper.RequestData) (url.Values, error) {
	query := url.Values{}
	query.Set("with_stats", "true")
	query.Set("sort", "asc")
	query.Set("page", strconv.Itoa(reqData.Pager.Page))
	query.Set("per_page", strconv.Itoa(reqData.Pager.Size))
	return query, nil
}

func CreateRawDataSubTaskArgs(taskCtx core.SubTaskContext, Table string) (*helper.RawDataSubTaskArgs, *GitlabTaskData) {
	data := taskCtx.GetData().(*GitlabTaskData)
	RawDataSubTaskArgs := &helper.RawDataSubTaskArgs{
		Ctx: taskCtx,
		Params: GitlabApiParams{
			ProjectId:    data.Options.ProjectId,
			ConnectionId: data.Options.ConnectionId,
		},
		Table: Table,
	}
	return RawDataSubTaskArgs, data
}

func GetMergeRequestsIterator(taskCtx core.SubTaskContext) (*helper.DalCursorIterator, error) {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*GitlabTaskData)
	clauses := []dal.Clause{
		dal.Select("gmr.gitlab_id, gmr.iid"),
		dal.From("_tool_gitlab_merge_requests gmr"),
		dal.Where(
			`gmr.project_id = ? and gmr.connection_id = ?`,
			data.Options.ProjectId, data.Options.ConnectionId,
		),
	}
	// construct the input iterator
	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return nil, err
	}

	return helper.NewDalCursorIterator(db, cursor, reflect.TypeOf(GitlabInput{}))
}
