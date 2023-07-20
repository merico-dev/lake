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

import (
	"reflect"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/models/common"
	"github.com/apache/incubator-devlake/core/plugin"
)

type PluginExtension string

const (
	None       PluginExtension = ""
	Metric     PluginExtension = "metric"
	Datasource PluginExtension = "datasource"
)

type PluginInfo struct {
	Name                 string                  `json:"name" validate:"required"`
	Description          string                  `json:"description"`
	ConnectionModelInfo  *DynamicModelInfo       `json:"connection_model_info" validate:"required"`
	ScopeConfigModelInfo *DynamicModelInfo       `json:"scope_config_model_info"`
	ScopeModelInfo       *DynamicModelInfo       `json:"scope_model_info" validate:"required"`
	ToolModelInfos       []*DynamicModelInfo     `json:"tool_model_infos"`
	MigrationScripts     []RemoteMigrationScript `json:"migration_scripts"`
	PluginPath           string                  `json:"plugin_path" validate:"required"`
	SubtaskMetas         []SubtaskMeta           `json:"subtask_metas"`
	Extension            PluginExtension         `json:"extension"`
}

// Type aliases used by the API helper for better readability
type (
	RemoteScopeConfig any
	RemoteConnection  any
)

type DynamicModelInfo struct {
	JsonSchema map[string]any `json:"json_schema" validate:"required"`
	TableName  string         `json:"table_name" validate:"required"`
}

func (d DynamicModelInfo) LoadDynamicTabler(parentModel any) (models.DynamicTabler, errors.Error) {
	return LoadTableModel(d.TableName, d.JsonSchema, parentModel)
}

type ScopeModel struct {
	common.NoPKModel `swaggerignore:"true"`
	Id               string `gorm:"primarykey;type:varchar(255)" json:"id"`
	ConnectionId     uint64 `gorm:"primaryKey" json:"connectionId"`
	Name             string `json:"name" validate:"required"`
	ScopeConfigId    uint64 `json:"scopeConfigId"`
}

type ApiParams struct {
	ConnectionId uint64
	ScopeId      string
}

type DynamicScopeModel struct {
	models.DynamicTabler
}

func NewDynamicScopeModel(model models.DynamicTabler) *DynamicScopeModel {
	return &DynamicScopeModel{
		DynamicTabler: model.New(),
	}
}

func (d DynamicScopeModel) ConnectionId() uint64 {
	return reflect.ValueOf(d.DynamicTabler.Unwrap()).Elem().FieldByName("ConnectionId").Uint()
}

func (d DynamicScopeModel) ScopeId() string {
	return reflect.ValueOf(d.DynamicTabler.Unwrap()).Elem().FieldByName("Id").String()
}

func (d DynamicScopeModel) ScopeName() string {
	return reflect.ValueOf(d.DynamicTabler.Unwrap()).Elem().FieldByName("Name").String()
}

func (d DynamicScopeModel) ScopeParams() interface{} {
	return &ApiParams{
		ConnectionId: d.ConnectionId(),
		ScopeId:      d.ScopeId(),
	}
}

func (d *DynamicScopeModel) MarshalJSON() ([]byte, error) {
	return d.DynamicTabler.MarshalJSON()
}

func (d *DynamicScopeModel) UnmarshalJSON(b []byte) error {
	return d.DynamicTabler.UnmarshalJSON(b)
}

type ScopeConfigModel struct {
	common.ScopeConfig `mapstructure:",squash"`
	ConnectionId       uint64 `json:"connectionId"`
	Name               string `json:"name"`
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

type DynamicDomainScope struct {
	TypeName string `json:"type_name"`
	Data     string `json:"data"`
}

type PipelineData struct {
	Plan   plugin.PipelinePlan  `json:"plan"`
	Scopes []DynamicDomainScope `json:"scopes"`
}

type ToolModel struct {
	common.NoPKModel
	ConnectionId uint64 `json:"connectionId" gorm:"column:connection_id;not null"`
}

var _ plugin.ToolLayerScope = (*DynamicScopeModel)(nil)
