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

package e2ehelper

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/helpers/pluginhelper"
	"github.com/apache/incubator-devlake/impl/dalgorm"
	"github.com/apache/incubator-devlake/logger"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/runner"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"os"
	"strings"
	"testing"
	"time"
)

// DataFlowTester provides a universal data integrity validation facility to help `Plugin` verifying records between
// each step.
//
// How it works:
//
//   1. Flush specified `table` and import data from a `csv` file by `ImportCsv` method
//   2. Execute specified `subtask` by `Subtask` method
//   3. Verify actual data from specified table against expected data from another `csv` file
//   4. Repeat step 2 and 3
//
// Recommended Usage:

//   1. Create a folder under your plugin root folder. i.e. `plugins/gitlab/e2e/ to host all your e2e-tests`
//   2. Create a folder named `tables` to hold all data in `csv` format
//   3. Create e2e test-cases to cover all possible data-flow routes
//
// Example code:
//
//   See [Gitlab Project Data Flow Test](plugins/gitlab/e2e/project_test.go) for detail
//
// DataFlowTester use `N`
type DataFlowTester struct {
	Cfg    *viper.Viper
	Db     *gorm.DB
	Dal    dal.Dal
	T      *testing.T
	Name   string
	Plugin core.PluginMeta
	Log    core.Logger
}

// NewDataFlowTester create a *DataFlowTester to help developer test their subtasks data flow
func NewDataFlowTester(t *testing.T, pluginName string, pluginMeta core.PluginMeta) *DataFlowTester {
	err := core.RegisterPlugin(pluginName, pluginMeta)
	if err != nil {
		panic(err)
	}
	cfg := config.GetConfig()
	e2eDbUrl := cfg.GetString(`E2E_DB_URL`)
	if e2eDbUrl == `` {
		panic(fmt.Errorf(`e2e can only run with E2E_DB_URL, please set it in .env`))
	}
	cfg.Set(`DB_URL`, cfg.GetString(`E2E_DB_URL`))
	db, err := runner.NewGormDb(cfg, logger.Global)
	if err != nil {
		panic(err)
	}
	return &DataFlowTester{
		Cfg:    cfg,
		Db:     db,
		Dal:    dalgorm.NewDalgorm(db),
		T:      t,
		Name:   pluginName,
		Plugin: pluginMeta,
		Log:    logger.Global,
	}
}

// ImportCsvIntoRawTable imports records from specified csv file into target raw table, note that existing data would be deleted first.
func (t *DataFlowTester) ImportCsvIntoRawTable(csvRelPath string, rawTableName string) {
	csvIter := pluginhelper.NewCsvFileIterator(csvRelPath)
	defer csvIter.Close()
	t.FlushRawTable(rawTableName)
	// load rows and insert into target table
	for csvIter.HasNext() {
		toInsertValues := csvIter.Fetch()
		toInsertValues[`data`] = json.RawMessage(toInsertValues[`data`].(string))
		result := t.Db.Table(rawTableName).Create(toInsertValues)
		if result.Error != nil {
			panic(result.Error)
		}
		assert.Equal(t.T, int64(1), result.RowsAffected)
	}
}

// ImportCsvIntoTabler imports records from specified csv file into target tabler, note that existing data would be deleted first.
func (t *DataFlowTester) ImportCsvIntoTabler(csvRelPath string, dst schema.Tabler) {
	csvIter := pluginhelper.NewCsvFileIterator(csvRelPath)
	defer csvIter.Close()
	t.FlushTabler(dst)
	// load rows and insert into target table
	for csvIter.HasNext() {
		toInsertValues := csvIter.Fetch()
		for i := range toInsertValues {
			if toInsertValues[i].(string) == `` {
				toInsertValues[i] = nil
			}
		}
		result := t.Db.Model(dst).Create(toInsertValues)
		if result.Error != nil {
			panic(result.Error)
		}
		assert.Equal(t.T, int64(1), result.RowsAffected)
	}
}

// MigrateRawTableAndFlush migrate table and deletes all records from specified table
func (t *DataFlowTester) FlushRawTable(rawTableName string) {
	// flush target table
	err := t.Db.Migrator().DropTable(rawTableName)
	if err != nil {
		panic(err)
	}
	err = t.Db.Table(rawTableName).AutoMigrate(&helper.RawData{})
	if err != nil {
		panic(err)
	}
}

// FlushTabler migrate table and deletes all records from specified table
func (t *DataFlowTester) FlushTabler(dst schema.Tabler) {
	// flush target table
	err := t.Db.Migrator().DropTable(dst)
	if err != nil {
		panic(err)
	}
	err = t.Db.AutoMigrate(dst)
	if err != nil {
		panic(err)
	}
}

// Subtask executes specified subtasks
func (t *DataFlowTester) Subtask(subtaskMeta core.SubTaskMeta, taskData interface{}) {
	subtaskCtx := helper.NewStandaloneSubTaskContext(t.Cfg, t.Log, t.Db, context.Background(), t.Name, taskData)
	err := subtaskMeta.EntryPoint(subtaskCtx)
	if err != nil {
		panic(err)
	}
}

// CreateSnapshot reads rows from database and write them into .csv file.
func (t *DataFlowTester) CreateSnapshot(dst schema.Tabler, csvRelPath string, pkfields []string, targetfields []string) {
	location, _ := time.LoadLocation(`UTC`)
	allFields := append(pkfields, targetfields...)
	dbCursor, err := t.Dal.Cursor(
		dal.Select(strings.Join(allFields, `,`)),
		dal.From(dst.TableName()),
		dal.Orderby(strings.Join(pkfields, `,`)),
	)
	if err != nil {
		panic(err)
	}

	columns, err := dbCursor.Columns()
	if err != nil {
		panic(err)
	}
	csvWriter := pluginhelper.NewCsvFileWriter(csvRelPath, columns)
	defer csvWriter.Close()

	// define how to scan value
	columnTypes, _ := dbCursor.ColumnTypes()
	forScanValues := make([]interface{}, len(allFields))
	for i, columnType := range columnTypes {
		if columnType.ScanType().Name() == `Time` || columnType.ScanType().Name() == `NullTime` {
			forScanValues[i] = new(sql.NullTime)
		} else if columnType.ScanType().Name() == `bool` {
			forScanValues[i] = new(bool)
		} else {
			forScanValues[i] = new(string)
		}
	}

	for dbCursor.Next() {
		err = dbCursor.Scan(forScanValues...)
		if err != nil {
			panic(err)
		}
		values := make([]string, len(allFields))
		for i := range forScanValues {
			switch forScanValues[i].(type) {
			case *sql.NullTime:
				value := forScanValues[i].(*sql.NullTime)
				if value.Valid {
					values[i] = value.Time.In(location).Format("2006-01-02T15:04:05.000-07:00")
				} else {
					values[i] = ``
				}
			case *bool:
				if *forScanValues[i].(*bool) {
					values[i] = `1`
				} else {
					values[i] = `0`
				}
			case *string:
				values[i] = fmt.Sprint(*forScanValues[i].(*string))
			}
		}
		csvWriter.Write(values)
	}
}

// VerifyTable reads rows from csv file and compare with records from database one by one. You must specified the
// Primary Key Fields with `pkfields` so DataFlowTester could select the exact record from database, as well as which
// fields to compare with by specifying `targetfields` parameter.
func (t *DataFlowTester) VerifyTable(dst schema.Tabler, csvRelPath string, pkfields []string, targetfields []string) {
	_, err := os.Stat(csvRelPath)
	if os.IsNotExist(err) {
		t.CreateSnapshot(dst, csvRelPath, pkfields, targetfields)
		return
	}

	csvIter := pluginhelper.NewCsvFileIterator(csvRelPath)
	location, _ := time.LoadLocation(`UTC`)
	defer csvIter.Close()

	expectedTotal := int64(0)
	for csvIter.HasNext() {
		expected := csvIter.Fetch()
		pkvalues := make([]interface{}, 0, len(pkfields))
		for _, pkf := range pkfields {
			pkvalues = append(pkvalues, expected[pkf])
		}
		actual := make(map[string]interface{})
		where := []string{}
		for _, field := range pkfields {
			where = append(where, fmt.Sprintf(" %s = ?", field))
		}

		var actualCount int64
		err := t.Db.Table(dst.TableName()).Where(strings.Join(where, ` AND `), pkvalues...).Count(&actualCount).Error
		if err != nil {
			panic(err)
		}
		assert.Equal(t.T, int64(1), actualCount, fmt.Sprintf(`%s found not eq 1 but %d (with params from csv %s)`, dst.TableName(), actualCount, pkvalues))
		if actualCount != 1 {
			continue
		}

		err = t.Db.Table(dst.TableName()).Where(strings.Join(where, ` AND `), pkvalues...).Find(actual).Error
		if err != nil {
			panic(err)
		}
		for _, field := range targetfields {
			actualValue := ""
			switch actual[field].(type) {
			case time.Time:
				if actual[field] != nil {
					actualValue = actual[field].(time.Time).In(location).Format("2006-01-02T15:04:05.000-07:00")
				}
			case bool:
				if actual[field].(bool) {
					actualValue = `1`
				} else {
					actualValue = `0`
				}
			default:
				if actual[field] != nil {
					actualValue = fmt.Sprint(actual[field])
				}
			}
			assert.Equal(t.T, expected[field], actualValue, fmt.Sprintf(`%s.%s not match (with params from csv %s)`, dst.TableName(), field, pkvalues))
		}
		expectedTotal++
	}

	var actualTotal int64
	err = t.Db.Table(dst.TableName()).Count(&actualTotal).Error
	if err != nil {
		panic(err)
	}
	assert.Equal(t.T, expectedTotal, actualTotal, fmt.Sprintf(`%s count not match`, dst.TableName()))
}
