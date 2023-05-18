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
	"strings"

	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/server/api/blueprints"
	"github.com/apache/incubator-devlake/server/api/domainlayer"
	"github.com/apache/incubator-devlake/server/api/pipelines"
	"github.com/apache/incubator-devlake/server/api/plugininfo"
	"github.com/apache/incubator-devlake/server/api/project"
	"github.com/apache/incubator-devlake/server/api/push"
	"github.com/apache/incubator-devlake/server/api/shared"
	"github.com/apache/incubator-devlake/server/api/task"
	"github.com/apache/incubator-devlake/server/services"

	"github.com/gin-gonic/gin"
)

func RegisterRouter(r *gin.Engine) {
	r.GET("/pipelines", pipelines.Index)
	r.POST("/pipelines", pipelines.Post)
	r.GET("/pipelines/:pipelineId", pipelines.Get)
	r.PATCH("/blueprints/:blueprintId", blueprints.Patch)
	r.POST("/blueprints/:blueprintId/trigger", blueprints.Trigger)
	r.DELETE("/blueprints/:blueprintId", blueprints.Delete)

	r.GET("/blueprints", blueprints.Index)
	r.POST("/blueprints", blueprints.Post)
	r.GET("/blueprints/:blueprintId", blueprints.Get)
	r.GET("/blueprints/:blueprintId/pipelines", blueprints.GetBlueprintPipelines)
	r.DELETE("/pipelines/:pipelineId", pipelines.Delete)
	r.GET("/pipelines/:pipelineId/tasks", task.GetTaskByPipeline)
	r.POST("/pipelines/:pipelineId/rerun", pipelines.PostRerun)
	r.POST("/tasks/:taskId/rerun", task.PostRerun)

	r.GET("/pipelines/:pipelineId/logging.tar.gz", pipelines.DownloadLogs)

	//r.GET("/ping", ping.Get)
	//r.GET("/version", version.Get)
	r.POST("/push/:tableName", push.Post)
	r.GET("/domainlayer/repos", domainlayer.ReposIndex)

	// plugin api
	r.GET("/plugininfo", plugininfo.Get)
	r.GET("/plugins", plugininfo.GetPluginMetas)

	// project api
	r.GET("/projects/*projectName", project.GetProject)
	r.PATCH("/projects/*projectName", project.PatchProject)
	//r.DELETE("/projects/:projectName", project.DeleteProject)
	r.POST("/projects", project.PostProject)
	r.GET("/projects", project.GetProjects)

	// mount all api resources for all plugins
	resources, err := services.GetPluginsApiResources()
	if err != nil {
		panic(err)
	}
	// mount all api resources for all plugins
	for pluginName, apiResources := range resources {
		registerPluginEndpoints(r, pluginName, apiResources)
	}
}

func registerPluginEndpoints(r *gin.Engine, pluginName string, apiResources map[string]map[string]plugin.ApiResourceHandler) {
	for resourcePath, resourceHandlers := range apiResources {
		for method, h := range resourceHandlers {
			r.Handle(
				method,
				fmt.Sprintf("/plugins/%s/%s", pluginName, resourcePath),
				handlePluginCall(pluginName, h),
			)
		}
	}
}

func handlePluginCall(pluginName string, handler plugin.ApiResourceHandler) func(c *gin.Context) {
	return func(c *gin.Context) {
		var err error
		input := &plugin.ApiResourceInput{}
		input.Params = make(map[string]string)
		if len(c.Params) > 0 {
			for _, param := range c.Params {
				input.Params[param.Key] = param.Value
			}
		}
		input.Params["plugin"] = pluginName
		input.Query = c.Request.URL.Query()
		if c.Request.Body != nil {
			if strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data;") {
				input.Request = c.Request
			} else {
				err = c.ShouldBindJSON(&input.Body)
				if err != nil && err.Error() != "EOF" {
					shared.ApiOutputError(c, err)
					return
				}
			}
		}
		output, err := handler(input)
		if err != nil {
			shared.ApiOutputError(c, err)
		} else if output != nil {
			status := output.Status
			if status < http.StatusContinue {
				status = http.StatusOK
			}
			if output.File != nil {
				c.Data(status, output.File.ContentType, output.File.Data)
				return
			}
			if blob, ok := output.Body.([]byte); ok && output.ContentType != "" {
				c.Data(status, output.ContentType, blob)
			} else {
				shared.ApiOutputSuccess(c, output.Body, status)
			}
		} else {
			shared.ApiOutputSuccess(c, nil, http.StatusOK)
		}
	}
}
