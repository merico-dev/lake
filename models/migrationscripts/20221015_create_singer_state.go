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
	"context"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/models/migrationscripts/archived"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type createSingerTapState struct{}

type SingerTapState20220928 struct {
	archived.NoPKModel
	Id           string `gorm:"primaryKey;type:varchar(255)"`
	ConnectionId uint64
	Type         string `gorm:"primaryKey;type:varchar(255)"`
	Value        datatypes.JSON
}

func (SingerTapState20220928) TableName() string {
	return "singer_tap_state"
}

func (*createSingerTapState) Up(ctx context.Context, db *gorm.DB) errors.Error {
	err := db.Migrator().AutoMigrate(SingerTapState20220928{})
	if err != nil {
		return errors.Convert(err)
	}
	return nil
}

func (*createSingerTapState) Version() uint64 {
	return 20221015000001
}

func (*createSingerTapState) Name() string {
	return "Create singer tap state table"
}
