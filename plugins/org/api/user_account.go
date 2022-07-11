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
	"github.com/apache/incubator-devlake/models/domainlayer/crossdomain"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocarina/gocsv"
)

func (h *Handlers) GetUserAccount(c *gin.Context) {
	aus, err := h.store.findAllUserAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	blob, err := gocsv.MarshalBytes(aus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.Data(http.StatusOK, "text/csv", blob)
}

func (h *Handlers) CreateUserAccount(c *gin.Context) {
	var aa []userAccount
	err := h.unmarshal(c, &aa)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	var au *userAccount
	var items []interface{}
	userAccounts := au.toDomainLayer(aa)
	for _, userAccount := range userAccounts {
		items = append(items, userAccount)
	}
	err = h.store.deleteAll(&crossdomain.UserAccount{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	err = h.store.save(items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, nil)
}
