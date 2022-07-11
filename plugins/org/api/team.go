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

func (h *Handlers) GetTeam(c *gin.Context) {
	var query struct {
		FakeData bool `form:"fake_data"`
	}
	_ = c.BindQuery(&query)
	var teams []team
	var t *team
	var err error
	if query.FakeData {
		teams = t.fakeData()
	} else {
		teams, err = h.store.findAllTeams()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
	}
	blob, err := gocsv.MarshalBytes(teams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.Data(http.StatusOK, "text/csv", blob)
}

func (h *Handlers) CreateTeam(c *gin.Context) {
	var tt []team
	err := h.unmarshal(c, &tt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	var t *team
	var items []interface{}
	for _, tm := range t.toDomainLayer(tt) {
		items = append(items, tm)
	}
	err = h.store.save(items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, nil)
}
