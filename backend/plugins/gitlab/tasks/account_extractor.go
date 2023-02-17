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

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/gitlab/models"
)

var ExtractAccountsMeta = plugin.SubTaskMeta{
	Name:             "extractAccounts",
	EntryPoint:       ExtractAccounts,
	EnabledByDefault: true,
	Description:      "Extract raw workspace data into tool layer table _tool_gitlab_accounts",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_CROSS},
}

func ExtractAccounts(taskCtx plugin.SubTaskContext) errors.Error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_USER_TABLE)

	extractor, err := api.NewApiExtractor(api.ApiExtractorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Extract: func(row *api.RawData) ([]interface{}, errors.Error) {
			var userRes models.GitlabAccount
			err := errors.Convert(json.Unmarshal(row.Data, &userRes))
			if err != nil {
				return nil, err
			}
			userRes.ConnectionId = data.Options.ConnectionId
			return []interface{}{&userRes}, nil
		},
	})

	if err != nil {
		return err
	}

	err = extractor.Execute()
	if err != nil {
		return err
	}

	return nil
}
