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
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/apache/incubator-devlake/core/config"
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/dgrijalva/jwt-go"
)

type AwsCognitoProvider struct {
	jwks         Jwks
	logger       log.Logger
	client       *cognitoidentityprovider.CognitoIdentityProvider
	clientId     *string
	expectClaims jwt.MapClaims
}

func NewCognitoProvider(basicRes context.BasicRes) *AwsCognitoProvider {
	// Get configuration
	v := config.GetConfig()
	// TODO: verify the configuration
	// Create an AWS session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(v.GetString("AWS_AUTH_REGION")),
	}))
	// Create a Cognito Identity Provider client
	client := cognitoidentityprovider.New(sess)
	cgt := &AwsCognitoProvider{
		client:   client,
		clientId: aws.String(v.GetString("AWS_AUTH_USER_POOL_WEB_CLIENT_ID")),
		logger:   basicRes.GetLogger().Nested("cognito"),
	}
	// Fetch the JWKS from the Cognito User Pool
	jwksURL := fmt.Sprintf(
		"https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json",
		v.GetString("AWS_AUTH_REGION"),
		v.GetString("AWS_AUTH_USER_POOL_ID"),
	)
	err := cgt.fetchJWKS(jwksURL)
	if err != nil {
		panic(err)
	}
	// Optional expect claims
	expect_claims := strings.TrimSpace(v.GetString("AWS_AUTH_EXPECT_CLAIMS"))
	if expect_claims != "" {
		e := json.Unmarshal([]byte(expect_claims), &cgt.expectClaims)
		if e != nil {
			panic(e)
		}
	}
	return cgt
}

func (cgt *AwsCognitoProvider) fetchJWKS(jwksURL string) errors.Error {
	// Get the JWKS from the URL
	resp, err := http.Get(jwksURL)
	if err != nil {
		return errors.Default.Wrap(err, "Failed to fetch JWKS")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Default.Wrap(err, "Failed to read JWKS")
	}
	// Unmarshal the response into a Jwks struct
	err = json.Unmarshal(body, &cgt.jwks)
	if err != nil {
		return errors.Default.Wrap(err, "Failed to unmarshall JWKS")
	}
	return nil
}

func (cgt *AwsCognitoProvider) SignIn(loginReq *LoginRequest) (*LoginResponse, errors.Error) {
	// Create the input for InitiateAuth
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		ClientId: cgt.clientId,
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(loginReq.Username),
			"PASSWORD": aws.String(loginReq.Password),
		},
	}

	// Call Cognito to get auth tokens
	response, err := cgt.client.InitiateAuth(input)
	if err != nil {
		return nil, errors.BadInput.New(err.Error())
	}

	loginRes := &LoginResponse{
		ChallengeName:       response.ChallengeName,
		ChallengeParameters: response.ChallengeParameters,
		Session:             response.Session,
	}
	if response.AuthenticationResult != nil {
		loginRes.AuthenticationResult = &AuthenticationResult{
			AccessToken:  response.AuthenticationResult.AccessToken,
			ExpiresIn:    response.AuthenticationResult.ExpiresIn,
			IdToken:      response.AuthenticationResult.IdToken,
			RefreshToken: response.AuthenticationResult.RefreshToken,
			TokenType:    response.AuthenticationResult.TokenType,
		}
	}

	return loginRes, nil
}

func (cgt *AwsCognitoProvider) CheckAuth(tokenString string) (*jwt.Token, errors.Error) {
	// Parse the JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.Unauthorized.New(fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]))
		}

		// Get the key ID from the header
		kid := token.Header["kid"].(string)

		// Look for the key that matches the kid
		for _, key := range cgt.jwks.Keys {
			if key.Kid == kid {
				// Construct the RSA public key
				n := pemHeader(key.N)
				e := pemHeader(key.E)
				parsedKey := &rsa.PublicKey{
					N: new(big.Int).SetBytes(n),
					E: int(new(big.Int).SetBytes(e).Int64()),
				}
				return parsedKey, nil
			}
		}

		return nil, fmt.Errorf("Public key not found")
	})

	// Check if the token is invalid
	if err != nil || !token.Valid {
		cgt.logger.Error(err, "Invalid token")
		return nil, errors.Unauthorized.New("Invalid token")
	}

	// verify claims
	if len(cgt.expectClaims) > 0 {
		if actualClaims, ok := token.Claims.(jwt.MapClaims); ok {
			for key, expected := range cgt.expectClaims {
				if expected != actualClaims[key] {
					return nil, errors.Unauthorized.New("Invalid token")
				}
			}
		} else {
			return nil, errors.Unauthorized.New("Invalid token")
		}
	}

	return token, nil
}

func pemHeader(encodedKey string) []byte {
	// Decode the base64 encoded key
	key, err := base64.RawURLEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil
	}
	return key
}

// Jwks represents the JSON web key set
type Jwks struct {
	Keys []struct {
		Kid string `json:"kid"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"keys"`
}

func (cgt *AwsCognitoProvider) NewPassword(newPasswordReq *NewPasswordRequest) (*LoginResponse, errors.Error) {
	input := &cognitoidentityprovider.RespondToAuthChallengeInput{
		ChallengeName: aws.String("NEW_PASSWORD_REQUIRED"),
		ChallengeResponses: map[string]*string{
			"USERNAME":     aws.String(newPasswordReq.Username),
			"NEW_PASSWORD": aws.String(newPasswordReq.NewPassword),
		},
		Session:  aws.String(newPasswordReq.Session),
		ClientId: cgt.clientId,
	}
	response, err := cgt.client.RespondToAuthChallenge(input)
	if err != nil {
		return nil, errors.BadInput.Wrap(err, "Error setting up new password: "+err.Error())
	}
	// yes , it is identical to the login response, and yet they are 2 different structs
	loginRes := &LoginResponse{
		ChallengeName:       response.ChallengeName,
		ChallengeParameters: response.ChallengeParameters,
		Session:             response.Session,
	}
	if response.AuthenticationResult != nil {
		loginRes.AuthenticationResult = &AuthenticationResult{
			AccessToken:  response.AuthenticationResult.AccessToken,
			ExpiresIn:    response.AuthenticationResult.ExpiresIn,
			IdToken:      response.AuthenticationResult.IdToken,
			RefreshToken: response.AuthenticationResult.RefreshToken,
			TokenType:    response.AuthenticationResult.TokenType,
		}
	}
	return loginRes, nil
}

// func (cgt *AwsCognitorProvider) ChangePassword(ctx *gin.Context, oldPassword, newPassword string) errors.Error {
// 	token := ctx.GetString(("token"))
// 	if token == "" {
// 		return errors.Unauthorized.New("Token is missing")
// 	}
// 	input := &cognitoidentityprovider.ChangePasswordInput{
// 		AccessToken:      &token,
// 		PreviousPassword: &oldPassword,
// 		ProposedPassword: &newPassword,
// 	}
// 	_, err := cgt.client.ChangePassword(input)
// 	if err != nil {
// 		return errors.BadInput.Wrap(err, "Error changing password")
// 	}
// 	return nil
// }
