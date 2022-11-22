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

package experimental

import (
	"fmt"
	"github.com/apache/incubator-devlake/helpers/e2ehelper"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
	"strings"
)

// TestRawData FIXME
type TestRawData struct {
	helper.RawData
	tableName string
}

// TableName FIXME
func (t TestRawData) TableName() string {
	if strings.HasPrefix(t.tableName, "_raw_") {
		return t.tableName
	}
	return "_raw_" + t.tableName
}

func newTestRawData(tableName string) TestRawData {
	return TestRawData{
		tableName: tableName,
	}
}

// InitCSVsForE2E creates csv files for raw, tools, and domain models existing in the E2E DB
func InitCSVsForE2E(rawTables []string, datamodels ...core.Tabler) {
	dataflowTester := e2ehelper.NewDataFlowTester(nil, "", nil)
	for _, rawTable := range rawTables {
		d := newTestRawData(rawTable)
		dataflowTester.CreateSnapshot(
			d,
			e2ehelper.TableOptions{
				CSVRelPath: fmt.Sprintf("./raw_tables/%s.csv", d.TableName()),
			},
		)
	}
	for _, datamodel := range datamodels {
		dataflowTester.CreateSnapshot(
			datamodel,
			e2ehelper.TableOptions{
				CSVRelPath: fmt.Sprintf("./snapshot_tables/%s.csv", datamodel.TableName()),
			},
		)
	}
}
