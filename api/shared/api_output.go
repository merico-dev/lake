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

package shared

import (
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/logger"
	"github.com/apache/incubator-devlake/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ApiBody struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ResponsePipelines struct {
	Count     int64                 `json:"count"`
	Pipelines []*models.ApiPipeline `json:"pipelines"`
}

// ApiOutputError writes a JSON error message to the HTTP response body
func ApiOutputError(c *gin.Context, err error) {
	if e, ok := err.(errors.Error); ok {
		logger.Global.Error(err, "HTTP %d error", e.GetType().GetHttpCode())
		c.JSON(e.GetType().GetHttpCode(), &ApiBody{
			Success: false,
			Message: e.UserMessage(),
		})
	} else {
		logger.Global.Error(err, "HTTP %d error (native)", http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, &ApiBody{
			Success: false,
			Message: err.Error(),
		})
	}
	c.Writer.Header().Set("Content-Type", "application/json")
}

// ApiOutputSuccess writes a JSON success message to the HTTP response body
func ApiOutputSuccess(c *gin.Context, body interface{}, status int) {
	if body == nil {
		body = &ApiBody{
			Success: true,
			Message: "success",
		}
	}
	c.JSON(status, body)
}

// ApiOutputAbort writes the HTTP response code header and saves the error internally, but doesn't push it to the response
func ApiOutputAbort(c *gin.Context, err error) {
	if e, ok := err.(errors.Error); ok {
		logger.Global.Error(err, "HTTP %d abort-error", e.GetType().GetHttpCode())
		_ = c.AbortWithError(e.GetType().GetHttpCode(), fmt.Errorf(e.UserMessage()))
	} else {
		logger.Global.Error(err, "HTTP %d abort-error (native)", http.StatusInternalServerError)
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}
}
