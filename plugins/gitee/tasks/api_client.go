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

package tasks

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/utils"
)

func NewGiteeApiClient(taskCtx core.TaskContext) (*helper.ApiAsyncClient, error) {
	endpoint := taskCtx.GetConfig("GITEE_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("endpint is required")
	}
	userRateLimit, err := utils.StrToIntOr(taskCtx.GetConfig("GITEE_API_REQUESTS_PER_HOUR"), 0)
	if err != nil {
		return nil, err
	}
	auth := taskCtx.GetConfig("GITEE_AUTH")
	if auth == "" {
		return nil, fmt.Errorf("GITEE_AUTH is required")
	}
	proxy := taskCtx.GetConfig("GITEE_PROXY")

	headers := map[string]string{}

	apiClient, err := helper.NewApiClient(endpoint, headers, 0, proxy, taskCtx.GetContext())
	if err != nil {
		return nil, err
	}
	apiClient.SetAfterFunction(func(res *http.Response) error {
		if res.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("authentication failed, please check your Basic Auth Token")
		}
		return nil
	})

	rateLimiter := &helper.ApiRateLimitCalculator{
		UserRateLimitPerHour: userRateLimit,
		DynamicRateLimit: func(res *http.Response) (int, time.Duration, error) {
			rateLimitHeader := res.Header.Get("RateLimit-Limit")
			if rateLimitHeader == "" {
				// use default
				return 0, 0, nil
			}
			rateLimit, err := strconv.Atoi(rateLimitHeader)
			if err != nil {
				return 0, 0, fmt.Errorf("failed to parse RateLimit-Limit header: %w", err)
			}
			// seems like gitlab rate limit is on minute basis
			return rateLimit, 1 * time.Minute, nil
		},
	}
	asyncApiClient, err := helper.CreateAsyncApiClient(
		taskCtx,
		apiClient,
		rateLimiter,
	)
	if err != nil {
		return nil, err
	}
	return asyncApiClient, nil
}
