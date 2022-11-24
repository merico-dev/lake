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
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/services/remote/models"
	remote "github.com/apache/incubator-devlake/services/remote/plugin"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	remotePlugins = make(map[string]models.RemotePlugin)
)

func Init(config *viper.Viper, logger core.Logger, database *gorm.DB) {
	remote.Init(nil, config, logger, database)
}

func NewPlugin(info *models.PluginInfo) (models.RemotePlugin, errors.Error) {
	if _, ok := remotePlugins[info.Name]; ok {
		return nil, errors.BadInput.New("plugin already registered")
	}
	plugin, err := remote.NewPlugin(info)
	if err != nil {
		return nil, errors.BadInput.New("unsupported plugin type")
	}
	forceMigration := config.GetConfig().GetBool("FORCE_MIGRATION")
	err = plugin.RunMigrations(forceMigration)
	if err != nil {
		return nil, err
	}
	err = core.RegisterPlugin(info.Name, plugin)
	if err != nil {
		return nil, err
	}
	remotePlugins[info.Name] = plugin
	return plugin, nil
}
