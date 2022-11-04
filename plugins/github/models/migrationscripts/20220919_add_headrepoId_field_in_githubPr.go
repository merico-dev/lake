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
	"github.com/apache/incubator-devlake/plugins/core"
)

type pullRequest20220919 struct {
	HeadRepoId int
}

func (pullRequest20220919) TableName() string {
	return "_tool_github_pull_requests"
}

type addHeadRepoIdFieldInGithubPr struct{}

func (*addHeadRepoIdFieldInGithubPr) Up(basicRes core.BasicRes) errors.Error {
	err := basicRes.GetDal().AutoMigrate(&pullRequest20220919{})
	if err != nil {
		return err
	}
	return nil
}

func (*addHeadRepoIdFieldInGithubPr) Version() uint64 {
	return 20220919223048
}

func (*addHeadRepoIdFieldInGithubPr) Name() string {
	return "add column `head_repo_id` at tool_github_pull_requests"
}
