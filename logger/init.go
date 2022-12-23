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

package logger

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var inner *logrus.Logger
var Global core.Logger

func init() {
	inner = logrus.New()
	logLevel := logrus.InfoLevel
	cfg := config.GetConfig()
	switch strings.ToLower(cfg.GetString("LOGGING_LEVEL")) {
	case "debug":
		logLevel = logrus.DebugLevel
	case "info":
		logLevel = logrus.InfoLevel
	case "warn":
		logLevel = logrus.WarnLevel
	case "error":
		logLevel = logrus.ErrorLevel
	}
	inner.SetLevel(logLevel)
	inner.SetFormatter(&prefixed.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	basePath := cfg.GetString("LOGGING_DIR")
	if basePath == "" {
		basePath = "./logs"
	}
	if abs, err := filepath.Abs(basePath); err != nil {
		panic(err)
	} else {
		basePath = filepath.Join(abs, "devlake.log")
	}
	var err errors.Error
	Global, err = NewDefaultLogger(inner)
	if err != nil {
		panic(err)
	}
	stream, err := GetFileStream(basePath)
	if err != nil {
		stream = os.Stdout
	}
	Global.SetStream(&core.LoggerStreamConfig{
		Path:   basePath,
		Writer: stream,
	})
	if err != nil {
		Global.Warn(err, "Failed to create filestream for logs. Logs will not be piped to files.")
	}
}
