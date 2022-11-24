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

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apache/incubator-devlake/errors"
	goerror "github.com/cockroachdb/errors"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const defaultConfigName = ".env"

// Lowcase V for private this. You can use it by call GetConfig.
var v *viper.Viper

// GetConfig return a viper.Viper
func GetConfig() *viper.Viper {
	return v
}

func initConfig(v *viper.Viper) {
	v.SetConfigName(getConfigName())
	v.SetConfigType("env")
	envPath := getEnvPath()
	// AddConfigPath adds a path for Viper to search for the config file in.
	v.AddConfigPath("./../../..")
	v.AddConfigPath("./../..")
	v.AddConfigPath("./../")
	v.AddConfigPath("./")
	v.AddConfigPath(envPath)
}

func getConfigName() string {
	return defaultConfigName
}

// return the env path
func getEnvPath() string {
	envPath := os.Getenv("ENV_PATH")
	return filepath.Dir(envPath)
}

// Set default value for no .env or .env not set it
func setDefaultValue(v *viper.Viper) {
	v.SetDefault("DB_URL", "mysql://merico:merico@mysql:3306/lake?charset=utf8mb4&parseTime=True")
	v.SetDefault("PORT", ":8080")
	v.SetDefault("PLUGIN_DIR", "bin/plugins")
	v.SetDefault("TEMPORAL_TASK_QUEUE", "DEVLAKE_TASK_QUEUE")
	v.SetDefault("TAP_PROPERTIES_DIR", "config/tap")
	v.SetDefault("ENABLE_REMOTE_PLUGINS", "true")
	v.SetDefault("REMOTE_PLUGINS_STARTUP_PATH", "python/plugins/start.sh")
}

// replaceNewEnvItemInOldContent replace old config to new config in env file content
func replaceNewEnvItemInOldContent(v *viper.Viper, envFileContent string) (string, errors.Error) {
	// prepare reg exp
	encodeEnvNameReg := regexp.MustCompile(`[^a-zA-Z0-9]`)
	if encodeEnvNameReg == nil {
		return ``, errors.Default.New("encodeEnvNameReg err")
	}

	for _, key := range v.AllKeys() {
		envName := strings.ToUpper(key)
		val := v.Get(envName)
		encodeEnvName := encodeEnvNameReg.ReplaceAllStringFunc(envName, func(s string) string {
			return fmt.Sprintf(`\%v`, s)
		})
		envItemReg, err := regexp.Compile(fmt.Sprintf(`(?im)^\s*%v\s*\=.*$`, encodeEnvName))
		if err != nil {
			return ``, errors.Default.Wrap(err, "regexp Compile failed")
		}
		envFileContent = envItemReg.ReplaceAllStringFunc(envFileContent, func(s string) string {
			switch ret := val.(type) {
			case string:
				ret = strings.Replace(ret, `\`, `\\`, -1)
				//ret = strings.Replace(ret, `=`, `\=`, -1)
				//ret = strings.Replace(ret, `'`, `\'`, -1)
				ret = strings.Replace(ret, `"`, `\"`, -1)
				return fmt.Sprintf(`%v="%v"`, envName, ret)
			default:
				if val == nil {
					return fmt.Sprintf(`%v=`, envName)
				}
				return fmt.Sprintf(`%v="%v"`, envName, ret)
			}
		})
	}
	return envFileContent, nil
}

// WriteConfig save viper to .env file
func WriteConfig(v *viper.Viper) errors.Error {
	envPath := getEnvPath()
	fileName := getConfigName()

	if envPath != "" {
		fileName = envPath + string(os.PathSeparator) + fileName
	}

	return WriteConfigAs(v, fileName)
}

// WriteConfigAs save viper to custom filename
func WriteConfigAs(v *viper.Viper, filename string) errors.Error {
	aferoFile := afero.NewOsFs()
	fmt.Println("Attempting to write configuration to .env file.")
	var configType string

	ext := filepath.Ext(filename)
	if ext != "" {
		configType = ext[1:]
	}
	if configType != "env" && configType != "dotenv" {
		return errors.Convert(v.WriteConfigAs(filename))
	}

	// FIXME viper just have setter and have no getter so create new configPermissions and file
	flags := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	configPermissions := os.FileMode(0644)
	file, err := afero.ReadFile(aferoFile, filename)
	if err != nil && !goerror.Is(err, os.ErrNotExist) {
		return errors.Convert(err)
	}

	envFileContent := string(file)
	f, err := aferoFile.OpenFile(filename, flags, configPermissions)
	if err != nil {
		return errors.Convert(err)
	}
	defer f.Close()

	for _, key := range v.AllKeys() {
		envName := strings.ToUpper(key)
		if !strings.Contains(envFileContent, envName) {
			envFileContent = fmt.Sprintf("%s\n%s=", envFileContent, envName)
		}
	}
	envFileContent, err = replaceNewEnvItemInOldContent(v, envFileContent)
	if err != nil {
		return errors.Convert(err)
	}
	if _, err := f.WriteString(envFileContent); err != nil {
		return errors.Convert(err)
	}
	return errors.Convert(f.Sync())
}

func init() {
	// create the object and load the .env file
	v = viper.New()
	initConfig(v)
	err := v.ReadInConfig()
	if err != nil {
		logrus.Warn("Failed to read [.env] file:", err)
	}
	v.AutomaticEnv()

	setDefaultValue(v)
	// This line is essential for reading and writing
	v.WatchConfig()
}
