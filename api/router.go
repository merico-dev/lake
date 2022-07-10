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
	"fmt"
	"net/http"

	"github.com/apache/incubator-devlake/api/blueprints"
	"github.com/apache/incubator-devlake/api/domainlayer"

	"github.com/apache/incubator-devlake/api/ping"
	"github.com/apache/incubator-devlake/api/pipelines"
	"github.com/apache/incubator-devlake/api/push"
	"github.com/apache/incubator-devlake/api/shared"
	"github.com/apache/incubator-devlake/api/task"
	"github.com/apache/incubator-devlake/api/version"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/services"
	"github.com/gin-gonic/gin"
)

func RegisterRouter(r *gin.Engine) {
	r.GET("/pipelines", pipelines.Index)
	r.POST("/pipelines", pipelines.Post)
	r.GET("/pipelines/:pipelineId", pipelines.Get)
	r.PATCH("/blueprints/:blueprintId", blueprints.Patch)
	r.POST("/blueprints/:blueprintId/trigger", blueprints.Trigger)
	// r.DELETE("/blueprints/:blueprintId", blueprints.Delete)

	r.GET("/blueprints", blueprints.Index)
	r.POST("/blueprints", blueprints.Post)
	r.GET("/blueprints/:blueprintId", blueprints.Get)
	r.DELETE("/pipelines/:pipelineId", pipelines.Delete)
	r.GET("/pipelines/:pipelineId/tasks", task.Index)

	r.GET("/ping", ping.Get)
	r.GET("/version", version.Get)
	r.POST("/push/:tableName", push.Post)
	r.GET("/domainlayer/repos", domainlayer.ReposIndex)

	for name, pluginEntry := range core.AllPlugins() {
		if routerSetter, ok := pluginEntry.(core.PluginRouterSetter); ok {
			routerSetter.SetRouter(r.Group("/plugins/" + name))
		}
	}
	// mount all api resources for all plugins
	pluginsApiResources, err := services.GetPluginsApiResources()
	if err != nil {
		panic(err)
	}
	for pluginName, apiResources := range pluginsApiResources {
		for resourcePath, resourceHandlers := range apiResources {
			for method, h := range resourceHandlers {
				handler := h // block scoping
				r.Handle(
					method,
					fmt.Sprintf("/plugins/%s/%s", pluginName, resourcePath),
					func(c *gin.Context) {
						// connect http request to plugin interface
						input := &core.ApiResourceInput{}
						if len(c.Params) > 0 {
							input.Params = make(map[string]string)
							for _, param := range c.Params {
								input.Params[param.Key] = param.Value
							}
						}
						input.Query = c.Request.URL.Query()
						if c.Request.Body != nil {
							err := c.ShouldBindJSON(&input.Body)
							if err != nil && err.Error() != "EOF" {
								shared.ApiOutputError(c, err, http.StatusBadRequest)
								return
							}
						}
						output, err := handler(input)
						if err != nil {
							shared.ApiOutputError(c, err, http.StatusBadRequest)
						} else if output != nil {
							status := output.Status
							if status < http.StatusContinue {
								status = http.StatusOK
							}
							shared.ApiOutputSuccess(c, output.Body, status)
						} else {
							shared.ApiOutputSuccess(c, nil, http.StatusOK)
						}
					},
				)
			}
		}
	}
}
