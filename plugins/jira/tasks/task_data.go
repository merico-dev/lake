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

package tasks

import (
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"time"

	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/plugins/jira/models"
)

type StatusMapping struct {
	StandardStatus string `json:"standardStatus"`
}

type StatusMappings map[string]StatusMapping

type TypeMapping struct {
	StandardType   string         `json:"standardType"`
	StatusMappings StatusMappings `json:"statusMappings"`
}

type TypeMappings map[string]TypeMapping

type JiraTransformationRule struct {
	Name                       string       `gorm:"type:varchar(255)"`
	EpicKeyField               string       `json:"epicKeyField"`
	StoryPointField            string       `json:"storyPointField"`
	RemotelinkCommitShaPattern string       `json:"remotelinkCommitShaPattern"`
	TypeMappings               TypeMappings `json:"typeMappings"`
}

func (r *JiraTransformationRule) ToDb() (rule *models.JiraTransformationRule, error2 errors.Error) {
	blob, err := json.Marshal(r.TypeMappings)
	if err != nil {
		return nil, errors.Default.Wrap(err, "error marshaling TypeMappings")
	}
	return &models.JiraTransformationRule{
		Name:                       r.Name,
		EpicKeyField:               r.EpicKeyField,
		StoryPointField:            r.StoryPointField,
		RemotelinkCommitShaPattern: r.RemotelinkCommitShaPattern,
		TypeMappings:               blob,
	}, nil
}
func (r *JiraTransformationRule) FromDb(rule *models.JiraTransformationRule) (*JiraTransformationRule, errors.Error) {
	mappings := make(map[string]TypeMapping)
	err := json.Unmarshal(rule.TypeMappings, &mappings)
	if err != nil {
		return nil, errors.Default.Wrap(err, "error marshaling TypeMappings")
	}
	r.Name = rule.Name
	r.EpicKeyField = rule.EpicKeyField
	r.StoryPointField = rule.StoryPointField
	r.RemotelinkCommitShaPattern = rule.RemotelinkCommitShaPattern
	r.TypeMappings = mappings
	return r, nil
}

func MakeTransformationRules(rule models.JiraTransformationRule) (*JiraTransformationRule, errors.Error) {
	var typeMapping TypeMappings
	err := json.Unmarshal(rule.TypeMappings, &typeMapping)
	if err != nil {
		return nil, errors.Default.Wrap(err, "unable to unmarshal the typeMapping")
	}
	result := &JiraTransformationRule{
		Name:                       rule.Name,
		EpicKeyField:               rule.EpicKeyField,
		StoryPointField:            rule.StoryPointField,
		RemotelinkCommitShaPattern: rule.RemotelinkCommitShaPattern,
		TypeMappings:               typeMapping,
	}
	return result, nil
}

type JiraOptions struct {
	ConnectionId         uint64 `json:"connectionId"`
	BoardId              uint64 `json:"boardId"`
	CreatedDateAfter     string
	TransformationRules  *JiraTransformationRule `json:"transformationRules"`
	ScopeId              string
	TransformationRuleId uint64
}

type JiraTaskData struct {
	Options          *JiraOptions
	ApiClient        *helper.ApiAsyncClient
	CreatedDateAfter *time.Time
	JiraServerInfo   models.JiraServerInfo
}

type JiraApiParams struct {
	ConnectionId uint64
	BoardId      uint64
}

func ValidateTaskOptions(op *JiraOptions) errors.Error {
	if op.ConnectionId == 0 {
		return errors.BadInput.New(fmt.Sprintf("invalid connectionId:%d", op.ConnectionId))
	}
	if op.BoardId == 0 {
		return errors.BadInput.New(fmt.Sprintf("invalid boardId:%d", op.BoardId))
	}
	if op.TransformationRules == nil && op.TransformationRuleId == 0 {
		op.TransformationRules = &JiraTransformationRule{}
	}
	return nil
}

func DecodeTaskOptions(options map[string]interface{}) (*JiraOptions, errors.Error) {
	var op JiraOptions
	err := helper.Decode(options, &op, nil)
	if err != nil {
		return nil, err
	}
	return &op, nil
}
