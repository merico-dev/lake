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

package migrationscripts

import (
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/helpers/migrationhelper"
	"github.com/apache/incubator-devlake/models/migrationscripts/archived"
	"github.com/apache/incubator-devlake/plugins/core"
)

type addDoraBenchmark struct{}

type doraBenchmark struct {
	archived.Model
	Metric string `gorm:"type:varchar(255)"`
	Low    string `gorm:"type:varchar(255)"`
	Medium string `gorm:"type:varchar(255)"`
	High   string `gorm:"type:varchar(255)"`
	Elite  string `gorm:"type:varchar(255)"`
}

func (doraBenchmark) TableName() string {
	return "dora_benchmarks"
}

func (u *addDoraBenchmark) Up(baseRes core.BasicRes) errors.Error {
	db := baseRes.GetDal()
	err := migrationhelper.AutoMigrateTables(
		baseRes,
		&doraBenchmark{},
	)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = db.DropTables(&doraBenchmark{})
			if err != nil {
				return
			}
		}
	}()
	doraBenchmarkDF := &doraBenchmark{
		Model: archived.Model{
			ID: 1,
		},
		Metric: "Deployment frequency",
		Low:    "Fewer than once per six months",
		Medium: "Between once per month and once every 6 months",
		High:   "Between once per week and once per month",
		Elite:  "On-demand",
	}
	err = db.Create(doraBenchmarkDF)
	if err != nil {
		return errors.Convert(err)
	}
	doraBenchmarkLTC := &doraBenchmark{
		Model: archived.Model{
			ID: 2,
		},
		Metric: "Lead time for changes",
		Low:    "More than six months",
		Medium: "Between one week and six months",
		High:   "Less than one week",
		Elite:  "Less than one hour",
	}
	err = db.Create(doraBenchmarkLTC)
	if err != nil {
		return errors.Convert(err)
	}
	doraBenchmarkTTS := &doraBenchmark{
		Model: archived.Model{
			ID: 3,
		},
		Metric: "Time to restore service",
		Low:    "More than one week",
		Medium: "Between one day and one week",
		High:   "Less than one day",
		Elite:  "Less than one hour",
	}
	err = db.Create(doraBenchmarkTTS)
	if err != nil {
		return errors.Convert(err)
	}
	doraBenchmarkCFR := &doraBenchmark{
		Model: archived.Model{
			ID: 4,
		},
		Metric: "Change failure rate",
		Low:    "> 30%",
		Medium: "21%-30%",
		High:   "16%-20%",
		Elite:  "0-15%",
	}
	err = db.Create(doraBenchmarkCFR)
	if err != nil {
		return errors.Convert(err)
	}
	return errors.Convert(err)
}

func (*addDoraBenchmark) Version() uint64 {
	return 20220928000001
}

func (*addDoraBenchmark) Name() string {
	return "add dora benchmark"
}
