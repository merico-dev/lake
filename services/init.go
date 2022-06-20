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

package services

import (
	"context"

	"github.com/apache/incubator-devlake/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/core"

	"time"

	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/logger"
	"github.com/apache/incubator-devlake/migration"
	"github.com/apache/incubator-devlake/runner"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var cfg *viper.Viper
var db *gorm.DB
var cronManager *cron.Cron
var log core.Logger

func init() {
	var err error
	cfg = config.GetConfig()
	log = logger.Global
	db, err = runner.NewGormDb(cfg, logger.Global.Nested("db"))
	location := cron.WithLocation(time.UTC)
	cronManager = cron.New(location)
	if err != nil {
		panic(err)
	}
	migration.Init(db)
	runner.RegisterMigrationScripts(migrationscripts.All(), "Framework", cfg, logger.Global)
	// load plugins
	err = runner.LoadPlugins(
		cfg.GetString("PLUGIN_DIR"),
		cfg,
		logger.Global.Nested("plugin"),
		db,
	)
	if err != nil {
		panic(err)
	}
	err = migration.Execute(context.Background())
	if err != nil {
		panic(err)
	}

	// call service init
	pipelineServiceInit()
}
