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

package plugin

import (
	"fmt"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	coreModels "github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/models/common"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/server/services/remote/bridge"
	"github.com/apache/incubator-devlake/server/services/remote/models"
	"github.com/apache/incubator-devlake/server/services/remote/plugin/doc"
)

type (
	remotePluginImpl struct {
		name                     string
		subtaskMetas             []plugin.SubTaskMeta
		pluginPath               string
		description              string
		invoker                  bridge.Invoker
		connectionTabler         *coreModels.DynamicTabler
		scopeTabler              *coreModels.DynamicTabler
		transformationRuleTabler *coreModels.DynamicTabler
		resources                map[string]map[string]plugin.ApiResourceHandler
		openApiSpec              string
		tables                   []dal.Tabler
	}
	RemotePluginTaskData struct {
		DbUrl              string                 `json:"db_url"`
		Scope              interface{}            `json:"scope"`
		Connection         interface{}            `json:"connection"`
		TransformationRule interface{}            `json:"transformation_rule"`
		Options            map[string]interface{} `json:"options"`
	}
)

func newPlugin(info *models.PluginInfo, invoker bridge.Invoker) (*remotePluginImpl, errors.Error) {
	connectionTabler, err := info.ConnectionModelInfo.LoadDynamicTabler(common.Model{})
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("Couldn't load Connection type for plugin %s", info.Name))
	}

	var txRuleTabler *coreModels.DynamicTabler
	if info.TransformationRuleModelInfo != nil {
		txRuleTabler, err = info.TransformationRuleModelInfo.LoadDynamicTabler(models.TransformationModel{})
		if err != nil {
			return nil, errors.Default.Wrap(err, fmt.Sprintf("Couldn't load TransformationRule type for plugin %s", info.Name))
		}
	}
	scopeTabler, err := info.ScopeModelInfo.LoadDynamicTabler(models.ScopeModel{})
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("Couldn't load Scope type for plugin %s", info.Name))
	}
	openApiSpec, err := doc.GenerateOpenApiSpec(info)
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("Couldn't generate OpenAPI spec for plugin %s", info.Name))
	}
	p := remotePluginImpl{
		name:                     info.Name,
		invoker:                  invoker,
		pluginPath:               info.PluginPath,
		description:              info.Description,
		connectionTabler:         connectionTabler,
		scopeTabler:              scopeTabler,
		transformationRuleTabler: txRuleTabler,
		resources:                GetDefaultAPI(invoker, connectionTabler, txRuleTabler, scopeTabler, connectionHelper),
		openApiSpec:              *openApiSpec,
	}
	remoteBridge := bridge.NewBridge(invoker)
	for _, subtask := range info.SubtaskMetas {
		p.subtaskMetas = append(p.subtaskMetas, plugin.SubTaskMeta{
			Name:             subtask.Name,
			EntryPoint:       remoteBridge.RemoteSubtaskEntrypointHandler(subtask),
			Required:         subtask.Required,
			EnabledByDefault: subtask.EnabledByDefault,
			Description:      subtask.Description,
			DomainTypes:      subtask.DomainTypes,
		})
	}
	for _, tableName := range info.Tables {
		p.tables = append(p.tables, coreModels.NewDynamicTabler(tableName, nil))
	}
	return &p, nil
}

func (p *remotePluginImpl) SubTaskMetas() []plugin.SubTaskMeta {
	return p.subtaskMetas
}

func (p *remotePluginImpl) GetTablesInfo() []dal.Tabler {
	return p.tables
}

func (p *remotePluginImpl) PrepareTaskData(taskCtx plugin.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	dbUrl := taskCtx.GetConfig("db_url")
	connectionId := uint64(options["connectionId"].(float64))

	helper := api.NewConnectionHelper(
		taskCtx,
		nil,
	)

	wrappedConnection := p.connectionTabler.New()
	err := helper.FirstById(wrappedConnection, connectionId)
	if err != nil {
		return nil, errors.Convert(err)
	}
	connection := wrappedConnection.Unwrap()

	scopeId, ok := options["scopeId"].(string)
	if !ok {
		return nil, errors.BadInput.New("missing scopeId")
	}

	db := taskCtx.GetDal()
	wrappedScope := p.scopeTabler.New()
	err = api.CallDB(db.First, wrappedScope, dal.Where("connection_id = ? AND id = ?", connectionId, scopeId))
	if err != nil {
		return nil, errors.BadInput.New("Invalid scope id")
	}
	var scope models.ScopeModel
	err = wrappedScope.To(&scope)
	if err != nil {
		return nil, err
	}

	txRule, err := p.getTxRule(db, scope)
	if err != nil {
		return nil, err
	}

	return RemotePluginTaskData{
		DbUrl:              dbUrl,
		Scope:              wrappedScope.Unwrap(),
		Connection:         connection,
		TransformationRule: txRule,
		Options:            options,
	}, nil
}

func (p *remotePluginImpl) getTxRule(db dal.Dal, scope models.ScopeModel) (interface{}, errors.Error) {
	if scope.TransformationRuleId > 0 {
		if p.transformationRuleTabler == nil {
			return nil, errors.Default.New(fmt.Sprintf("Cannot load transformation rule %v: plugin %s has no transformation rule model", scope.TransformationRuleId, p.name))
		}
		wrappedTxRule := p.transformationRuleTabler.New()
		err := api.CallDB(db.First, wrappedTxRule, dal.From(p.transformationRuleTabler.TableName()), dal.Where("id = ?", scope.TransformationRuleId))
		if err != nil {
			return nil, err
		}
		return wrappedTxRule.Unwrap(), nil
	} else {
		return nil, nil
	}
}

func (p *remotePluginImpl) Description() string {
	return p.description
}

func (p *remotePluginImpl) RootPkgPath() string {
	// RootPkgPath is used by DomainIdGenerator to find the name of the plugin that defines a given type.
	// While remote plugins do not use the DomainIdGenerator, we still need to implement this function.
	// Indeed, DomainIdGenerator uses FindPluginNameBySubPkgPath that returns the name of the first plugin
	// whose RootPkgPath is a prefix of the type package path.
	// So we forge a fake package path that is not a prefix of any go plugin package path.
	return "github.com/apache/incubator-devlake/services/remote/fakepackages/" + p.name
}

func (p *remotePluginImpl) ApiResources() map[string]map[string]plugin.ApiResourceHandler {
	return p.resources
}

func (p *remotePluginImpl) RunMigrations(forceMigrate bool) errors.Error {
	err := api.CallDB(basicRes.GetDal().AutoMigrate, p.connectionTabler.New())
	if err != nil {
		return err
	}
	err = api.CallDB(basicRes.GetDal().AutoMigrate, p.scopeTabler.New())
	if err != nil {
		return err
	}
	if p.transformationRuleTabler != nil {
		err = api.CallDB(basicRes.GetDal().AutoMigrate, p.transformationRuleTabler.New())
		if err != nil {
			return err
		}
	}
	dbUrl := basicRes.GetConfig("db_url")
	err = p.invoker.Call("run-migrations", bridge.DefaultContext, dbUrl, forceMigrate).Err
	return err
}

func (p *remotePluginImpl) OpenApiSpec() string {
	return p.openApiSpec
}

var _ models.RemotePlugin = (*remotePluginImpl)(nil)
