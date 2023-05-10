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
	"github.com/apache/incubator-devlake/core/config"
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/log"
	"github.com/apache/incubator-devlake/core/models/migrationscripts"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/core/runner"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/services"
	"github.com/apache/incubator-devlake/impls/dalgorm"
	"github.com/apache/incubator-devlake/impls/logruslog"
	"github.com/go-playground/validator/v10"
	"github.com/robfig/cron/v3"
	"sync"
	"time"
)

var cfg config.ConfigReader
var logger log.Logger
var db dal.Dal

var bpManager *services.BlueprintManager
var basicRes context.BasicRes
var migrator plugin.Migrator
var cronManager *cron.Cron
var cronLocker sync.Mutex
var vld *validator.Validate

const failToCreateCronJob = "created cron job failed"

// InitResources creates resources needed by services module
func InitResources() {
	var err error

	// basic resources initialization
	vld = validator.New()
	basicRes = runner.CreateAppBasicRes()
	cfg = basicRes.GetConfigReader()
	logger = basicRes.GetLogger()
	db = basicRes.GetDal()
	bpManager = services.NewBlueprintManager(db)
	// initialize db migrator
	migrator, err = runner.InitMigrator(basicRes)
	if err != nil {
		panic(err)
	}
	logger.Info("migration initialized")
	migrator.Register(migrationscripts.All(), "Framework")
}

// GetBasicRes returns the context.BasicRes instance used by services module
func GetBasicRes() context.BasicRes {
	return basicRes
}

// GetMigrator returns the core.Migrator instance used by services module
func GetMigrator() plugin.Migrator {
	return migrator
}

// Init the services module
func Init() {
	InitResources()

	// lock the database to avoid multiple devlake instances from sharing the same one
	lockDb()

	var err error
	// now, load the plugins
	err = runner.LoadPlugins(basicRes)
	if err != nil {
		logger.Error(err, "failed to load plugins")
		panic(err)
	}

	// pull migration scripts from plugins to migrator
	for pluginName, pluginInst := range plugin.AllPlugins() {
		if migratable, ok := pluginInst.(plugin.PluginMigration); ok {
			migrator.Register(migratable.MigrationScripts(), pluginName)
		}
	}

	// check if there are pending migration
	forceMigration := cfg.GetBool("FORCE_MIGRATION")
	if !migrator.HasPendingScripts() || forceMigration {
		err = ExecuteMigration()
		if err != nil {
			panic(err)
		}
	}
	logger.Info("Db migration confirmation needed")
}

// ExecuteMigration executes all pending migration scripts and initialize services module
func ExecuteMigration() errors.Error {
	// apply all pending migration scripts
	err := migrator.Execute()
	if err != nil {
		return err
	}

	// cronjob for blueprint triggering
	location := cron.WithLocation(time.UTC)
	cronManager = cron.New(location)
	if err != nil {
		panic(err)
	}

	// initialize pipeline server, mainly to start the pipeline consuming process
	pipelineServiceInit()
	return nil
}

// MigrationRequireConfirmation returns if there were migration scripts waiting to be executed
func MigrationRequireConfirmation() bool {
	return migrator.HasPendingScripts()
}

func lockDb() {
	// gorm doesn't support creating a PrepareStmt=false session from a PrepareStmt=true
	// but the lockDatabase needs PrepareStmt=false for table locking, we have to deal with it here
	lockingDb, err := runner.NewGormDbEx(cfg, logruslog.Global.Nested("migrator db"), &dal.SessionConfig{
		PrepareStmt:            false,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic(err)
	}
	err = lockDatabase(dalgorm.NewDalgorm(lockingDb))
	if err != nil {
		panic(err)
	}
}
