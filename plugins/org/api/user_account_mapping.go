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
	"net/http"

	"github.com/apache/incubator-devlake/models/domainlayer/crossdomain"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/gocarina/gocsv"
)

// GetUserAccountMapping godoc
// @Summary      Get user_account_mapping.csv.csv file
// @Description  get user_account_mapping.csv.csv file
// @Tags 		 plugins/org
// @Produce      text/csv
// @Success      200
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router       /plugins/org/accounts.csv [get]
func (h *Handlers) GetUserAccountMapping(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	accounts, err := h.store.findAllAccounts()
	if err != nil {
		return nil, err
	}
	blob, err := gocsv.MarshalBytes(accounts)
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

// CreateUserAccountMapping godoc
// @Summary      Upload user_account_mapping.csv.csv file
// @Description  upload user_account_mapping.csv.csv file
// @Tags 		 plugins/org
// @Accept       text/csv
// @Produce      json
// @Success      200
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router       /plugins/org/accounts.csv [put]
func (h *Handlers) CreateUserAccountMapping(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
	var aa []account
	err := h.unmarshal(input.Request, &aa)
	if err != nil {
		return nil, err
	}
	var a *account
	var items []interface{}
	userAccounts := a.toDomainLayer(aa)
	for _, userAccount := range userAccounts {
		items = append(items, userAccount)
	}
	err = h.store.deleteAll(&crossdomain.UserAccount{})
	if err != nil {
		return nil, err
	}
	err = h.store.save(items)
	if err != nil {
		return nil, err
	}
	return &core.ApiResourceOutput{Status: http.StatusOK}, nil
}
