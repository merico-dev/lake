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

package srvhelper

import (
	"fmt"
	"time"

	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/log"
	"github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/helpers/dbhelper"
	"github.com/go-playground/validator/v10"
)

type CustomValidator interface {
	CustomValidate(entitiy interface{}, validate *validator.Validate) errors.Error
}

type ModelSrvHelper[M dal.Tabler] struct {
	basicRes      context.BasicRes
	log           log.Logger
	db            dal.Dal
	validator     *validator.Validate
	modelName     string
	pk            []dal.ColumnMeta
	pkWhere       string
	pkCount       int
	searchColumns []string
}

func NewModelSrvHelper[M dal.Tabler](basicRes context.BasicRes) *ModelSrvHelper[M] {
	m := new(M)
	modelName := fmt.Sprintf("%T", m)
	db := basicRes.GetDal()
	if db == nil {
		db = basicRes.GetDal()
	}
	pk := errors.Must1(dal.GetPrimarykeyColumns(db, *m))
	pkWhere := ""
	for _, col := range pk {
		if pkWhere != "" {
			pkWhere += " AND "
		}
		pkWhere += fmt.Sprintf("%s = ? ", col.Name())
	}
	return &ModelSrvHelper[M]{
		basicRes:  basicRes,
		log:       basicRes.GetLogger().Nested(fmt.Sprintf("%s_dal", modelName)),
		db:        db,
		validator: validator.New(),
		modelName: modelName,
		pk:        pk,
		pkWhere:   pkWhere,
		pkCount:   len(pk),
	}
}

func (self *ModelSrvHelper[M]) NewTx(tx dal.Transaction) *ModelSrvHelper[M] {
	helper := new(ModelSrvHelper[M])
	*helper = *self
	helper.db = tx
	return helper
}

func (self *ModelSrvHelper[M]) Validate(model *M) errors.Error {
	// the model can validate itself
	if customValidator, ok := (interface{}(model)).(CustomValidator); ok {
		return customValidator.CustomValidate(model, self.validator)
	}
	// basic validator
	if e := self.validator.Struct(model); e != nil {
		return errors.BadInput.Wrap(e, "validation faild")
	}
	return nil
}

// Create validates given model and insert it into database if validation passed
func (self *ModelSrvHelper[M]) Create(model *M) errors.Error {
	println("create model")
	err := self.Validate(model)
	if err != nil {
		return err
	}
	err = self.db.Create(model)
	if err != nil && self.db.IsDuplicationError(err) {
		return errors.Conflict.Wrap(err, fmt.Sprintf("%s already exists", self.modelName))
	}
	return err
}

// Update validates given model and update it into database if validation passed
func (self *ModelSrvHelper[M]) Update(model *M) errors.Error {
	err := self.Validate(model)
	if err != nil {
		return err
	}
	if err != nil && self.db.IsDuplicationError(err) {
		return errors.Conflict.Wrap(err, fmt.Sprintf("%s already exists", self.modelName))
	}
	return self.db.Update(model)
}

// CreateOrUpdate validates given model and insert or update it into database if validation passed
func (self *ModelSrvHelper[M]) CreateOrUpdate(model *M) errors.Error {
	err := self.Validate(model)
	if err != nil {
		return err
	}
	return self.db.CreateOrUpdate(model)
}

// Delete deletes given model from database
func (self *ModelSrvHelper[M]) Delete(model *M) errors.Error {
	return self.db.Delete(model)
}

// FindByPk returns model with given primary key from database
func (self *ModelSrvHelper[M]) FindByPk(pk ...interface{}) (*M, errors.Error) {
	if len(pk) != self.pkCount {
		return nil, errors.BadInput.New("invalid primary key")
	}
	model := new(M)
	err := self.db.First(model, dal.Where(self.pkWhere, pk...))
	if err != nil {
		if self.db.IsErrorNotFound(err) {
			return nil, errors.NotFound.Wrap(err, fmt.Sprintf("%s not found", self.modelName))
		}
		return nil, err
	}
	return model, nil
}

// GetAll returns all models from database
func (self *ModelSrvHelper[M]) GetAll() ([]*M, errors.Error) {
	models := make([]*M, 0)
	return models, self.db.All(&models)
}

func (self *ModelSrvHelper[M]) GetPage(pagination *Pagination, query ...dal.Clause) ([]*M, int64, errors.Error) {
	query = append(query, dal.From(new(M)))
	// process keyword
	searchTerm := pagination.SearchTerm
	if searchTerm != "" && len(self.searchColumns) > 0 {
		sql := ""
		value := "%" + searchTerm + "%"
		values := make([]interface{}, len(self.searchColumns))
		for i, field := range self.searchColumns {
			if sql != "" {
				sql += " OR "
			}
			sql += fmt.Sprintf("%s LIKE ?", field)
			values[i] = value
		}
		sql = fmt.Sprintf("(%s)", sql)
		query = append(query,
			dal.Where(sql, values...),
		)
	}
	count, err := self.db.Count(query...)
	if err != nil {
		return nil, 0, err
	}
	query = append(query, dal.Limit(pagination.GetLimit()), dal.Offset(pagination.GetOffset()))
	var scopes []*M
	return scopes, count, self.db.All(&scopes, query...)
}

func (self *ModelSrvHelper[M]) NoRunningPipeline(fn func(tx dal.Transaction) errors.Error, tablesToLock ...*dal.LockTable) (err errors.Error) {
	// make sure no pipeline is running
	tablesToLock = append(tablesToLock, &dal.LockTable{Table: "_devlake_pipelines", Exclusive: true})
	txHelper := dbhelper.NewTxHelper(self.basicRes, &err)
	defer txHelper.End()
	tx := txHelper.Begin()
	err = txHelper.LockTablesTimeout(2*time.Second, tablesToLock)
	if err != nil {
		err = errors.Conflict.Wrap(err, "lock pipelines table timedout")
		return
	}
	count := errors.Must1(tx.Count(
		dal.From("_devlake_pipelines"),
		dal.Where("status = ?", models.TASK_RUNNING),
	))
	if count > 0 {
		err = errors.Conflict.New("at least one pipeline is running")
		return
	}
	// time.Sleep(1 * time.Minute) # uncomment this line if you were to verify pipelines get blocked while deleting data
	// creating a nested transaction to avoid mysql complaining about table(s) NOT being locked
	nextedTxHelper := dbhelper.NewTxHelper(self.basicRes, &err)
	defer nextedTxHelper.End()
	nestedTX := nextedTxHelper.Begin()
	err = fn(nestedTX)
	return
}
