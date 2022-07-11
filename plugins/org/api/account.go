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

	"github.com/gin-gonic/gin"
	"github.com/gocarina/gocsv"
)

func (h *Handlers) GetAccount(c *gin.Context) {
	accounts, err := h.store.findAllAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	blob, err := gocsv.MarshalBytes(accounts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.Data(http.StatusOK, "text/csv", blob)
}

func (h *Handlers) CreateAccount(c *gin.Context) {
	var aa []account
	err := h.unmarshal(c, &aa)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	var a *account
	var items []interface{}
	accounts, userAccounts := a.toDomainLayer(aa)
	for _, acc := range accounts {
		items = append(items, acc)
	}
	for _, userAccount := range userAccounts {
		items = append(items, userAccount)
	}
	err = h.store.save(items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, nil)
}
