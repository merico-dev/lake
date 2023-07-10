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
	Mode        string
}

type BlueprintProjectPairs struct {
	Projects   []string `json:"projects"`
	Blueprints []string `json:"blueprints"`
}

func NewBlueprintManager(db dal.Dal) *BlueprintManager {
	return &BlueprintManager{
		db: db,
	}
}

func NewBlueprintProjectPairs(bps []*models.Blueprint) *BlueprintProjectPairs {
	pairs := &BlueprintProjectPairs{}
	for _, bp := range bps {
		pairs.Blueprints = append(pairs.Blueprints, bp.Name)
		pairs.Projects = append(pairs.Projects, bp.ProjectName)
	}
	return pairs
}

// SaveDbBlueprint accepts a Blueprint instance and upsert it to database
func (b *BlueprintManager) SaveDbBlueprint(blueprint *models.Blueprint) errors.Error {
	var err errors.Error
	if err = b.transact(func(tx dal.Transaction) errors.Error {
		if blueprint.ID != 0 {
			err = b.db.Update(&blueprint)
		} else {
			err = b.db.Create(&blueprint)
		}
		if err != nil {
			return err
		}
		if err = saveDbBlueprintSettings(tx, blueprint); err != nil {
			return err
		}
		return err
	}); err != nil {
		return err
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
	if query.Mode != "" {
		clauses = append(clauses, dal.Where("mode = ?", query.Mode))
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

// GetBlueprintSettings perhaps optimize this to support grabbing settings for multiple IDs without needing for-loops?
func (b *BlueprintManager) GetBlueprintSettings(blueprintId uint64) (*models.BlueprintSettings, errors.Error) {
	settings := models.BlueprintSettings{
		BlueprintId: blueprintId,
	}
	err := b.db.First(&settings)
	if err != nil {
		return nil, err
	}
	var connections []*models.BlueprintConnection
	err = b.db.All(&connections, dal.Where("blueprint_id", blueprintId))
	if err != nil {
		return nil, err
	}
	settings.Connections = connections
	var scopes []*models.BlueprintScope
	err = b.db.All(&scopes, dal.Where("blueprint_id", blueprintId))
	if err != nil {
		return nil, err
	}
	mappedScopes := mapScopesByBpConnectionId(scopes)
	for _, connection := range settings.Connections {
		if connectionScopes, ok := mappedScopes[connection.ID]; ok {
			connection.Scopes = connectionScopes
		}
	}
	return &settings, nil
}

func mapScopesByBpConnectionId(scopes []*models.BlueprintScope) map[uint64][]*models.BlueprintScope {
	mapped := map[uint64][]*models.BlueprintScope{}
	for _, scope := range scopes {
		if list, ok := mapped[scope.ConnectionId]; !ok {
			mapped[scope.ConnectionId] = []*models.BlueprintScope{scope}
		} else {
			mapped[scope.ConnectionId] = append(list, scope)
		}
	}
	return mapped
}

// GetBlueprintsByScopes returns all blueprints that have these scopeIds and this connection Id
func (b *BlueprintManager) GetBlueprintsByScopes(connectionId uint64, pluginName string, scopeIds ...string) (map[string][]*models.Blueprint, errors.Error) {
	bps, _, err := b.GetDbBlueprints(&GetBlueprintQuery{Mode: "NORMAL"})
	if err != nil {
		return nil, err
	}
	scopeMap := map[string][]*models.Blueprint{}
	for _, bp := range bps {
		scopes, err := bp.GetScopes(connectionId, pluginName)
		if err != nil {
			return nil, err
		}
		for _, scope := range scopes {
			if contains(scopeIds, scope.ScopeId) {
				scopeMap[scope.ScopeId] = append(scopeMap[scope.ScopeId], bp)
			}
		}
	}
	return scopeMap, nil
}

// GetBlueprintsByScopes returns all blueprints that have these scopeIds and this connection Id
func (b *BlueprintManager) GetBlueprintsByConnection(plugin string, connectionId uint64) ([]*models.Blueprint, errors.Error) {
	bps, _, err := b.GetDbBlueprints(&GetBlueprintQuery{})
	if err != nil {
		return nil, err
	}
	var filteredBps []*models.Blueprint
	for _, bp := range bps {
		connections := bp.Settings.Connections
		for _, connection := range connections {
			if connection.ConnectionId == connectionId && connection.Plugin == plugin {
				filteredBps = append(filteredBps, bp)
				break
			}
		}
	}
	return filteredBps, nil
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
	return b.transact(func(tx dal.Transaction) errors.Error {
		err := tx.Delete(&models.DbBlueprintLabel{}, dal.Where("blueprint_id = ?", id))
		if err != nil {
			return err
		}
		err = tx.Delete(&models.Blueprint{
			Model: common.Model{
				ID: id,
			},
		})
		if err != nil {
			return err
		}
		err = tx.Delete(&models.BlueprintSettings{
			BlueprintId: id,
		})
		if err != nil {
			return err
		}
		err = tx.Delete(&models.BlueprintConnection{
			BlueprintId: id,
		})
		if err != nil {
			return err
		}
		err = tx.Delete(&models.BlueprintScope{
			BlueprintId: id,
		})
		if err != nil {
			return err
		}
		return err
	})
}

func (b *BlueprintManager) fillBlueprintDetail(blueprint *models.Blueprint) errors.Error {
	err := b.db.Pluck("name", &blueprint.Labels, dal.From(&models.DbBlueprintLabel{}), dal.Where("blueprint_id = ?", blueprint.ID))
	if err != nil {
		return errors.Default.Wrap(err, "error getting the blueprint labels from database")
	}
	settings, err := b.GetBlueprintSettings(blueprint.ID)
	if err != nil {
		return err
	}
	blueprint.Settings = settings
	return nil
}

func saveDbBlueprintSettings(t dal.Transaction, bp *models.Blueprint) errors.Error {
	bp.Settings.BlueprintId = bp.ID
	err := t.CreateOrUpdate(bp.Settings)
	if err != nil {
		return err
	}
	return saveDbBlueprintConnections(t, bp.Settings)
}

func saveDbBlueprintConnections(t dal.Transaction, settings *models.BlueprintSettings) errors.Error {
	for _, connection := range settings.Connections {
		connection.BlueprintId = settings.BlueprintId
		connection.SettingsId = settings.ID
	}
	err := t.CreateOrUpdate(&settings.Connections)
	if err != nil {
		return err
	}
	for _, connection := range settings.Connections {
		err = saveDbBlueprintScopes(t, connection)
		if err != nil {
			return err
		}
	}
	return nil
}

func saveDbBlueprintScopes(t dal.Transaction, connection *models.BlueprintConnection) errors.Error {
	for _, scope := range connection.Scopes {
		scope.BlueprintId = connection.BlueprintId
		scope.ConnectionId = connection.ID
	}
	err := t.CreateOrUpdate(&connection.Scopes)
	if err != nil {
		return err
	}
	return nil
}

func (b *BlueprintManager) transact(f func(tx dal.Transaction) errors.Error) errors.Error {
	tx := b.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			err := tx.Rollback()
			if err != nil {
				panic(fmt.Errorf("rollback failure: %v\noriginal error: %v", err, r.(error).Error()))
			}
		}
	}()
	err := f(tx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func contains(list []string, target string) bool {
	for _, t := range list {
		if t == target {
			return true
		}
	}
	return false
}
