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

package cmd

import (
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/generator/util"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/stoewer/go-strcase"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func init() {
	rootCmd.AddCommand(createCollectorCmd)
}

func collectorNameNotExistValidateHoc(pluginName string) func(input string) error {
	collectorNameValidate := func(input string) error {
		if input == `` {
			return errors.Default.New("please input which data would you will collect (snake_format)")
		}
		snakeNameReg := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
		if !snakeNameReg.MatchString(input) {
			return errors.Default.New("collector name invalid (start with a-z and consist with a-z0-9_)")
		}
		_, err := os.Stat(filepath.Join(`plugins`, pluginName, `tasks`, input+`_collector.go`))
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		return errors.Default.New("collector exists")
	}
	return collectorNameValidate
}

func collectorNameExistValidateHoc(pluginName string) func(input string) error {
	collectorNameValidate := func(input string) error {
		if input == `` {
			return errors.Default.New("please input which data would you will collect (snake_format)")
		}
		_, err := os.Stat(filepath.Join(`plugins`, pluginName, `tasks`, input+`_collector.go`))
		return err
	}
	return collectorNameValidate
}

var createCollectorCmd = &cobra.Command{
	Use:   "create-collector [plugin_name] [collector_name]",
	Short: "Create a new collector",
	Long: `Create a new collector
Type in what the name of collector is, then generator will create a new collector in plugins/$plugin_name/tasks/$collector_name for you`,
	Run: func(cmd *cobra.Command, args []string) {
		var pluginName string
		var collectorName string
		var err error

		// try to get plugin name and collector name
		if len(args) > 0 {
			pluginName = args[0]
		}
		prompt := promptui.Prompt{
			Label:    "plugin_name",
			Validate: pluginNameExistValidate,
			Default:  pluginName,
		}
		pluginName, err = prompt.Run()
		cobra.CheckErr(err)
		pluginName = strings.ToLower(pluginName)

		if len(args) > 1 {
			collectorName = args[1]
		}
		prompt = promptui.Prompt{
			Label:    "collector_data_name",
			Validate: collectorNameNotExistValidateHoc(pluginName),
			Default:  collectorName,
		}
		collectorName, err = prompt.Run()
		cobra.CheckErr(err)
		collectorName = strings.ToLower(collectorName)

		prompt = promptui.Prompt{
			Label: "http_path",
			Validate: func(input string) error {
				if input == `` {
					return errors.BadInput.New("http_path required", errors.AsUserMessage())
				}
				if strings.HasPrefix(input, `/`) {
					return errors.BadInput.New("http_path shouldn't start with '/'", errors.AsUserMessage())
				}
				return nil
			},
		}
		httpPath, err := prompt.Run()
		cobra.CheckErr(err)

		// read template
		templates := map[string]string{
			collectorName + `_collector.go`: util.ReadTemplate("generator/template/plugin/tasks/api_collector.go-template"),
		}

		// create vars
		values := map[string]string{}
		util.GenerateAllFormatVar(values, `plugin_name`, pluginName)
		util.GenerateAllFormatVar(values, `collector_data_name`, collectorName)
		values[`HttpPath`] = httpPath
		collectorDataNameUpperCamel := strcase.UpperCamelCase(collectorName)
		values = util.DetectExistVars(templates, values)
		println(`vars in template:`, fmt.Sprint(values))

		// write template
		util.ReplaceVarInTemplates(templates, values)
		util.WriteTemplates(filepath.Join(`plugins`, pluginName, `tasks`), templates)
		if modifyExistCode {
			util.ReplaceVarInFile(
				filepath.Join(`plugins`, pluginName, `impl/impl.go`),
				regexp.MustCompile(`(return +\[]core\.SubTaskMeta ?\{ ?\n?)((\s*[\w.]+,\n)*)(\s*})`),
				fmt.Sprintf("$1$2\t\ttasks.Collect%sMeta,\n$4", collectorDataNameUpperCamel),
			)
			util.ReplaceVarInFile(
				filepath.Join(`plugins`, pluginName, fmt.Sprintf(`%s.go`, pluginName)),
				regexp.MustCompile(`(return +\[]core\.SubTaskMeta ?\{ ?\n?)((\s*[\w.]+,\n)*)(\s*})`),
				fmt.Sprintf("$1$2\t\ttasks.Collect%sMeta,\n$4", collectorDataNameUpperCamel),
			)
		}
	},
}
