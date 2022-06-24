package cmd

import (
	"errors"
	"fmt"
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
	rootCmd.AddCommand(createExtractorCmd)
}

func extractorNameNotExistValidateHoc(pluginName string) func(input string) error {
	extractorNameValidate := func(input string) error {
		if input == `` {
			return errors.New("please input which data would you will extract (snake_format)")
		}
		snakeNameReg := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
		if !snakeNameReg.MatchString(input) {
			return errors.New("extractor name invalid (start with a-z and consist with a-z0-9_)")
		}
		_, err := os.Stat(filepath.Join(`plugins`, pluginName, `tasks`, input+`_extractor.go`))
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		return errors.New("extractor exists")
	}
	return extractorNameValidate
}

var createExtractorCmd = &cobra.Command{
	Use:   "create-extractor [plugin_name] [extractor_name]",
	Short: "Create a new extractor",
	Long: `Create a new extractor
Type in what the name of extractor is, then generator will create a new extractor in plugins/$plugin_name/tasks/$extractor_name for you`,
	Run: func(cmd *cobra.Command, args []string) {
		var pluginName string
		var extractorName string
		var err error

		// try to get plugin name and extractor name
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

		prompt = promptui.Prompt{
			Label:    "collector_name",
			Validate: collectorNameExistValidateHoc(pluginName),
		}
		collectorName, err := prompt.Run()
		cobra.CheckErr(err)
		collectorName = strings.ToLower(collectorName)

		if len(args) > 1 {
			extractorName = args[1]
		}
		prompt = promptui.Prompt{
			Label:    "extractor_name",
			Validate: extractorNameNotExistValidateHoc(pluginName),
			Default:  extractorName,
		}
		extractorName, err = prompt.Run()
		cobra.CheckErr(err)
		extractorName = strings.ToLower(extractorName)

		// read template
		templates := map[string]string{
			extractorName + `_extractor.go`: util.ReadTemplate("generator/template/plugin/tasks/extractor.go-template"),
		}

		// create vars
		values := map[string]string{}
		util.GenerateAllFormatVar(values, `plugin_name`, pluginName)
		util.GenerateAllFormatVar(values, `collector_data_name`, collectorName)
		util.GenerateAllFormatVar(values, `extractor_data_name`, extractorName)
		extractorDataNameUpperCamel := strcase.UpperCamelCase(extractorName)
		values = util.DetectExistVars(templates, values)
		println(`vars in template:`, fmt.Sprint(values))

		// write template
		util.ReplaceVarInTemplates(templates, values)
		util.WriteTemplates(filepath.Join(`plugins`, pluginName, `tasks`), templates)
		if modifyExistCode {
			util.ReplaceVarInFile(
				filepath.Join(`plugins`, pluginName, `plugin_main.go`),
				regexp.MustCompile(`(return +\[]core\.SubTaskMeta ?\{ ?\n?)((\s*[\w.]+,\n)*)(\s*})`),
				fmt.Sprintf("$1$2\t\ttasks.Extract%sMeta,\n$4", extractorDataNameUpperCamel),
			)
		}
	},
}
