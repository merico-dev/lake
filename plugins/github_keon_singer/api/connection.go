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

package api

import (
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/github/api"
)

// TODO Please modify the following code to fit your needs
func TestConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	return api.TestConnection(input)
}

//TODO Please modify the folowing code to adapt to your plugin
/*
POST /plugins/GithubSinger/connections
{
	"name": "GithubSinger data connection name",
	"endpoint": "GithubSinger api endpoint, i.e. https://example.com",
	"username": "username, usually should be email address",
	"password": "GithubSinger api access token"
}
*/
func PostConnections(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	return api.PostConnections(input)
}

//TODO Please modify the folowing code to adapt to your plugin
/*
PATCH /plugins/GithubSinger/connections/:connectionId
{
	"name": "GithubSinger data connection name",
	"endpoint": "GithubSinger api endpoint, i.e. https://example.com",
	"username": "username, usually should be email address",
	"password": "GithubSinger api access token"
}
*/
func PatchConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	return api.PatchConnection(input)
}

/*
DELETE /plugins/GithubSinger/connections/:connectionId
*/
func DeleteConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	return api.DeleteConnection(input)
}

/*
GET /plugins/GithubSinger/connections
*/
func ListConnections(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	return api.ListConnections(input)
}

//TODO Please modify the folowing code to adapt to your plugin
/*
GET /plugins/GithubSinger/connections/:connectionId
{
	"name": "GithubSinger data connection name",
	"endpoint": "GithubSinger api endpoint, i.e. https://merico.atlassian.net/rest",
	"username": "username, usually should be email address",
	"password": "GithubSinger api access token"
}
*/
func GetConnection(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
	return api.GetConnection(input)
}
