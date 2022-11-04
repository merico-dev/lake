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
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/helpers/migrationhelper"
	"github.com/apache/incubator-devlake/plugins/ae/models/migrationscripts/archived"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
)

type addInitTables20220714 struct{}

func (u *addInitTables20220714) Up(basicRes core.BasicRes) errors.Error {
	err := basicRes.GetDal().DropTables(
		&archived.AECommit{},
		&archived.AEProject{},
		"_raw_ae_project",
		"_raw_ae_commits",
	)
	if err != nil {
		return err
	}

	err = migrationhelper.AutoMigrateTables(
		basicRes,
		&archived.AECommit{},
		&archived.AEProject{},
		&archived.AeConnection{},
	)
	if err != nil {
		return err
	}

	c := config.GetConfig()
	encodeKey := c.GetString(core.EncodeKeyEnvStr)
	connection := &archived.AeConnection{}
	connection.Endpoint = c.GetString("AE_ENDPOINT")
	connection.Proxy = c.GetString("AE_PROXY")
	connection.SecretKey = c.GetString("AE_SECRET_KEY")
	connection.AppId = c.GetString("AE_APP_ID")
	connection.Name = "AE"
	if connection.Endpoint != "" && connection.AppId != "" && connection.SecretKey != "" && encodeKey != "" {
		err = helper.UpdateEncryptFields(connection, func(plaintext string) (string, errors.Error) {
			return core.Encrypt(encodeKey, plaintext)
		})
		if err != nil {
			return err
		}
		// update from .env and save to db
		err = basicRes.GetDal().Create(connection)
		if err != nil {
			return err
		}
	}

	return nil
}

func (*addInitTables20220714) Version() uint64 {
	return 20220714201133
}

func (*addInitTables20220714) Name() string {
	return "AE init schemas"
}
