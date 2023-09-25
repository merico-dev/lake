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
	"fmt"
	"reflect"

	"github.com/apache/incubator-devlake/core/errors"
)

const (
	Varchar ColumnType = "varchar(255)"
	Text    ColumnType = "text"
	Int     ColumnType = "bigint"
	Time    ColumnType = "timestamp"
	Float   ColumnType = "float"
)

var columnTypes = map[string]ColumnType{
	Varchar.String(): Varchar,
	Text.String():    Text,
	Int.String():     Int,
	Time.String():    Time,
	Float.String():   Float,
}

type ColumnType string

func (c ColumnType) String() string {
	return string(c)
}

// ToColumnType converts a string to ColumnType
func ToColumnType(s string) (ColumnType, bool) {
	t, ok := columnTypes[s]
	return t, ok
}

type Tabler interface {
	TableName() string
}

// DefaultTabler is working for the Tabler interface witch only need TableName
type DefaultTabler struct {
	Name string
}

var _ Tabler = (*DefaultTabler)(nil)

func (d DefaultTabler) TableName() string {
	return d.Name
}

// Clause represents SQL Clause
type Clause struct {
	Type string
	Data interface{}
}

// ClauseColumn quote with name
type ClauseColumn struct {
	Table string
	Name  string
	Alias string
	Raw   bool
}

// ClauseTable quote with name
type ClauseTable struct {
	Name  string
	Alias string
	Raw   bool
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

// SessionConfig specify options for the session
type SessionConfig struct {
	PrepareStmt            bool
	SkipDefaultTransaction bool
}

// Dal aims to facilitate an isolation between DBS and our System by defining a set of operations should a DBS provide
type Dal interface {
	// AutoMigrate runs auto migration for given entity
	AutoMigrate(entity interface{}, clauses ...Clause) errors.Error
	// AddColumn add column for the table
	AddColumn(table, columnName string, columnType ColumnType) errors.Error
	// DropColumns drop column from the table
	DropColumns(table string, columnName ...string) errors.Error
	// Exec executes raw sql query
	Exec(query string, params ...interface{}) errors.Error
	// Cursor returns a database cursor, cursor is especially useful when handling big amount of rows of data
	Cursor(clauses ...Clause) (Rows, errors.Error)
	// Fetch loads row data from `cursor` into `dst`
	Fetch(cursor Rows, dst interface{}) errors.Error
	// All loads matched rows from database to `dst`, USE IT WITH CAUTIOUS!!
	All(dst interface{}, clauses ...Clause) errors.Error
	// First loads first matched row from database to `dst`, error will be returned if no records were found
	First(dst interface{}, clauses ...Clause) errors.Error
	// Count matched rows from database
	Count(clauses ...Clause) (int64, errors.Error)
	// Pluck used to query single column
	Pluck(column string, dest interface{}, clauses ...Clause) errors.Error
	// Create insert record to database
	Create(entity interface{}, clauses ...Clause) errors.Error
	// CreateWithMap insert record to database, the record is organized as map
	CreateWithMap(entity interface{}, record map[string]interface{}) errors.Error
	// Update updates record
	Update(entity interface{}, clauses ...Clause) errors.Error
	// UpdateColumn allows you to update multiple records
	UpdateColumn(entityOrTable interface{}, columnName string, value interface{}, clauses ...Clause) errors.Error
	// UpdateColumns allows you to update multiple columns of multiple records
	UpdateColumns(entityOrTable interface{}, set []DalSet, clauses ...Clause) errors.Error
	// UpdateAllColumn updated all Columns of entity
	UpdateAllColumn(entity interface{}, clauses ...Clause) errors.Error
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
	// HasTable checks if table exists
	HasTable(table interface{}) bool
	// HasColumn checks if column exists
	HasColumn(table interface{}, columnName string) bool
	// RenameTable renames table name
	RenameTable(oldName, newName string) errors.Error
	// GetColumns returns table columns in database
	GetColumns(dst Tabler, filter func(columnMeta ColumnMeta) bool) (cms []ColumnMeta, err errors.Error)
	// GetPrimaryKeyFields get the PrimaryKey from `gorm` tag
	GetPrimaryKeyFields(t reflect.Type) []reflect.StructField
	// RenameColumn renames column name for specified table
	RenameColumn(table, oldColumnName, newColumnName string) errors.Error
	// DropIndexes drops all specified tables
	DropIndexes(table string, indexes ...string) errors.Error
	// Dialect returns the dialect of current database
	Dialect() string
	// Session creates a new manual session for special scenarios
	Session(config SessionConfig) Dal
	// Begin create a new transaction
	Begin() Transaction
	// IsErrorNotFound returns true if error is record-not-found
	IsErrorNotFound(err error) bool
	// IsDuplicationError returns true if error is duplicate-error
	IsDuplicationError(err error) bool
	// RawCursor (Deprecated) executes raw sql query and returns a database cursor.
	RawCursor(query string, params ...interface{}) (*sql.Rows, errors.Error)
}

type LockTable struct {
	Table     interface{}
	Exclusive bool
}

func (l *LockTable) TableName() string {
	switch table := l.Table.(type) {
	case Tabler:
		return table.TableName()
	case string:
		return table
	default:
		panic("either Table or Tabler must be specified")
	}
}

type LockTables []*LockTable

type Transaction interface {
	Dal
	Rollback() errors.Error
	Commit() errors.Error
	// table: exclusive
	LockTables(lockTables LockTables) errors.Error
	UnlockTables() errors.Error
	// End(err *errors.Error)
}

type Rows interface {
	// Next prepares the next result row for reading with the Scan method. It
	// returns true on success, or false if there is no next result row or an error
	// happened while preparing it. Err should be consulted to distinguish between
	// the two cases.
	//
	// Every call to Scan, even the first one, must be preceded by a call to Next.
	Next() bool

	// Close closes the Rows, preventing further enumeration. If Next is called
	// and returns false and there are no further result sets,
	// the Rows are closed automatically and it will suffice to check the
	// result of Err. Close is idempotent and does not affect the result of Err.
	Close() error

	// Scan copies the columns in the current row into the values pointed at by dest.
	// The number of values in dest must be the same as the number of columns in Rows.
	Scan(dest ...any) error

	// Columns returns the column names.
	// Columns returns an error if the rows are closed.
	Columns() ([]string, error)

	// ColumnTypes returns column information such as column type, length,
	// and nullable. Some information may not be available from some drivers.
	ColumnTypes() ([]*sql.ColumnType, error)
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
		// in case the column name is a reserved identifier
		names = append(names, fmt.Sprintf("%s.%s", dst.TableName(), pkColumn.Name()))
	}
	return
}

type DalClause struct {
	Expr   string
	Params []interface{}
}

type DalSet struct {
	ColumnName string
	Value      interface{}
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
func From(table interface{}, params ...interface{}) Clause {
	if len(params) == 0 {
		return Clause{Type: FromClause, Data: table}
	} else {
		return Clause{Type: FromClause, Data: DalClause{table.(string), params}}
	}
}

const SelectClause string = "Select"

// Select creates a new TableClause
func Select(clause string, params ...interface{}) Clause {
	return Clause{Type: SelectClause, Data: DalClause{clause, params}}
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

const LockClause string = "Lock"

// Having creates a new Having clause
func Lock(write bool, nowait bool) Clause {
	return Clause{Type: LockClause, Data: []bool{write, nowait}}
}

func Expr(expr string, params ...interface{}) DalClause {
	return DalClause{Expr: expr, Params: params}
}
