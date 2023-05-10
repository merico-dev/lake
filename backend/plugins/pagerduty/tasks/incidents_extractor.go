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
	"github.com/apache/incubator-devlake/plugins/pagerduty/models"
	"github.com/apache/incubator-devlake/plugins/pagerduty/models/raw"
)

var _ plugin.SubTaskEntryPoint = ExtractIncidents

func ExtractIncidents(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*PagerDutyTaskData)
	extractor, err := api.NewApiExtractor(api.ApiExtractorArgs{
		RawDataSubTaskArgs: api.RawDataSubTaskArgs{
			Ctx:     taskCtx,
			Options: data.Options,
			Table:   RAW_INCIDENTS_TABLE,
		},
		Extract: func(row *api.RawData) ([]interface{}, errors.Error) {
			incidentRaw := &raw.Incidents{}
			err := errors.Convert(json.Unmarshal(row.Data, incidentRaw))
			if err != nil {
				return nil, err
			}
			results := make([]interface{}, 0, 1)
			incident := models.Incident{
				ConnectionId: data.Options.ConnectionId,
				Number:       *incidentRaw.IncidentNumber,
				Url:          *incidentRaw.HtmlUrl,
				Summary:      *incidentRaw.Summary,
				Status:       models.IncidentStatus(*incidentRaw.Status),
				Urgency:      models.IncidentUrgency(*incidentRaw.Urgency),
				CreatedDate:  *incidentRaw.CreatedAt,
				UpdatedDate:  *incidentRaw.LastStatusChangeAt,
			}
			results = append(results, &incident)
			if incidentRaw.Service != nil {
				service := models.Service{
					ConnectionId: data.Options.ConnectionId,
					Url:          resolve(incidentRaw.Service.HtmlUrl),
					Id:           *incidentRaw.Service.Id,
					Name:         *incidentRaw.Service.Summary,
				}
				incident.ServiceId = service.Id
				results = append(results, &service)
			}
			for _, assignmentRaw := range incidentRaw.Assignments {
				userRaw := assignmentRaw.Assignee
				results = append(results, &models.Assignment{
					ConnectionId:   data.Options.ConnectionId,
					UserId:         *userRaw.Id,
					IncidentNumber: *incidentRaw.IncidentNumber,
					AssignedAt:     *assignmentRaw.At,
				})
				results = append(results, &models.User{
					ConnectionId: data.Options.ConnectionId,
					Id:           *userRaw.Id,
					Url:          resolve(userRaw.HtmlUrl),
					Name:         *userRaw.Summary,
				})
			}
			return results, nil
		},
	})
	if err != nil {
		return err
	}
	return extractor.Execute()
}

func resolve[T any](t *T) T {
	if t == nil {
		return *new(T)
	}
	return *t
}

var ExtractIncidentsMeta = plugin.SubTaskMeta{
	Name:             "extractIncidents",
	EntryPoint:       ExtractIncidents,
	EnabledByDefault: true,
	Description:      "Extract PagerDuty incidents",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
}
