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

package experimental

import (
	"bytes"
	"context"
	"encoding/json"
	goerror "errors"
	"fmt"
	"github.com/apache/incubator-devlake/api"
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/impl"
	"github.com/apache/incubator-devlake/impl/dalgorm"
	"github.com/apache/incubator-devlake/impl/migration"
	"github.com/apache/incubator-devlake/logger"
	"github.com/apache/incubator-devlake/models"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"reflect"
	"syscall"
	"testing"
	"time"
)

var throwawayDir string

func init() {
	tempDir, err := errors.Convert01(os.MkdirTemp("", "devlake_test"+"_*"))
	if err != nil {
		panic(err)
	}
	throwawayDir = tempDir
}

// DevlakeSandbox FIXME
type DevlakeSandbox struct {
	Endpoint string
	db       *gorm.DB
	log      core.Logger
	cfg      *viper.Viper
	testCtx  *testing.T
}

// ServerSandboxConfig FIXME
type ServerSandboxConfig struct {
	ServerPort           uint
	DbURL                string
	CreateServer         bool
	DropDb               bool
	Plugins              map[string]core.PluginMeta
	DynamicPlugins       []string
	ForceCompilePlugins  bool
	AdditionalMigrations func() []core.MigrationScript
}

// ExistingSandboxConfig FIXME
type ExistingSandboxConfig struct {
	Endpoint string
}

// ConnectExistingServer FIXME
func ConnectExistingServer(t *testing.T, sbConfig *ExistingSandboxConfig) *DevlakeSandbox {
	return &DevlakeSandbox{
		Endpoint: sbConfig.Endpoint,
		db:       nil,
		log:      nil,
		testCtx:  t,
	}
}

// ConnectLocalServer FIXME
func ConnectLocalServer(t *testing.T, sbConfig *ServerSandboxConfig) *DevlakeSandbox {
	t.Helper()
	fmt.Printf("Using test temp directory: %s\n", throwawayDir)
	log := logger.Global.Nested("test")
	cfg := config.GetConfig()
	cfg.Set("DB_URL", sbConfig.DbURL)
	db, err := runner.NewGormDb(cfg, log)
	require.NoError(t, err)
	addr := fmt.Sprintf("http://localhost:%d", sbConfig.ServerPort)
	d := &DevlakeSandbox{
		Endpoint: addr,
		db:       db,
		log:      log,
		cfg:      cfg,
		testCtx:  t,
	}
	dynamicPluginDir := cfg.GetString("PLUGIN_DIR")
	d.initPlugins(dynamicPluginDir, sbConfig)
	if sbConfig.DropDb {
		d.dropDB()
	}
	if sbConfig.CreateServer {
		cfg.Set("PORT", fmt.Sprintf(":%d", sbConfig.ServerPort))
		cfg.Set("PLUGIN_DIR", throwawayDir)
		cfg.Set("LOGGING_DIR", throwawayDir)
		go api.CreateApiService()
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/proceed-db-migration", addr), nil)
		require.NoError(t, err)
		d.forceSendHttpRequest(10, req, func(err errors.Error) bool {
			e := err.Unwrap()
			return goerror.Is(e, syscall.ECONNREFUSED)
		})
	}
	d.runMigrations(sbConfig)
	return d
}

// CreateConnection FIXME
func (d *DevlakeSandbox) CreateConnection(dest string, testDest string, testCon interface{}, payload map[string]interface{}) *helper.BaseConnection {
	d.testCtx.Helper()
	_ = sendHttpRequest[any](d.testCtx, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPost, fmt.Sprintf("%s/plugins/%s", d.Endpoint, testDest), testCon)
	created := sendHttpRequest[helper.BaseConnection](d.testCtx, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPost, fmt.Sprintf("%s/plugins/%s", d.Endpoint, dest), payload)
	return &created
}

// ListConnections FIXME
func (d *DevlakeSandbox) ListConnections(dest string) []*helper.BaseConnection {
	d.testCtx.Helper()
	all := sendHttpRequest[[]*helper.BaseConnection](d.testCtx, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodGet, fmt.Sprintf("%s/plugins/%s", d.Endpoint, dest), nil)
	return all
}

// CreateBasicBlueprint FIXME
func (d *DevlakeSandbox) CreateBasicBlueprint(name string, connection *core.BlueprintConnectionV100) models.Blueprint {
	settings := &models.BlueprintSettings{
		Version:     "1.0.0",
		Connections: ToJson([]*core.BlueprintConnectionV100{connection}),
	}
	blueprint := models.Blueprint{
		Name:       name,
		Mode:       models.BLUEPRINT_MODE_NORMAL,
		Plan:       nil,
		Enable:     true,
		CronConfig: "manual",
		IsManual:   true,
		Settings:   ToJson(settings),
	}
	d.testCtx.Helper()
	blueprint = sendHttpRequest[models.Blueprint](d.testCtx, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPost, fmt.Sprintf("%s/blueprints", d.Endpoint), &blueprint)
	return blueprint
}

// TriggerBlueprint FIXME
func (d *DevlakeSandbox) TriggerBlueprint(blueprintId uint64) models.Pipeline {
	d.testCtx.Helper()
	pipeline := sendHttpRequest[models.Pipeline](d.testCtx, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPost, fmt.Sprintf("%s/blueprints/%d/trigger", d.Endpoint, blueprintId), nil)
	return d.monitorPipeline(pipeline.ID)
}

// RunPipeline FIXME
func (d *DevlakeSandbox) RunPipeline(pipeline models.NewPipeline) models.Pipeline {
	d.testCtx.Helper()
	pipelineResult := sendHttpRequest[models.Pipeline](d.testCtx, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPost, fmt.Sprintf("%s/pipelines", d.Endpoint), &pipeline)
	return d.monitorPipeline(pipelineResult.ID)
}

// MonitorPipeline FIXME
func (d *DevlakeSandbox) monitorPipeline(id uint64) models.Pipeline {
	d.testCtx.Helper()
	var previousResult models.Pipeline
	endpoint := fmt.Sprintf("%s/pipelines/%d", d.Endpoint, id)
	coloredPrintf("calling:\n\t%s %s\nwith:\n%s\n", http.MethodGet, endpoint, string(ToCleanJson(false, nil)))
	for {
		time.Sleep(1 * time.Second)
		pipelineResult := sendHttpRequest[models.Pipeline](d.testCtx, debugInfo{
			print: false,
		}, http.MethodGet, fmt.Sprintf("%s/pipelines/%d", d.Endpoint, id), nil)
		if pipelineResult.Status != models.TASK_RUNNING {
			return pipelineResult
		}
		if !reflect.DeepEqual(pipelineResult, previousResult) {
			coloredPrintf("result: %s\n", ToCleanJson(true, &pipelineResult))
		}
		previousResult = pipelineResult
	}
}

// RunPlugin FIXME
func (d *DevlakeSandbox) RunPlugin(ctx context.Context, cmd *cobra.Command, pluginTask core.PluginTask, options map[string]interface{}, subtaskNames ...string) errors.Error {
	return runner.RunPluginSubTasks(
		ctx,
		d.cfg,
		d.log,
		d.db,
		0,
		cmd.Use,
		subtaskNames,
		options,
		pluginTask,
		nil,
	)
}

// GetSubtaskNames FIXME
func GetSubtaskNames(metas ...core.SubTaskMeta) []string {
	var names []string
	for _, m := range metas {
		names = append(names, m.Name)
	}
	return names
}

func (d *DevlakeSandbox) forceSendHttpRequest(retries uint, req *http.Request, onError func(err errors.Error) bool) {
	d.testCtx.Helper()
	for {
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			if !onError(errors.Default.WrapRaw(err)) {
				require.NoError(d.testCtx, err)
			}
		} else {
			if res.StatusCode != http.StatusOK {
				panic(fmt.Sprintf("received HTTP status %d", res.StatusCode))
			}
			return
		}
		retries--
		if retries == 0 {
			panic("retry limit exceeded")
		}
		fmt.Printf("retrying http call to %s\n", req.URL.String())
		time.Sleep(1 * time.Second)
	}
}

func (d *DevlakeSandbox) initPlugins(dynamicPluginDir string, sbConfig *ServerSandboxConfig) {
	d.testCtx.Helper()
	dynamicallyLoad(d, sbConfig.ForceCompilePlugins, dynamicPluginDir, sbConfig.DynamicPlugins...)
	if sbConfig.Plugins != nil {
		for name, plugin := range sbConfig.Plugins {
			require.NoError(d.testCtx, core.RegisterPlugin(name, plugin))
		}
	}
	for _, p := range core.AllPlugins() {
		if pi, ok := p.(core.PluginInit); ok {
			err := pi.Init(d.cfg, d.log, d.db)
			require.NoError(d.testCtx, err)
		}
	}
}

func (d *DevlakeSandbox) runMigrations(sbConfig *ServerSandboxConfig) {
	d.testCtx.Helper()
	basicRes := impl.NewDefaultBasicRes(d.cfg, d.log, dalgorm.NewDalgorm(d.db))
	getMigrator := func() core.Migrator {
		migrator, err := migration.NewMigrator(basicRes)
		require.NoError(d.testCtx, err)
		return migrator
	}
	{
		migrator := getMigrator()
		for pluginName, pluginInst := range sbConfig.Plugins {
			if migratable, ok := pluginInst.(core.PluginMigration); ok {
				migrator.Register(migratable.MigrationScripts(), pluginName)
			}
		}
		require.NoError(d.testCtx, migrator.Execute())
	}
	{
		migrator := getMigrator()
		if sbConfig.AdditionalMigrations != nil {
			scripts := sbConfig.AdditionalMigrations()
			migrator.Register(scripts, "extra migrations")
		}
		require.NoError(d.testCtx, migrator.Execute())
	}
}

func (d *DevlakeSandbox) dropDB() {
	d.testCtx.Helper()
	migrator := d.db.Migrator()
	tables, err := migrator.GetTables()
	require.NoError(d.testCtx, err)
	var tablesRaw []any
	for _, table := range tables {
		tablesRaw = append(tablesRaw, table)
	}
	err = migrator.DropTable(tablesRaw...)
	require.NoError(d.testCtx, err)
}

// AddToPath FIXME
func AddToPath(newPaths ...string) {
	path := os.ExpandEnv("$PATH")
	for _, newPath := range newPaths {
		path = fmt.Sprintf("%s:%s", newPath, path)
	}
	_ = os.Setenv("PATH", path)
}

func sendHttpRequest[Res any](t *testing.T, debug debugInfo, httpMethod string, endpoint string, body any) Res {
	t.Helper()
	b := ToJson(body)
	if debug.print {
		coloredPrintf("calling:\n\t%s %s\nwith:\n%s\n", httpMethod, endpoint, string(ToCleanJson(debug.inlineJson, body)))
	}
	request, err := http.NewRequest(httpMethod, endpoint, bytes.NewReader(b))
	require.NoError(t, err)
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	require.True(t, response.StatusCode < 300, "unexpected http status code: %d", response.StatusCode)
	var result Res
	b, _ = io.ReadAll(response.Body)
	require.NoError(t, json.Unmarshal(b, &result))
	if debug.print {
		coloredPrintf("result: %s\n", ToCleanJson(debug.inlineJson, b))
	}
	require.NoError(t, response.Body.Close())
	return result
}

func coloredPrintf(msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	colorifier := "\033[1;33m%+v\033[0m" //yellow
	fmt.Printf(colorifier, msg)
}

type debugInfo struct {
	print      bool
	inlineJson bool
}
