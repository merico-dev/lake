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

package models

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/apache/incubator-devlake/core/errors"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api/apihelperabstract"
)

// GithubAccessToken supports fetching data with multiple tokens
type GithubAccessToken struct {
	helper.AccessToken `mapstructure:",squash"`
	tokens             []string `gorm:"-" json:"-" mapstructure:"-"`
	tokenIndex         int      `gorm:"-" json:"-" mapstructure:"-"`
}

// GithubConn holds the essential information to connect to the Github API
type GithubConn struct {
	helper.RestConnection `mapstructure:",squash"`
	GithubAccessToken     `mapstructure:",squash"`
}

// PrepareApiClient splits Token to tokens for SetupAuthentication to utilize
func (conn *GithubConn) PrepareApiClient(apiClient apihelperabstract.ApiClientAbstract) errors.Error {
	conn.tokens = strings.Split(conn.Token, ",")
	return nil
}

// SetupAuthentication sets up the HTTP Request Authentication
func (gat *GithubAccessToken) SetupAuthentication(req *http.Request) errors.Error {
	// Rotates token on each request.
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", gat.tokens[gat.tokenIndex]))
	// Set next token index
	gat.tokenIndex = (gat.tokenIndex + 1) % len(gat.tokens)
	return nil
}

// GetTokensCount returns total number of tokens
func (gat *GithubAccessToken) GetTokensCount() int {
	return len(gat.tokens)
}

// GithubConnection holds GithubConn plus ID/Name for database storage
type GithubConnection struct {
	helper.BaseConnection `mapstructure:",squash"`
	GithubConn            `mapstructure:",squash"`
	EnableGraphql         bool `mapstructure:"enableGraphql" json:"enableGraphql"`
}

func (GithubConnection) TableName() string {
	return "_tool_github_connections"
}

// Using GithubUserOfToken because it requires authentication, and it is public information anyway.
type GithubUserOfToken struct {
	Login string `json:"login"`
}
