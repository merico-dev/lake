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

package main

import (
	"strings"

	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitextractor/models"
	"github.com/apache/incubator-devlake/plugins/gitextractor/parser"
	"github.com/apache/incubator-devlake/plugins/gitextractor/store"
	"github.com/apache/incubator-devlake/plugins/gitextractor/tasks"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var _ core.PluginMeta = (*GitExtractor)(nil)
var _ core.PluginTask = (*GitExtractor)(nil)

type GitExtractor struct{}

func (plugin GitExtractor) GetTablesInfo() []core.Tabler {
	return []core.Tabler{}
}

func (plugin GitExtractor) Description() string {
	return "extract infos from git repository"
}

// return all available subtasks, framework will run them for you in order
func (plugin GitExtractor) SubTaskMetas() []core.SubTaskMeta {
	return []core.SubTaskMeta{
		tasks.CollectGitCommitMeta,
		tasks.CollectGitBranchMeta,
		tasks.CollectGitTagMeta,
		tasks.CollectGitDiffLineMeta,
	}
}

// based on task context and user input options, return data that shared among all subtasks
func (plugin GitExtractor) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	var op tasks.GitExtractorOptions
	if err := helper.Decode(options, &op, nil); err != nil {
		return nil, err
	}
	if err := op.Valid(); err != nil {
		return nil, err
	}
	storage := store.NewDatabase(taskCtx, op.Url)
	repo, err := newGitRepo(taskCtx.GetLogger(), storage, op)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (plugin GitExtractor) Close(taskCtx core.TaskContext) errors.Error {
	if repo, ok := taskCtx.GetData().(*parser.GitRepo); ok {
		if err := repo.Close(); err != nil {
			return errors.Convert(err)
		}
	}
	return nil
}

func (plugin GitExtractor) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/gitextractor"
}

func newGitRepo(logger core.Logger, storage models.Store, op tasks.GitExtractorOptions) (*parser.GitRepo, errors.Error) {
	var err errors.Error
	var repo *parser.GitRepo
	p := parser.NewGitRepoCreator(storage, logger)
	if strings.HasPrefix(op.Url, "http") {
		repo, err = p.CloneOverHTTP(op.RepoId, op.Url, op.User, op.Password, op.Proxy)
	} else if url := strings.TrimPrefix(op.Url, "ssh://"); strings.HasPrefix(url, "git@") {
		repo, err = p.CloneOverSSH(op.RepoId, url, op.PrivateKey, op.Passphrase)
	} else if strings.HasPrefix(op.Url, "/") {
		repo, err = p.LocalRepo(op.Url, op.RepoId)
	}
	return repo, err
}

// PluginEntry is a variable exported for Framework to search and load
var PluginEntry GitExtractor //nolint
