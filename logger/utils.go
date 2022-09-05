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
	"fmt"
	"github.com/apache/incubator-devlake/models"
	"github.com/apache/incubator-devlake/plugins/core"
	"os"
	"path/filepath"
)

func GetTaskLoggerPath(config *core.LoggerConfig, t *models.Task) string {
	if config.Path == "" {
		return ""
	}
	info, err := os.Stat(config.Path)
	if err != nil {
		panic(err)
	}
	basePath := config.Path
	if !info.IsDir() {
		basePath = filepath.Dir(config.Path)
	}
	return filepath.Join(basePath, fmt.Sprintf("task-%d-%d-%d-%s.log", t.ID, t.PipelineRow, t.PipelineCol, t.Plugin))
}

func GetPipelineLoggerPath(config *core.LoggerConfig, p *models.ApiPipeline) string {
	if config.Path == "" {
		return ""
	}
	info, err := os.Stat(config.Path)
	if err != nil {
		panic(err)
	}
	basePath := config.Path
	if !info.IsDir() {
		basePath = filepath.Dir(config.Path)
	}
	formattedCreationTime := p.CreatedAt.UTC().Format("20060102-1504")
	return filepath.Join(basePath, fmt.Sprintf("pipeline-%d-%s", p.ID, formattedCreationTime), "pipeline.log")
}
