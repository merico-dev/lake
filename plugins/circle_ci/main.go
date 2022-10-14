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

package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/runner"
	"github.com/apache/incubator-devlake/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// Make sure interfaces are implemented
var _ core.PluginMeta = (*CircleCIPlugin)(nil)
var _ core.PluginInit = (*CircleCIPlugin)(nil)
var _ core.PluginTask = (*CircleCIPlugin)(nil)
var _ core.PluginApi = (*CircleCIPlugin)(nil)

// PluginEntry exports a symbol loaded by the framework
var PluginEntry CircleCIPlugin //nolint

type (
	CircleCIPlugin         struct{}
	CircleCIPluginTaskData struct {
		Options map[string]interface{}
	}
)

func (plugin CircleCIPlugin) Description() string {
	return "Wraps a plugin implemented in python"
}

// Store configuration
var configuration map[string]interface{}

func (plugin CircleCIPlugin) Init(config *viper.Viper, logger core.Logger, db *gorm.DB) errors.Error {
	configuration = config.AllSettings()
	return nil
}

func CmdEntryPoint(cmd string, args ...string) core.SubTaskEntryPoint {
	return func(ctx core.SubTaskContext) errors.Error {
		logger := ctx.GetLogger()

		cmd := exec.Command(cmd, args...)

		_, err := utils.StreamProcess(cmd, func(b []byte) (string, error) {
			log := string(b[:])
			level, message, valid := strings.Cut(log, ": ")

			if !valid {
				message = fmt.Sprintf("Invalid log format: %s", log)
				logger.Log(core.LOG_WARN, message)
			}

			switch level {
			case "DEBUG":
				logger.Log(core.LOG_DEBUG, message)
			case "INFO":
				logger.Log(core.LOG_INFO, message)
			case "WARN":
				logger.Log(core.LOG_WARN, message)
			case "ERROR":
				logger.Log(core.LOG_ERROR, message)
			default:
				message = fmt.Sprintf("Invalid log level %s for message: %s", level, message)
				logger.Log(core.LOG_WARN, message)
			}
			return log, nil
		})

		if err != nil {
			return errors.Default.Wrap(err, "error starting process stream from singer-tap")
		}

		return nil
	}
}

func PyEntryPoint(args ...string) core.SubTaskEntryPoint {
	return func(ctx core.SubTaskContext) errors.Error {
		dburl := ctx.GetConfig("db_url")
		args = append(args, "--db-url", dburl)
		return CmdEntryPoint("/Users/camille/Documents/Merico/incubator-devlake/plugins/circleci/circleci", args...)(ctx)
	}
}

func (plugin CircleCIPlugin) SubTaskMetas() []core.SubTaskMeta {
	subtask_meta := core.SubTaskMeta{
		Name:             "extract dummy",
		EntryPoint:       PyEntryPoint("extract", "dummy"),
		EnabledByDefault: true,
		Description:      "extract POC",
		DomainTypes:      []string{},
	}

	return []core.SubTaskMeta{
		subtask_meta,
	}
}

func (plugin CircleCIPlugin) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	return &CircleCIPluginTaskData{
		Options: options,
	}, nil
}

// PkgPath information lost when compiled as plugin(.so)
func (plugin CircleCIPlugin) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/circleci"
}

func (plugin CircleCIPlugin) ApiResources() map[string]map[string]core.ApiResourceHandler {
	// TODO
	return nil
}

// standalone mode for debugging
func main() {
	cmd := &cobra.Command{Use: "CircleCIPlugin"}

	// TODO add your cmd flag if necessary
	// yourFlag := cmd.Flags().IntP("yourFlag", "y", 8, "TODO add description here")
	// _ = cmd.MarkFlagRequired("yourFlag")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		runner.DirectRun(cmd, args, PluginEntry, map[string]interface{}{
			// TODO add more custom params here
		})
	}
	runner.RunCmd(cmd)
}
