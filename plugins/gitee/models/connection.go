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

import "github.com/apache/incubator-devlake/plugins/helper"

type GiteeConnection struct {
	helper.RestConnection `mapstructure:",squash"`
	helper.AccessToken    `mapstructure:",squash"`
}

type GiteeResponse struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
	GiteeConnection
}

type ApiUserResponse struct {
	Id   int
	Name string `json:"name"`
}

type TestConnectionRequest struct {
	Endpoint           string `json:"endpoint" validate:"required"`
	Proxy              string `json:"proxy"`
	helper.AccessToken `mapstructure:",squash"`
}

type TransformationRules struct {
	PrType               string `mapstructure:"prType" env:"GITEE_PR_TYPE" json:"prType"`
	PrComponent          string `mapstructure:"prComponent" env:"GITEE_PR_COMPONENT" json:"prComponent"`
	PrBodyClosePattern   string `mapstructure:"prBodyClosePattern" json:"prBodyClosePattern"`
	IssueSeverity        string `mapstructure:"issueSeverity" env:"GITEE_ISSUE_SEVERITY" json:"issueSeverity"`
	IssuePriority        string `mapstructure:"issuePriority" env:"GITEE_ISSUE_PRIORITY" json:"issuePriority"`
	IssueComponent       string `mapstructure:"issueComponent" env:"GITEE_ISSUE_COMPONENT" json:"issueComponent"`
	IssueTypeBug         string `mapstructure:"issueTypeBug" env:"GITEE_ISSUE_TYPE_BUG" json:"issueTypeBug"`
	IssueTypeIncident    string `mapstructure:"issueTypeIncident" env:"GITEE_ISSUE_TYPE_INCIDENT" json:"issueTypeIncident"`
	IssueTypeRequirement string `mapstructure:"issueTypeRequirement" env:"GITEE_ISSUE_TYPE_REQUIREMENT" json:"issueTypeRequirement"`
	DeploymentPattern    string `mapstructure:"deploymentPattern" json:"deploymentPattern"`
}

func (GiteeConnection) TableName() string {
	return "_tool_gitee_connections"
}
