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

package models

const (
	PythonPoetryCmd PluginType      = "python-poetry"
	PythonCmd       PluginType      = "python"
	None            PluginExtension = ""
	Metric          PluginExtension = "metric"
	Datasource      PluginExtension = "datasource"
)

type (
	PluginType      string
	PluginExtension string
)

type PluginInfo struct {
	Type         PluginType      `json:"type" validate:"required"`
	Name         string          `json:"name" validate:"required"`
	Extension    PluginExtension `json:"extension"`
	Connection   map[string]any  `json:"connection" validate:"required"`
	Description  string          `json:"description"`
	PluginPath   string          `json:"plugin_path" validate:"required"`
	ApiEndpoints []Endpoint      `json:"api_endpoints" validate:"dive"`
	SubtaskMetas []SubtaskMeta   `json:"subtask_metas" validate:"dive"`
}

type SubtaskMeta struct {
	Name             string   `json:"name" validate:"required"`
	EntryPointName   string   `json:"entry_point_name" validate:"required"`
	Arguments        []string `json:"arguments"`
	Required         bool     `json:"required"`
	EnabledByDefault bool     `json:"enabled_by_default"`
	Description      string   `json:"description" validate:"required"`
	DomainTypes      []string `json:"domain_types" validate:"required"`
}

type Endpoint struct {
	Resource string `json:"resource" validate:"required"`
	Handler  string `json:"handler" validate:"required"`
	Method   string `json:"method" validate:"required"`
}
