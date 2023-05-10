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

package helper

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/server/api/blueprints"
	apiProject "github.com/apache/incubator-devlake/server/api/project"
	"github.com/stretchr/testify/require"
)

// CreateConnection FIXME
func (d *DevlakeClient) TestConnection(pluginName string, connection any) {
	d.testCtx.Helper()
	_ = sendHttpRequest[Connection](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPost, fmt.Sprintf("%s/plugins/%s/test", d.Endpoint, pluginName), nil, connection)
}

// CreateConnection FIXME
func (d *DevlakeClient) CreateConnection(pluginName string, connection any) *Connection {
	d.testCtx.Helper()
	created := sendHttpRequest[Connection](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPost, fmt.Sprintf("%s/plugins/%s/connections", d.Endpoint, pluginName), nil, connection)
	return &created
}

// ListConnections FIXME
func (d *DevlakeClient) ListConnections(pluginName string) []*Connection {
	d.testCtx.Helper()
	all := sendHttpRequest[[]*Connection](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodGet, fmt.Sprintf("%s/plugins/%s/connections", d.Endpoint, pluginName), nil, nil)
	return all
}

// CreateBasicBlueprintV2 FIXME
func (d *DevlakeClient) CreateBasicBlueprintV2(name string, config *BlueprintV2Config) models.Blueprint {
	settings := &models.BlueprintSettings{
		Version:   "2.0.0",
		TimeAfter: config.TimeAfter,
		Connections: ToJson([]*plugin.BlueprintConnectionV200{
			config.Connection,
		}),
	}
	blueprint := models.Blueprint{
		Name:        name,
		ProjectName: config.ProjectName,
		Mode:        models.BLUEPRINT_MODE_NORMAL,
		Plan:        nil,
		Enable:      true,
		CronConfig:  "manual",
		IsManual:    true,
		SkipOnFail:  config.SkipOnFail,
		Labels:      []string{"test-label"},
		Settings:    ToJson(settings),
	}
	d.testCtx.Helper()
	blueprint = sendHttpRequest[models.Blueprint](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPost, fmt.Sprintf("%s/blueprints", d.Endpoint), nil, &blueprint)
	return blueprint
}

func (d *DevlakeClient) ListBlueprints() blueprints.PaginatedBlueprint {
	return sendHttpRequest[blueprints.PaginatedBlueprint](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodGet, fmt.Sprintf("%s/blueprints", d.Endpoint), nil, nil)
}

func (d *DevlakeClient) GetBlueprint(blueprintId uint64) models.Blueprint {
	return sendHttpRequest[models.Blueprint](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodGet, fmt.Sprintf("%s/blueprint/%d", d.Endpoint, blueprintId), nil, nil)
}

func (d *DevlakeClient) CreateProject(project *ProjectConfig) models.ApiOutputProject {
	var metrics []models.BaseMetric
	doraSeen := false
	for _, p := range project.MetricPlugins {
		if p.Name == "dora" {
			doraSeen = true
		}
		metrics = append(metrics, models.BaseMetric{
			PluginName:   p.Name,
			PluginOption: string(ToJson(p.Options)),
			Enable:       true,
		})
	}
	if project.EnableDora && !doraSeen {
		metrics = append(metrics, models.BaseMetric{
			PluginName:   "dora",
			PluginOption: string(ToJson(nil)),
			Enable:       true,
		})
	}
	return sendHttpRequest[models.ApiOutputProject](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPost, fmt.Sprintf("%s/projects", d.Endpoint), nil, &models.ApiInputProject{
		BaseProject: models.BaseProject{
			Name:        project.ProjectName,
			Description: project.ProjectDescription,
		},
		Enable:  Val(true),
		Metrics: &metrics,
	})
}

func (d *DevlakeClient) GetProject(projectName string) models.ApiOutputProject {
	return sendHttpRequest[models.ApiOutputProject](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodGet, fmt.Sprintf("%s/projects/%s", d.Endpoint, projectName), nil, nil)
}

func (d *DevlakeClient) ListProjects() apiProject.PaginatedProjects {
	return sendHttpRequest[apiProject.PaginatedProjects](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodGet, fmt.Sprintf("%s/projects", d.Endpoint), nil, nil)
}

func (d *DevlakeClient) CreateScope(pluginName string, connectionId uint64, scopes ...any) any {
	request := map[string]any{
		"data": scopes,
	}
	return sendHttpRequest[any](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPut, fmt.Sprintf("%s/plugins/%s/connections/%d/scopes", d.Endpoint, pluginName, connectionId), nil, request)
}

func (d *DevlakeClient) UpdateScope(pluginName string, connectionId uint64, scopeId string, scope any) any {
	return sendHttpRequest[any](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPatch, fmt.Sprintf("%s/plugins/%s/connections/%d/scopes/%s", d.Endpoint, pluginName, connectionId, scopeId), nil, scope)
}

func (d *DevlakeClient) ListScopes(pluginName string, connectionId uint64, listBlueprints bool) []ScopeResponse {
	scopesRaw := sendHttpRequest[[]map[string]any](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodGet, fmt.Sprintf("%s/plugins/%s/connections/%d/scopes?blueprints=%v", d.Endpoint, pluginName, connectionId, listBlueprints), nil, nil)
	var responses []ScopeResponse
	for _, scopeRaw := range scopesRaw {
		responses = append(responses, getScopeResponse(scopeRaw))
	}
	return responses
}

func (d *DevlakeClient) GetScope(pluginName string, connectionId uint64, scopeId string, listBlueprints bool) any {
	return sendHttpRequest[api.ScopeRes[any]](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodGet, fmt.Sprintf("%s/plugins/%s/connections/%d/scopes/%s?blueprints=%v", d.Endpoint, pluginName, connectionId, scopeId, listBlueprints), nil, nil)
}

func (d *DevlakeClient) DeleteScope(pluginName string, connectionId uint64, scopeId string, deleteDataOnly bool) []models.Blueprint {
	return sendHttpRequest[[]models.Blueprint](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodDelete, fmt.Sprintf("%s/plugins/%s/connections/%d/scopes/%s?delete_data_only=%v", d.Endpoint, pluginName, connectionId, scopeId, deleteDataOnly), nil, nil)
}

func (d *DevlakeClient) CreateTransformationRule(pluginName string, connectionId uint64, rules any) any {
	return sendHttpRequest[any](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPost, fmt.Sprintf("%s/plugins/%s/connections/%d/transformation_rules",
		d.Endpoint, pluginName, connectionId), nil, rules)
}

func (d *DevlakeClient) PatchTransformationRule(pluginName string, connectionId uint64, txRuleId uint64, rule any) any {
	return sendHttpRequest[any](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPatch, fmt.Sprintf("%s/plugins/%s/connections/%d/transformation_rules/%d",
		d.Endpoint, pluginName, connectionId, txRuleId), nil, rule)
}

func (d *DevlakeClient) ListTransformationRules(pluginName string, connectionId uint64) []any {
	return sendHttpRequest[[]any](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodGet, fmt.Sprintf("%s/plugins/%s/connections/%d/transformation_rules?pageSize=20&page=1",
		d.Endpoint, pluginName, connectionId), nil, nil)
}

func (d *DevlakeClient) GetTransformationRule(pluginName string, connectionId uint64, txRuleId uint64) any {
	return sendHttpRequest[any](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodGet, fmt.Sprintf("%s/plugins/%s/connections/%d/transformation_rules/%d",
		d.Endpoint, pluginName, connectionId, txRuleId), nil, nil)
}

func (d *DevlakeClient) RemoteScopes(query RemoteScopesQuery) RemoteScopesOutput {
	url := fmt.Sprintf("%s/plugins/%s/connections/%d/remote-scopes",
		d.Endpoint,
		query.PluginName,
		query.ConnectionId,
	)
	if query.Params == nil {
		query.Params = make(map[string]string)
	}
	if query.GroupId != "" {
		query.Params["groupId"] = query.GroupId
	}
	if query.PageToken != "" {
		query.Params["pageToken"] = query.PageToken
	}
	if len(query.Params) > 0 {
		url = url + "?" + mapToQueryString(query.Params)
	}
	return sendHttpRequest[RemoteScopesOutput](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodGet, url, nil, nil)
}

// SearchRemoteScopes makes calls to the "scope API" indirectly. "Search" is the remote endpoint to hit.
func (d *DevlakeClient) SearchRemoteScopes(query SearchRemoteScopesQuery) SearchRemoteScopesOutput {
	return sendHttpRequest[SearchRemoteScopesOutput](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodGet, fmt.Sprintf("%s/plugins/%s/connections/%d/search-remote-scopes?search=%s&page=%d&pageSize=%d&%s",
		d.Endpoint,
		query.PluginName,
		query.ConnectionId,
		query.Search,
		query.Page,
		query.PageSize,
		mapToQueryString(query.Params)),
		nil, nil)
}

// CreateBasicBlueprint FIXME
func (d *DevlakeClient) CreateBasicBlueprint(name string, connection *plugin.BlueprintConnectionV100) models.Blueprint {
	settings := &models.BlueprintSettings{
		Version:     "1.0.0",
		Connections: ToJson([]*plugin.BlueprintConnectionV100{connection}),
	}
	blueprint := models.Blueprint{
		Name:       name,
		Mode:       models.BLUEPRINT_MODE_NORMAL,
		Plan:       nil,
		Enable:     true,
		CronConfig: "manual",
		IsManual:   true,
		Settings:   ToJson(settings),
	}
	d.testCtx.Helper()
	blueprint = sendHttpRequest[models.Blueprint](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPost, fmt.Sprintf("%s/blueprints", d.Endpoint), nil, &blueprint)
	return blueprint
}

// TriggerBlueprint FIXME
func (d *DevlakeClient) TriggerBlueprint(blueprintId uint64) models.Pipeline {
	d.testCtx.Helper()
	pipeline := sendHttpRequest[models.Pipeline](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPost, fmt.Sprintf("%s/blueprints/%d/trigger", d.Endpoint, blueprintId), nil, nil)
	return d.monitorPipeline(pipeline.ID)
}

// RunPipeline FIXME
func (d *DevlakeClient) RunPipeline(pipeline models.NewPipeline) models.Pipeline {
	d.testCtx.Helper()
	pipelineResult := sendHttpRequest[models.Pipeline](d.testCtx, d.timeout, debugInfo{
		print:      true,
		inlineJson: false,
	}, http.MethodPost, fmt.Sprintf("%s/pipelines", d.Endpoint), nil, &pipeline)
	return d.monitorPipeline(pipelineResult.ID)
}

func mapToQueryString(queryParams map[string]string) string {
	params := make([]string, 0)
	for k, v := range queryParams {
		params = append(params, k+"="+v)
	}
	return strings.Join(params, "&")
}

// MonitorPipeline FIXME
func (d *DevlakeClient) monitorPipeline(id uint64) models.Pipeline {
	d.testCtx.Helper()
	var previousResult models.Pipeline
	endpoint := fmt.Sprintf("%s/pipelines/%d", d.Endpoint, id)
	coloredPrintf("calling:\n\t%s %s\nwith:\n%s\n", http.MethodGet, endpoint, string(ToCleanJson(false, nil)))
	var pipelineResult models.Pipeline
	require.NoError(d.testCtx, runWithTimeout(d.pipelineTimeout, func() (bool, errors.Error) {
		pipelineResult = sendHttpRequest[models.Pipeline](d.testCtx, d.pipelineTimeout, debugInfo{
			print: false,
		}, http.MethodGet, fmt.Sprintf("%s/pipelines/%d", d.Endpoint, id), nil, nil)
		if pipelineResult.Status == models.TASK_COMPLETED {
			coloredPrintf("result: %s\n", ToCleanJson(true, &pipelineResult))
			return true, nil
		}
		if pipelineResult.Status == models.TASK_FAILED {
			coloredPrintf("result: %s\n", ToCleanJson(true, &pipelineResult))
			return true, errors.Default.New("pipeline task failed")
		}
		if !reflect.DeepEqual(pipelineResult, previousResult) {
			coloredPrintf("result: %s\n", ToCleanJson(true, &pipelineResult))
		}
		previousResult = pipelineResult
		return false, nil
	}))
	return pipelineResult
}

func getScopeResponse(scopeRaw map[string]any) ScopeResponse {
	response := Cast[ScopeResponse](scopeRaw)
	response.Scope = scopeRaw
	return response
}
