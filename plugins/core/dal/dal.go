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

package dal

import (
	"database/sql"
	"reflect"

	"github.com/apache/incubator-devlake/errors"
)

type Tabler interface {
	TableName() string
}

// Clause represents SQL Clause
type Clause struct {
	Type string
	Data interface{}
}

// ColumnMeta column type interface
type ColumnMeta interface {
	Name() string
	DatabaseTypeName() string                 // varchar
	ColumnType() (columnType string, ok bool) // varchar(64)
	PrimaryKey() (isPrimaryKey bool, ok bool)
	AutoIncrement() (isAutoIncrement bool, ok bool)
	Length() (length int64, ok bool)
	DecimalSize() (precision int64, scale int64, ok bool)
	Nullable() (nullable bool, ok bool)
	Unique() (unique bool, ok bool)
	ScanType() reflect.Type
	Comment() (value string, ok bool)
	DefaultValue() (value string, ok bool)
}

// Dal aims to facilitate an isolation between DBS and our System by defining a set of operations should a DBS provide
type Dal interface {
	// AutoMigrate runs auto migration for given entity
	AutoMigrate(entity interface{}, clauses ...Clause) errors.Error
	// AddColumn add column for the table
	AddColumn(table, columnName, columnType string) errors.Error
	// DropColumns drop column from the table
	DropColumns(table string, columnName ...string) errors.Error
	// Exec executes raw sql query
	Exec(query string, params ...interface{}) errors.Error
	// RawCursor executes raw sql query and returns a database cursor
	RawCursor(query string, params ...interface{}) (*sql.Rows, errors.Error)
	// Cursor returns a database cursor, cursor is especially useful when handling big amount of rows of data
	Cursor(clauses ...Clause) (*sql.Rows, errors.Error)
	// Fetch loads row data from `cursor` into `dst`
	Fetch(cursor *sql.Rows, dst interface{}) errors.Error
	// All loads matched rows from database to `dst`, USE IT WITH COUTIOUS!!
	All(dst interface{}, clauses ...Clause) errors.Error
	// First loads first matched row from database to `dst`, error will be returned if no records were found
	First(dst interface{}, clauses ...Clause) errors.Error
	// All loads matched rows from database to `dst`, USE IT WITH COUTIOUS!!
	Count(clauses ...Clause) (int64, errors.Error)
	// Pluck used to query single column
	Pluck(column string, dest interface{}, clauses ...Clause) errors.Error
	// Create insert record to database
	Create(entity interface{}, clauses ...Clause) errors.Error
	// Update updates record
	Update(entity interface{}, clauses ...Clause) errors.Error
	// UpdateColumn allows you to update mulitple records
	UpdateColumn(entity interface{}, columnName string, value interface{}, clauses ...Clause) errors.Error
	// UpdateColumn allows you to update multiple columns of mulitple records
	UpdateColumns(entity interface{}, clauses ...Clause) errors.Error
	// CreateOrUpdate tries to create the record, or fallback to update all if failed
	CreateOrUpdate(entity interface{}, clauses ...Clause) errors.Error
	// CreateIfNotExist tries to create the record if not exist
	CreateIfNotExist(entity interface{}, clauses ...Clause) errors.Error
	// Delete records from database
	Delete(entity interface{}, clauses ...Clause) errors.Error
	// AllTables returns all tables in database
	AllTables() ([]string, errors.Error)
	// DropTables drops all specified tables
	DropTables(dst ...interface{}) errors.Error
	// RenameTable renames table name
	RenameTable(oldName, newName string) errors.Error
	// GetColumns returns table columns in database
	GetColumns(dst Tabler, filter func(columnMeta ColumnMeta) bool) (cms []ColumnMeta, err errors.Error)
	// GetPrimarykeyFields get the PrimaryKey from `gorm` tag
	GetPrimaryKeyFields(t reflect.Type) []reflect.StructField
	// RenameColumn renames column name for specified table
	RenameColumn(table, oldColumnName, newColumnName string) errors.Error
	// DropIndexes drops all specified tables
	DropIndexes(table string, indexes ...string) errors.Error
	// Dialect returns the dialect of current database
	Dialect() string
}

// GetColumnNames returns table Column Names in database
func GetColumnNames(d Dal, dst Tabler, filter func(columnMeta ColumnMeta) bool) (names []string, err errors.Error) {
	columns, err := d.GetColumns(dst, filter)
	if err != nil {
		return
	}
	for _, pkColumn := range columns {
		names = append(names, pkColumn.Name())
	}
	return
}

// GetPrimarykeyColumns get returns PrimaryKey table Meta in database
func GetPrimarykeyColumns(d Dal, dst Tabler) ([]ColumnMeta, errors.Error) {
	return d.GetColumns(dst, func(columnMeta ColumnMeta) bool {
		isPrimaryKey, ok := columnMeta.PrimaryKey()
		return isPrimaryKey && ok
	})
}

// GetPrimarykeyColumnNames get returns PrimaryKey Column Names in database
func GetPrimarykeyColumnNames(d Dal, dst Tabler) (names []string, err errors.Error) {
	pkColumns, err := GetPrimarykeyColumns(d, dst)
	if err != nil {
		return
	}
	for _, pkColumn := range pkColumns {
		names = append(names, pkColumn.Name())
	}
	return
}

type DalClause struct {
	Expr   string
	Params []interface{}
}

const JoinClause string = "Join"

// Join creates a new JoinClause
func Join(clause string, params ...interface{}) Clause {
	return Clause{Type: JoinClause, Data: DalClause{clause, params}}
}

const WhereClause string = "Where"

// Where creates a new WhereClause
func Where(clause string, params ...interface{}) Clause {
	return Clause{Type: WhereClause, Data: DalClause{clause, params}}
}

const LimitClause string = "Limit"

// Limit creates a new LimitClause
func Limit(limit int) Clause {
	return Clause{Type: LimitClause, Data: limit}
}

const OffsetClause string = "Offset"

// Offset creates a new OffsetClause
func Offset(offset int) Clause {
	return Clause{Type: OffsetClause, Data: offset}
}

const FromClause string = "From"

// From creates a new TableClause
func From(table interface{}) Clause {
	return Clause{Type: FromClause, Data: table}
}

const SelectClause string = "Select"

// Select creates a new TableClause
func Select(fields string) Clause {
	return Clause{Type: SelectClause, Data: fields}
}

const OrderbyClause string = "OrderBy"

// Orderby creates a new Orderby clause
func Orderby(expr string) Clause {
	return Clause{Type: OrderbyClause, Data: expr}
}

const GroupbyClause string = "GroupBy"

// Groupby creates a new Groupby clause
func Groupby(expr string) Clause {
	return Clause{Type: GroupbyClause, Data: expr}
}

const HavingClause string = "Having"

// Having creates a new Having clause
func Having(clause string, params ...interface{}) Clause {
	return Clause{Type: HavingClause, Data: DalClause{clause, params}}
}
