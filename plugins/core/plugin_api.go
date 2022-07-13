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

package core

import (
	"net/http"
	"net/url"
)

// Contains api request information
type ApiResourceInput struct {
	Params  map[string]string      // path variables
	Query   url.Values             // query string
	Body    map[string]interface{} // json body
	Request *http.Request
}

// OutputFile is the file returned
type OutputFile struct {
	ContentType string
	Data        []byte
}

// Describe response data of a api
type ApiResourceOutput struct {
	Body   interface{} // response body
	Status int
	File   *OutputFile
}

type ApiResourceHandler func(input *ApiResourceInput) (*ApiResourceOutput, error)

// Implement this interface if plugin offered API
// Code sample to register a api on `sources/:connectionId`:
// func (plugin Jira) ApiResources() map[string]map[string]core.ApiResourceHandler {
// 	return map[string]map[string]core.ApiResourceHandler{
// 		"connections/:connectionId": {
// 			"PUT":    api.PutConnection,
// 			"DELETE": api.DeleteConnection,
// 			"GET":    api.GetConnection,
// 		},
// 	}
// }
type PluginApi interface {
	ApiResources() map[string]map[string]ApiResourceHandler
}
