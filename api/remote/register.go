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

package remote

import (
	"fmt"
	jsonmerge "github.com/RaveNoX/go-jsonmerge"
	"github.com/apache/incubator-devlake/api/shared"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/services/remote"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"github.com/swaggo/swag"
	"net/http"
)

var (
	vld        = validator.New()
	cachedDocs = map[string]*swag.Spec{}
)

type ApiResource struct {
	PluginName string
	Resources  map[string]map[string]core.ApiResourceHandler
}

// TODO add swagger doc
func RegisterPlugin(router *gin.Engine, registerEndpoints func(r *gin.Engine, pluginName string, apiResources map[string]map[string]core.ApiResourceHandler)) func(*gin.Context) {
	return func(c *gin.Context) {
		var details PluginDetails
		if err := c.ShouldBindJSON(&details); err != nil {
			shared.ApiOutputError(c, errors.BadInput.Wrap(err, shared.BadRequestBody))
			return
		}
		if err := vld.Struct(&details); err != nil {
			shared.ApiOutputError(c, errors.BadInput.Wrap(err, shared.BadRequestBody))
			return
		}
		plugin, err := remote.NewPlugin(&details.PluginInfo)
		if err != nil {
			shared.ApiOutputError(c, errors.Default.Wrap(err, "plugin could not be initialized"))
			return
		}
		resource := ApiResource{
			PluginName: details.PluginInfo.Name,
			Resources:  plugin.ApiResources(),
		}
		registerEndpoints(router, resource.PluginName, resource.Resources)
		registerSwagger(router, &details.Swagger)
		shared.ApiOutputSuccess(c, nil, http.StatusOK)
	}
}

func registerSwagger(router *gin.Engine, doc *SwaggerDoc) {
	if spec, ok := cachedDocs[doc.Name]; ok {
		spec.SwaggerTemplate = combineSpecs(spec.SwaggerTemplate, string(doc.Schema))
	} else {
		spec = &swag.Spec{
			Version:          "",
			Host:             "",
			BasePath:         "",
			Schemes:          nil,
			Title:            "",
			Description:      "",
			InfoInstanceName: doc.Name,
			SwaggerTemplate:  string(doc.Schema),
		}
		swag.Register(doc.Name, spec)
		cachedDocs[doc.Name] = spec
		router.GET(fmt.Sprintf("/plugins/swagger/%s/*any", doc.Resource), ginSwagger.CustomWrapHandler(
			&ginSwagger.Config{
				URL:                      "doc.json",
				DocExpansion:             "list",
				InstanceName:             doc.Name,
				Title:                    "",
				DefaultModelsExpandDepth: 1,
				DeepLinking:              true,
				PersistAuthorization:     false,
			},
			swaggerFiles.Handler))
	}
}

func combineSpecs(spec1 string, spec2 string) string {
	res, _ := jsonmerge.Merge(spec1, spec2)
	return res.(string)
}
