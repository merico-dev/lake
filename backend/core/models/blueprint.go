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
	"encoding/json"
	"time"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/common"
	"github.com/apache/incubator-devlake/core/plugin"
)

const (
	BLUEPRINT_MODE_NORMAL   = "NORMAL"
	BLUEPRINT_MODE_ADVANCED = "ADVANCED"
)

// @Description CronConfig
type Blueprint struct {
	Name        string          `json:"name" validate:"required"`
	ProjectName string          `json:"projectName" gorm:"type:varchar(255)"`
	Mode        string          `json:"mode" gorm:"varchar(20)" validate:"required,oneof=NORMAL ADVANCED"`
	Plan        json.RawMessage `json:"plan" gorm:"serializer:encdec"`
	Enable      bool            `json:"enable"`
	//please check this https://crontab.guru/ for detail
	CronConfig   string          `json:"cronConfig" format:"* * * * *" example:"0 0 * * 1"`
	IsManual     bool            `json:"isManual"`
	SkipOnFail   bool            `json:"skipOnFail"`
	Labels       []string        `json:"labels" gorm:"-"`
	Settings     json.RawMessage `json:"settings" swaggertype:"array,string" example:"please check api: /blueprints/<PLUGIN_NAME>/blueprint-setting" gorm:"serializer:encdec"`
	common.Model `swaggerignore:"true"`
}

type BlueprintSettings struct {
	Version     string          `json:"version" validate:"required,semver,oneof=1.0.0"`
	TimeAfter   *time.Time      `json:"timeAfter"`
	Connections json.RawMessage `json:"connections" validate:"required"`
	BeforePlan  json.RawMessage `json:"before_plan"`
	AfterPlan   json.RawMessage `json:"after_plan"`
}

// UpdateConnections unmarshals the connections on this BlueprintSettings
func (bps *BlueprintSettings) UnmarshalConnections() ([]*plugin.BlueprintConnectionV200, errors.Error) {
	var connections []*plugin.BlueprintConnectionV200
	err := json.Unmarshal(bps.Connections, &connections)
	if err != nil {
		return nil, errors.Default.Wrap(err, `unmarshal connections fail`)
	}
	return connections, nil
}

// UpdateConnections updates the connections on this BlueprintSettings reference according to the updater function
func (bps *BlueprintSettings) UpdateConnections(updater func(c *plugin.BlueprintConnectionV200) errors.Error) errors.Error {
	conns, err := bps.UnmarshalConnections()
	if err != nil {
		return err
	}
	for i, conn := range conns {
		err = updater(conn)
		if err != nil {
			return err
		}
		conns[i] = conn
	}
	bps.Connections, err = errors.Convert01(json.Marshal(&conns))
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalPlan unmarshals Plan in JSON to strong-typed plugin.PipelinePlan
func (bp *Blueprint) UnmarshalPlan() (plugin.PipelinePlan, errors.Error) {
	var plan plugin.PipelinePlan
	err := errors.Convert(json.Unmarshal(bp.Plan, &plan))
	if err != nil {
		return nil, errors.Default.Wrap(err, `unmarshal plan fail`)
	}
	return plan, nil
}

// UnmarshalSettings unmarshals the BlueprintSettings on the Blueprint
func (bp *Blueprint) UnmarshalSettings() (BlueprintSettings, errors.Error) {
	var settings BlueprintSettings
	err := errors.Convert(json.Unmarshal(bp.Settings, &settings))
	if err != nil {
		return settings, errors.Default.Wrap(err, `unmarshal settings fail`)
	}
	return settings, nil
}

// GetConnections Gets all the blueprint connections for this blueprint
func (bp *Blueprint) GetConnections() ([]*plugin.BlueprintConnectionV200, errors.Error) {
	settings, err := bp.UnmarshalSettings()
	if err != nil {
		return nil, err
	}
	conns, err := settings.UnmarshalConnections()
	if err != nil {
		return nil, err
	}
	return conns, nil
}

// UpdateSettings updates the blueprint instance with this settings reference
func (bp *Blueprint) UpdateSettings(settings *BlueprintSettings) errors.Error {
	if settings.Connections == nil {
		bp.Settings = nil
	} else {
		settingsRaw, err := errors.Convert01(json.Marshal(settings))
		if err != nil {
			return err
		}
		bp.Settings = settingsRaw
	}
	return nil
}

// GetScopes Gets all the scopes across all the connections for this blueprint
func (bp *Blueprint) GetScopes() ([]*plugin.BlueprintScopeV200, errors.Error) {
	conns, err := bp.GetConnections()
	if err != nil {
		return nil, err
	}
	visited := map[string]any{}
	var result []*plugin.BlueprintScopeV200
	for _, conn := range conns {
		for _, scope := range conn.Scopes {
			if _, ok := visited[scope.Id]; !ok {
				result = append(result, scope)
				visited[scope.Id] = true
			}
		}
	}
	return result, nil
}

func (Blueprint) TableName() string {
	return "_devlake_blueprints"
}

type DbBlueprintLabel struct {
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	BlueprintId uint64    `json:"blueprint_id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"primaryKey;index"`
}

func (DbBlueprintLabel) TableName() string {
	return "_devlake_blueprint_labels"
}
