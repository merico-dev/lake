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

package plugin

import (
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/server/services/remote/bridge"
	"github.com/apache/incubator-devlake/server/services/remote/models"
	"github.com/go-playground/validator/v10"
)

var (
	connectionHelper *api.ConnectionApiHelper
	scopeHelper      *api.GenericScopeApiHelper[any, any, any]
	basicRes         context.BasicRes
	vld              *validator.Validate
)

func Init(br context.BasicRes) {
	vld = validator.New()
	basicRes = br
	connectionHelper = api.NewConnectionHelper(
		br,
		vld,
	)
}

func NewRemotePlugin(info *models.PluginInfo) (models.RemotePlugin, errors.Error) {
	invoker := bridge.NewCmdInvoker(info.PluginPath)
	plugin, err := newPlugin(info, invoker)

	if err != nil {
		return nil, err
	}

	switch info.Extension {
	case models.None:
		return plugin, nil
	case models.Metric:
		return &remoteMetricPlugin{plugin}, nil
	case models.Datasource:
		return &remoteDatasourcePlugin{plugin}, nil
	default:
		return nil, errors.BadInput.New("unsupported plugin extension")
	}
}
