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
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

type createSingerTapState struct{}

type SingerTapState20221015 struct {
	Id           string `gorm:"primaryKey;type:varchar(255)"`
	ConnectionId uint64
	Type         string `gorm:"primaryKey;type:varchar(255)"`
	Value        datatypes.JSON
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (SingerTapState20221015) TableName() string {
	return "tap_state"
}

func (*createSingerTapState) Up(ctx context.Context, db *gorm.DB) errors.Error {
	err := db.Migrator().AutoMigrate(SingerTapState20221015{})
	if err != nil {
		return errors.Convert(err)
	}
	return nil
}

func (*createSingerTapState) Version() uint64 {
	return 20221015000001
}

func (*createSingerTapState) Name() string {
	return "Create tap state table"
}
