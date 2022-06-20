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

	"gorm.io/gorm"
)

type Blueprint20220616 struct {
	Mode     string `json:"mode" gorm:"varchar(20)" validate:"required,oneof=NORMAL ADVANCED"`
	IsManual bool   `json:"isManual"`
}

func (Blueprint20220616) TableName() string {
	return "_devlake_blueprints"
}

type updateSchemas20220616 struct{}

func (*updateSchemas20220616) Up(ctx context.Context, db *gorm.DB) error {
	err := db.Migrator().AutoMigrate(&Blueprint20220616{})
	if err != nil {
		return err
	}
	return nil
}

func (*updateSchemas20220616) Version() uint64 {
	return 20220616110537
}

func (*updateSchemas20220616) Name() string {
	return "add mode field to blueprint"
}
