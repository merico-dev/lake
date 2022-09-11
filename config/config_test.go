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
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestWriteConfig(t *testing.T) {
	filename := ".env"
	cwd, _ := os.Getwd()
	envFilePath := cwd + string(os.PathSeparator)
	os.Setenv("ENV_PATH", envFilePath)
	v := GetConfig()
	newDbUrl := "mysql://merico:merico@mysql:3307/lake?charset=utf8mb4&parseTime=True"
	v.Set("DB_URL", newDbUrl)
	fs := afero.NewOsFs()
	file, _ := fs.Create(filename)
	defer file.Close()
	_ = WriteConfig(v)
	isEmpty, _ := afero.IsEmpty(fs, filename)
	assert.False(t, isEmpty)
	err := fs.Remove(filename)
	assert.Equal(t, err == nil, true)
}

func TestWriteConfigAs(t *testing.T) {
	filename := ".env"
	v := GetConfig()
	newDbUrl := "mysql://merico:merico@mysql:3307/lake?charset=utf8mb4&parseTime=True"
	v.Set("DB_URL", newDbUrl)
	fs := afero.NewOsFs()
	file, _ := fs.Create(filename)
	defer file.Close()
	_ = WriteConfigAs(v, filename)
	isEmpty, _ := afero.IsEmpty(fs, filename)
	assert.False(t, isEmpty)
	err := fs.Remove(filename)
	assert.Equal(t, err == nil, true)
}

func TestSetConfigVariate(t *testing.T) {
	v := GetConfig()
	newDbUrl := "mysql://merico:merico@mysql:3307/lake?charset=utf8mb4&parseTime=True"
	v.Set("DB_URL", newDbUrl)
	currentDbUrl := v.GetString("DB_URL")
	logrus.Infof("current db url: %s\n", currentDbUrl)
	assert.Equal(t, currentDbUrl == newDbUrl, true)
}

func TestReplaceNewEnvItemInOldContent(t *testing.T) {
	v := GetConfig()
	v.Set(`aa`, `aaaa`)
	v.Set(`bb`, `1#1`)
	v.Set(`cc`, `1"'1`)
	v.Set(`dd`, `1\"1`)
	v.Set(`ee`, `=`)
	v.Set(`ff`, 1.01)
	v.Set(`gGg`, `gggg`)
	v.Set(`h.278`, 278)
	s, err := replaceNewEnvItemInOldContent(v, `
some unuseful message
# comment
a blank
 AA =123
bB=
  cc	=
  dd	 =
# some comment
eE=
ff="some content" and some comment
Ggg=132
h.278=1
`)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, `
some unuseful message
# comment
a blank
AA="aaaa"
BB="1#1"
CC="1\"'1"
DD="1\\\"1"
# some comment
EE="="
FF="1.01"
GGG="gggg"
H.278="278"
`, s)
}
