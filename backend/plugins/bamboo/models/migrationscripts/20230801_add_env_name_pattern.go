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

package migrationscripts

import (
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
)

var _ plugin.MigrationScript = (*addEnvNamePattern)(nil)

type scopeConfig20230801 struct {
	EnvNamePattern string `mapstructure:"envNamePattern,omitempty" json:"envNamePattern" gorm:"type:varchar(255)"`
}

func (scopeConfig20230801) TableName() string {
	return "_tool_bamboo_scope_configs"
}

type addEnvNamePattern struct{}

func (script *addEnvNamePattern) Up(basicRes context.BasicRes) errors.Error {
	return basicRes.GetDal().AutoMigrate(&scopeConfig20230801{})
}

func (*addEnvNamePattern) Version() uint64 {
	return 20230801213814
}

func (script *addEnvNamePattern) Name() string {
	return "add env_name_pattern to _tool_bamboo_scope_configs"
}
