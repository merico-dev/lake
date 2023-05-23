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

	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/common"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
)

type Operation interface {
	Execute(dal.Dal) errors.Error
}

type BaseOperation struct {
}

type ExecuteOperation struct {
	Sql string `json:"sql"`
}

func (o ExecuteOperation) Execute(dal dal.Dal) errors.Error {
	return dal.Exec(o.Sql)
}

var _ Operation = (*ExecuteOperation)(nil)

type AutoMigrateOperation struct {
	DynamicModelInfo DynamicModelInfo `json:"dynamic_model_info"`
}

func (o AutoMigrateOperation) Execute(dal dal.Dal) errors.Error {
	tabler, err := o.DynamicModelInfo.LoadDynamicTabler(common.NoPKModel{})
	if err != nil {
		return err
	}
	return api.CallDB(dal.AutoMigrate, tabler.New())
}

var _ Operation = (*AutoMigrateOperation)(nil)

type RemoteMigrationScript struct {
	operations []Operation
	version    uint64
	name       string
}

func (s *RemoteMigrationScript) UnmarshalJSON(data []byte) error {
	var rawScript map[string]interface{}
	err := json.Unmarshal(data, &rawScript)
	if err != nil {
		return err
	}
	s.version = uint64(rawScript["version"].(float64))
	s.name = rawScript["name"].(string)
	operationsRaw := rawScript["operations"].([]interface{})
	s.operations = make([]Operation, len(operationsRaw))
	for _, operationRaw := range operationsRaw {
		operationMap := operationRaw.(map[string]interface{})
		operationType := operationMap["type"].(string)
		var operation Operation
		switch operationType {
		case "execute":
			operation = &ExecuteOperation{}
		case "auto_migrate":
			operation = &AutoMigrateOperation{}
		default:
			return errors.BadInput.New("unsupported operation type")
		}
		json.Unmarshal(operationRaw.([]byte), operation)
		s.operations = append(s.operations, operation)
	}
	return nil
}

func (s *RemoteMigrationScript) Up(basicRes context.BasicRes) errors.Error {
	dal := basicRes.GetDal()
	for _, operation := range s.operations {
		err := operation.Execute(dal)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *RemoteMigrationScript) Version() uint64 {
	return s.version
}

func (s *RemoteMigrationScript) Name() string {
	return s.name
}

var _ plugin.MigrationScript = (*RemoteMigrationScript)(nil)
