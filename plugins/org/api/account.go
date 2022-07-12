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

	"github.com/apache/incubator-devlake/api/shared"
	"github.com/apache/incubator-devlake/models/domainlayer/crossdomain"
	"github.com/gin-gonic/gin"
	"github.com/gocarina/gocsv"
)

// GetAccount godoc
// @Summary      Get account.csv file
// @Description  get account.csv file
// @Tags 		 plugins/org
// @Produce      text/csv
// @Success      200
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router       /plugins/org/accounts.csv [get]
func (h *Handlers) GetAccount(c *gin.Context) {
	accounts, err := h.store.findAllAccounts()
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusInternalServerError)
		return
	}
	blob, err := gocsv.MarshalBytes(accounts)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "text/csv", blob)
}

// CreateAccount godoc
// @Summary      Upload account.csv file
// @Description  upload account.csv file
// @Tags 		 plugins/org
// @Accept       text/csv
// @Produce      json
// @Success      200
// @Failure 400  {object} shared.ApiBody "Bad Request"
// @Failure 500  {object} shared.ApiBody "Internal Error"
// @Router       /plugins/org/accounts.csv [put]
func (h *Handlers) CreateAccount(c *gin.Context) {
	var aa []account
	err := h.unmarshal(c, &aa)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusBadRequest)
		return
	}
	var a *account
	var items []interface{}
	userAccounts := a.toDomainLayer(aa)
	for _, userAccount := range userAccounts {
		items = append(items, userAccount)
	}
	err = h.store.deleteAll(&crossdomain.UserAccount{})
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusInternalServerError)
		return
	}
	err = h.store.save(items)
	if err != nil {
		shared.ApiOutputError(c, err, http.StatusInternalServerError)
		return
	}
	shared.ApiOutputSuccess(c, nil, http.StatusOK)
}
