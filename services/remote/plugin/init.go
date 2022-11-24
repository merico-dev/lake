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
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/services/remote/bridge"
	"github.com/apache/incubator-devlake/services/remote/models"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	connectionHelper *helper.ConnectionApiHelper
	basicRes         core.BasicRes
)

func Init(vld *validator.Validate, config *viper.Viper, logger core.Logger, database *gorm.DB) {
	basicRes = helper.NewDefaultBasicRes(config, logger, database)
	connectionHelper = helper.NewConnectionHelper(
		basicRes,
		vld,
	)
}

func NewPlugin(info *models.PluginInfo) (models.RemotePlugin, errors.Error) {
	var invoker bridge.Invoker
	var plugin models.RemotePlugin
	if info.Type == models.PythonCmd {
		invoker = bridge.NewPythonCmdInvoker(info.PluginPath)
	} else if info.Type == models.PythonPoetryCmd {
		invoker = bridge.NewPythonPoetryCmdInvoker(info.PluginPath)
	} else {
		return nil, errors.BadInput.New("unsupported plugin type")
	}
	if info.Extension == models.None {
		plugin = newPlugin(info, invoker)
		return plugin, nil
	}
	if info.Extension == models.Metric {
		plugin = newMetricPlugin(info, invoker)
		return plugin, nil
	}
	if info.Extension == models.Datasource {
		plugin = newDatasourcePlugin(info, invoker)
		return plugin, nil
	}
	return nil, errors.BadInput.New("unsupported plugin extension")
}
