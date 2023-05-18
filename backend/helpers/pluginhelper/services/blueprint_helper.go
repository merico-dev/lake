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

package services

import (
	"fmt"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/models/common"
	"github.com/apache/incubator-devlake/core/plugin"
)

type BlueprintManager struct {
	db dal.Dal
}

type GetBlueprintQuery struct {
	Enable      *bool
	IsManual    *bool
	Label       string
	SkipRecords int
	PageSize    int
}

func NewBlueprintManager(db dal.Dal) *BlueprintManager {
	return &BlueprintManager{
		db: db,
	}
}

// SaveDbBlueprint accepts a Blueprint instance and upsert it to database
func (b *BlueprintManager) SaveDbBlueprint(blueprint *models.Blueprint) errors.Error {
	var err error
	if blueprint.ID != 0 {
		err = b.db.Update(&blueprint)
	} else {
		err = b.db.Create(&blueprint)
	}
	if err != nil {
		return errors.Default.Wrap(err, "error creating DB blueprint")
	}
	err = b.db.Delete(&models.DbBlueprintLabel{}, dal.Where(`blueprint_id = ?`, blueprint.ID))
	if err != nil {
		return errors.Default.Wrap(err, "error delete DB blueprint's old labelModels")
	}
	if len(blueprint.Labels) > 0 {
		blueprintLabels := make([]*models.DbBlueprintLabel, 0)
		for i := range blueprint.Labels {
			blueprintLabels = append(blueprintLabels, &models.DbBlueprintLabel{
				BlueprintId: blueprint.ID,
				Name:        blueprint.Labels[i],
			})
		}
		err = b.db.Create(&blueprintLabels)
		if err != nil {
			return errors.Default.Wrap(err, "error creating DB blueprint's labelModels")
		}
	}
	return nil
}

// GetDbBlueprints returns a paginated list of Blueprints based on `query`
func (b *BlueprintManager) GetDbBlueprints(query *GetBlueprintQuery) ([]*models.Blueprint, int64, errors.Error) {
	// process query parameters
	clauses := []dal.Clause{dal.From(&models.Blueprint{})}
	if query.Enable != nil {
		clauses = append(clauses, dal.Where("enable = ?", *query.Enable))
	}
	if query.IsManual != nil {
		clauses = append(clauses, dal.Where("is_manual = ?", *query.IsManual))
	}
	if query.Label != "" {
		clauses = append(clauses,
			dal.Join("left join _devlake_blueprint_labels bl ON bl.blueprint_id = _devlake_blueprints.id"),
			dal.Where("bl.name = ?", query.Label),
		)
	}

	// count total records
	count, err := b.db.Count(clauses...)
	if err != nil {
		return nil, 0, err
	}
	clauses = append(clauses, dal.Orderby("id DESC"))
	// load paginated blueprints from database
	if query.SkipRecords != 0 {
		clauses = append(clauses, dal.Offset(query.SkipRecords))
	}
	if query.PageSize != 0 {
		clauses = append(clauses, dal.Limit(query.PageSize))
	}
	dbBlueprints := make([]*models.Blueprint, 0)
	err = b.db.All(&dbBlueprints, clauses...)
	if err != nil {
		return nil, 0, errors.Default.Wrap(err, "error getting DB count of blueprints")
	}

	// load labels for blueprints
	for _, dbBlueprint := range dbBlueprints {
		err = b.fillBlueprintDetail(dbBlueprint)
		if err != nil {
			return nil, 0, err
		}
	}

	return dbBlueprints, count, nil
}

// GetDbBlueprint returns the detail of a given Blueprint ID
func (b *BlueprintManager) GetDbBlueprint(blueprintId uint64) (*models.Blueprint, errors.Error) {
	blueprint := &models.Blueprint{}
	err := b.db.First(blueprint, dal.Where("id = ?", blueprintId))
	if err != nil {
		if b.db.IsErrorNotFound(err) {
			return nil, errors.NotFound.Wrap(err, "could not find blueprint in DB")
		}
		return nil, errors.Default.Wrap(err, "error getting blueprint from DB")
	}
	err = b.fillBlueprintDetail(blueprint)
	if err != nil {
		return nil, err
	}
	return blueprint, nil
}

// GetBlueprintsByScopes returns all blueprints that have these scopeIds and this connection Id
func (b *BlueprintManager) GetBlueprintsByScopes(connectionId uint64, scopeIds ...string) (map[string][]*models.Blueprint, errors.Error) {
	bps, _, err := b.GetDbBlueprints(&GetBlueprintQuery{})
	if err != nil {
		return nil, err
	}
	scopeMap := map[string][]*models.Blueprint{}
	for _, bp := range bps {
		scopes, err := bp.GetScopes(connectionId)
		if err != nil {
			return nil, err
		}
		for _, scope := range scopes {
			if contains(scopeIds, scope.Id) {
				if inserted, ok := scopeMap[scope.Id]; !ok {
					scopeMap[scope.Id] = []*models.Blueprint{bp}
				} else {
					inserted = append(inserted, bp)
					scopeMap[scope.Id] = inserted
				}
				break
			}
		}
	}
	return scopeMap, nil
}

// GetBlueprintConnections returns the connections associated with this blueprint Id
func (b *BlueprintManager) GetBlueprintConnections(blueprintId uint64) ([]*plugin.BlueprintConnectionV200, errors.Error) {
	bp, err := b.GetDbBlueprint(blueprintId)
	if err != nil {
		return nil, err
	}
	return bp.GetConnections()
}

// GetDbBlueprintByProjectName returns the detail of a given projectName
func (b *BlueprintManager) GetDbBlueprintByProjectName(projectName string) (*models.Blueprint, errors.Error) {
	dbBlueprint := &models.Blueprint{}
	err := b.db.First(dbBlueprint, dal.Where("project_name = ?", projectName))
	if err != nil {
		if b.db.IsErrorNotFound(err) {
			return nil, errors.NotFound.Wrap(err, fmt.Sprintf("could not find blueprint in DB by projectName %s", projectName))
		}
		return nil, errors.Default.Wrap(err, fmt.Sprintf("error getting blueprint from DB by projectName %s", projectName))
	}
	err = b.fillBlueprintDetail(dbBlueprint)
	if err != nil {
		return nil, err
	}
	return dbBlueprint, nil
}

// DeleteBlueprint deletes a blueprint by its id
func (b *BlueprintManager) DeleteBlueprint(id uint64) errors.Error {
	dbBlueprint := &models.Blueprint{
		Model: common.Model{
			ID: id,
		},
	}
	return b.db.Delete(&dbBlueprint)
}

func (b *BlueprintManager) fillBlueprintDetail(blueprint *models.Blueprint) errors.Error {
	err := b.db.Pluck("name", &blueprint.Labels, dal.From(&models.DbBlueprintLabel{}), dal.Where("blueprint_id = ?", blueprint.ID))
	if err != nil {
		return errors.Internal.Wrap(err, "error getting the blueprint labels from database")
	}
	return nil
}

func contains(list []string, target string) bool {
	for _, t := range list {
		if t == target {
			return true
		}
	}
	return false
}
