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
	"github.com/gin-gonic/gin"
	"github.com/gocarina/gocsv"
)

func (h *Handlers) GetUser(c *gin.Context) {
	var query struct {
		FakeData bool `form:"fake_data"`
	}
	_ = c.BindQuery(&query)
	var users []user
	var u *user
	if query.FakeData {
		users = u.fakeData()
	} else {
		var uu []crossdomain.User
		err := h.store.findAll(&uu)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		var tus []crossdomain.TeamUser
		users = u.fromDomainLayer(uu, tus)
	}
	blob, err := gocsv.MarshalBytes(users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.Data(http.StatusOK, "text/csv", blob)
}

func (h *Handlers) CreateUser(c *gin.Context) {
	var uu []user
	err := h.unmarshal(c, &uu)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	var u *user
	var items []interface{}
	users, teamUsers := u.toDomainLayer(uu)
	for _, user := range users {
		items = append(items, user)
	}
	for _, teamUser := range teamUsers {
		items = append(items, teamUser)
	}
	err = h.store.save(items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, nil)
}
