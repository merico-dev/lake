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
	"github.com/apache/incubator-devlake/plugins/core"
	"net/http"

	"github.com/apache/incubator-devlake/models/domainlayer/crossdomain"
	"github.com/gocarina/gocsv"
)

// GetTeam godoc
// @Summary      Get teams.csv file
// @Description  get teams.csv file
// @Tags 		 plugins/org
// @Produce      text/csv
// @Param        fake_data    query     bool  false  "return fake data or not"
// @Success      200
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router       /plugins/org/teams.csv [get]
func (h *Handlers) GetTeam(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	input.Query.Get("fake_data")
	var teams []team
	var t *team
	var err error
	if input.Query.Get("fake_data") == "true" {
		teams = t.fakeData()
	} else {
		teams, err = h.store.findAllTeams()
		if err != nil {
			return nil, err
		}
	}
	blob, err := gocsv.MarshalBytes(teams)
	if err != nil {
		return nil, err
	}
	return &core.ApiResourceOutput{
		Body:   nil,
		Status: http.StatusOK,
		File: &core.OutputFile{
			ContentType: "text/csv",
			Data:        blob,
		},
	}, nil
}

// CreateTeam godoc
// @Summary      Upload teams.csv file
// @Description  upload teams.csv file
// @Tags 		 plugins/org
// @Accept       text/csv
// @Produce      json
// @Success      200
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router       /plugins/org/teams.csv [put]
func (h *Handlers) CreateTeam(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	var tt []team
	err := h.unmarshal(input.Request, &tt)
	if err != nil {
		return nil, err
	}
	var t *team
	var items []interface{}
	for _, tm := range t.toDomainLayer(tt) {
		items = append(items, tm)
	}
	err = h.store.deleteAll(&crossdomain.Team{})
	if err != nil {
		return nil, err
	}
	err = h.store.save(items)
	if err != nil {
		return nil, err
	}
	return &core.ApiResourceOutput{Status: http.StatusOK}, nil
}
