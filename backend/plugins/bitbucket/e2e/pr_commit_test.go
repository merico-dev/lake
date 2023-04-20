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

package e2e

import (
	"github.com/apache/incubator-devlake/core/models/common"
	"github.com/apache/incubator-devlake/core/models/domainlayer/code"
	"github.com/apache/incubator-devlake/helpers/e2ehelper"
	"github.com/apache/incubator-devlake/plugins/bitbucket/impl"
	"github.com/apache/incubator-devlake/plugins/bitbucket/models"
	"github.com/apache/incubator-devlake/plugins/bitbucket/tasks"
	"testing"
)

func TestPrCommitDataFlow(t *testing.T) {
	var plugin impl.Bitbucket
	dataflowTester := e2ehelper.NewDataFlowTester(t, "bitbucket", plugin)

	taskData := &tasks.BitbucketTaskData{
		Options: &tasks.BitbucketOptions{
			ConnectionId: 1,
			FullName:     "likyh/likyhphp",
		},
	}

	// import raw data table
	dataflowTester.ImportCsvIntoRawTable("./raw_tables/_raw_bitbucket_api_pull_request_commits.csv", "_raw_bitbucket_api_pull_request_commits")

	// verify pr extraction
	dataflowTester.FlushTabler(&models.BitbucketPrCommit{})
	dataflowTester.FlushTabler(&models.BitbucketCommit{})
	dataflowTester.FlushTabler(&models.BitbucketRepoCommit{})
	dataflowTester.Subtask(tasks.ExtractApiPrCommitsMeta, taskData)
	dataflowTester.VerifyTableWithOptions(
		models.BitbucketPrCommit{}, e2ehelper.TableOptions{
			CSVRelPath:  "./snapshot_tables/_tool_bitbucket_pull_request_commits.csv",
			IgnoreTypes: []interface{}{common.NoPKModel{}},
		},
	)

	// verify pr conversion
	dataflowTester.FlushTabler(&code.PullRequestCommit{})
	dataflowTester.Subtask(tasks.ConvertPrCommitsMeta, taskData)
	dataflowTester.VerifyTableWithOptions(
		code.PullRequestCommit{},
		e2ehelper.TableOptions{
			CSVRelPath:  "./snapshot_tables/pull_request_commits.csv",
			IgnoreTypes: []interface{}{common.NoPKModel{}},
		},
	)
}
