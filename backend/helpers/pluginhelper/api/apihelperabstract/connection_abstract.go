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

package apihelperabstract

import (
	"net/http"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/go-playground/validator/v10"
)

// ApiConnection represents a API Connection
type ApiConnection interface {
	GetEndpoint() string
	GetProxy() string
	GetRateLimitPerHour() int
}

// ApiAuthenticator is to be implemented by a Concreate Connection if Authorization is required
type ApiAuthenticator interface {
	// SetupAuthentication is a hook function for connection to set up authentication for the HTTP request
	// before sending it to the server
	SetupAuthentication(request *http.Request) errors.Error
}

// ConnectionValidator represents the API Connection would validate its fields with customized logic
type ConnectionValidator interface {
	ValidateConnection(connection interface{}, valdator *validator.Validate) errors.Error
}

// PrepareApiClient is to be implemented by the concrete Connection which requires
// preparation for the ApiClient created by NewApiClientFromConnection, i.e. fetch token for future requests
type PrepareApiClient interface {
	PrepareApiClient(apiClient ApiClientAbstract) errors.Error
}

// MultiAuth
const (
	AUTH_METHOD_BASIC  = "BasicAuth"
	AUTH_METHOD_TOKEN  = "AccessToken"
	AUTH_METHOD_APPKEY = "AppKey"
)

var ALL_AUTH = map[string]bool{
	AUTH_METHOD_BASIC:  true,
	AUTH_METHOD_TOKEN:  true,
	AUTH_METHOD_APPKEY: true,
}

// MultiAuthenticator represents the API Connection supports multiple authorization methods
type MultiAuthenticator interface {
	GetAuthMethod() string
}

// BasicAuthenticator represents HTTP Basic Authentication
type BasicAuthenticator interface {
	GetBasicAuthenticator() ApiAuthenticator
}

// AccessTokenAuthenticator represents HTTP Bearer Authentication with Access Token
type AccessTokenAuthenticator interface {
	GetAccessTokenAuthenticator() ApiAuthenticator
}

// AppKeyAuthenticator represents the API Key and Secret authentication mechanism
type AppKeyAuthenticator interface {
	GetAppKeyAuthenticator() ApiAuthenticator
}
