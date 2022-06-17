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

package api

import (
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/github/models"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type ApiUserPublicEmailResponse []models.PublicEmail

var vld *validator.Validate
var connectionHelper *helper.ConnectionApiHelper
var basicRes core.BasicRes

func Init(config *viper.Viper, logger core.Logger, database *gorm.DB) {
	basicRes = helper.NewDefaultBasicRes(config, logger, database)
	vld = validator.New()
	connectionHelper = helper.NewConnectionHelper(
		basicRes,
		vld,
	)
}
