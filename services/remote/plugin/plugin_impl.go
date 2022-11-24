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
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/services/remote/bridge"
	"github.com/apache/incubator-devlake/services/remote/models"
)

type (
	remotePluginImpl struct {
		resources    map[string]map[string]core.ApiResourceHandler
		subtaskMetas []core.SubTaskMeta
		pluginPath   string
		description  string
		rootPkgPath  string
		invoker      bridge.Invoker
	}
)

func newPlugin(info *models.PluginInfo, invoker bridge.Invoker) *remotePluginImpl {
	p := remotePluginImpl{
		invoker:     invoker,
		pluginPath:  info.PluginPath,
		description: info.Description,
		rootPkgPath: "", //TODO how to resolve this?
		resources:   GetDefaultAPI(models.LoadTableModel(info.Connection), connectionHelper),
	}
	remoteBridge := bridge.NewBridge(invoker)
	for _, endpoint := range info.ApiEndpoints {
		var ok bool
		var methodMap map[string]core.ApiResourceHandler
		if methodMap, ok = p.resources[endpoint.Resource]; !ok {
			methodMap = map[string]core.ApiResourceHandler{}
			p.resources[endpoint.Resource] = methodMap
		}
		methodMap[endpoint.Method] = remoteBridge.RemoteAPIHandler(&endpoint)
	}
	for _, subtask := range info.SubtaskMetas {
		p.subtaskMetas = append(p.subtaskMetas, core.SubTaskMeta{
			Name:             subtask.Name,
			EntryPoint:       remoteBridge.RemoteSubtaskEntrypointHandler(&subtask),
			Required:         subtask.Required,
			EnabledByDefault: subtask.EnabledByDefault,
			Description:      subtask.Description,
			DomainTypes:      subtask.DomainTypes,
		})
	}
	return &p
}

func (p *remotePluginImpl) SubTaskMetas() []core.SubTaskMeta {
	return p.subtaskMetas
}

func (p *remotePluginImpl) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	var output map[string]any
	err := p.invoker.Call("PrepareTaskData", taskCtx, options).Get(&output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (p *remotePluginImpl) Description() string {
	return p.description
}

func (p *remotePluginImpl) RootPkgPath() string {
	return p.rootPkgPath
}

func (p *remotePluginImpl) ApiResources() map[string]map[string]core.ApiResourceHandler {
	return p.resources
}

func (p *remotePluginImpl) RunMigrations(forceMigrate bool) errors.Error {
	err := p.invoker.Call("RunMigrations", bridge.DefaultContext, forceMigrate).Get()
	return err
}

var _ models.RemotePlugin = (*remotePluginImpl)(nil)
var _ core.Scope = (*models.WrappedPipelineScope)(nil)
