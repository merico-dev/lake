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

package auth

import (
	"strings"

	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/server/api/shared"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// data structures

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AuthenticationResult *AuthenticationResult `json:"authenticationResult"`
	ChallengeName        *string               `json:"challengeName"`
	ChallengeParameters  map[string]*string    `json:"challengeParameters"`
	Session              *string               `json:"session"`
}

type AuthenticationResult struct {
	AccessToken  *string `json:"accessToken" type:"string" sensitive:"true"`
	ExpiresIn    *int64  `json:"expiresIn" type:"integer"`
	IdToken      *string `json:"idToken" type:"string" sensitive:"true"`
	RefreshToken *string `json:"refreshToken" type:"string" sensitive:"true"`
	TokenType    *string `json:"tokenType" type:"string"`
}

type NewPasswordRequest struct {
	Username    string `json:"username"`
	NewPassword string `json:"newPassword"`
	Session     string `json:"session"`
}

// auth provider interface
type AuthProvider interface {
	SignIn(*LoginRequest) (*LoginResponse, errors.Error)
	NewPassword(*NewPasswordRequest) (*LoginResponse, errors.Error)
	// ChangePassword(ctx *gin.Context, oldPassword, newPassword string) errors.Error
	CheckAuth(token string) (*jwt.Token, errors.Error)
}

var Provider AuthProvider

// initialize auth provider
func InitProvider(basicRes context.BasicRes) {
	v := basicRes.GetConfigReader()
	awsCognitoEnabled := v.GetBool("AWS_ENABLE_COGNITO")
	if awsCognitoEnabled {
		Provider = NewCognitoProvider(basicRes)
	}
}

func Middleware(ctx *gin.Context) {
	if Provider == nil {
		return
	}
	// Get the Auth header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		shared.ApiOutputAbort(ctx, errors.Unauthorized.New("Authorization header is missing"))
		return
	}

	// Split the header into "Bearer" and the actual token
	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 {
		shared.ApiOutputAbort(ctx, errors.Unauthorized.New("Invalid Authorization header"))
		return
	}
	token, err := Provider.CheckAuth(bearerToken[1])
	if err != nil {
		shared.ApiOutputAbort(ctx, err)
		return
	}

	ctx.Set("token", token)
}

func Enabled() bool {
	return Provider != nil
}
