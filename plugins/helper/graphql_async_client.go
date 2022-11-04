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
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/utils"
	"github.com/merico-dev/graphql"
)

// GraphqlAsyncClient send graphql one by one
type GraphqlAsyncClient struct {
	ctx       context.Context
	cancel    context.CancelFunc
	client    *graphql.Client
	logger    core.Logger
	mu        sync.Mutex
	waitGroup sync.WaitGroup

	maxRetry         int
	waitBeforeRetry  time.Duration
	rateExhaustCond  *sync.Cond
	rateRemaining    int
	getRateRemaining func(context.Context, *graphql.Client, core.Logger) (rateRemaining int, resetAt *time.Time, err errors.Error)
	getRateCost      func(q interface{}) int
}

// CreateAsyncGraphqlClient creates a new GraphqlAsyncClient
func CreateAsyncGraphqlClient(
	taskCtx core.TaskContext,
	graphqlClient *graphql.Client,
	logger core.Logger,
	getRateRemaining func(context.Context, *graphql.Client, core.Logger) (rateRemaining int, resetAt *time.Time, err errors.Error),
) (*GraphqlAsyncClient, errors.Error) {
	ctxWithCancel, cancel := context.WithCancel(taskCtx.GetContext())
	graphqlAsyncClient := &GraphqlAsyncClient{
		ctx:              ctxWithCancel,
		cancel:           cancel,
		client:           graphqlClient,
		logger:           logger,
		rateExhaustCond:  sync.NewCond(&sync.Mutex{}),
		rateRemaining:    0,
		getRateRemaining: getRateRemaining,
	}

	if getRateRemaining != nil {
		rateRemaining, resetAt, err := getRateRemaining(taskCtx.GetContext(), graphqlClient, logger)
		if err != nil {
			panic(err)
		}
		graphqlAsyncClient.updateRateRemaining(rateRemaining, resetAt)
	}

	// load retry/timeout from configuration
	// use API_RETRY as max retry time
	// use API_TIMEOUT as retry before wait seconds to confirm the prev request finish
	timeout := 30 * time.Second
	retry, err := utils.StrToIntOr(taskCtx.GetConfig("API_RETRY"), 3)
	if err != nil {
		return nil, errors.BadInput.Wrap(err, "failed to parse API_RETRY")
	}
	timeoutConf := taskCtx.GetConfig("API_TIMEOUT")
	if timeoutConf != "" {
		// override timeout value if API_TIMEOUT is provided
		timeout, err = errors.Convert01(time.ParseDuration(timeoutConf))
		if err != nil {
			return nil, errors.BadInput.Wrap(err, "failed to parse API_TIMEOUT")
		}
	}
	graphqlAsyncClient.SetMaxRetry(retry, timeout)

	return graphqlAsyncClient, nil
}

// GetMaxRetry returns the maximum retry attempts for a request
func (apiClient *GraphqlAsyncClient) GetMaxRetry() (int, time.Duration) {
	return apiClient.maxRetry, apiClient.waitBeforeRetry
}

// SetMaxRetry sets the maximum retry attempts for a request
func (apiClient *GraphqlAsyncClient) SetMaxRetry(
	maxRetry int,
	waitBeforeRetry time.Duration,
) {
	apiClient.maxRetry = maxRetry
	apiClient.waitBeforeRetry = waitBeforeRetry
}

// updateRateRemaining call getRateRemaining to update rateRemaining periodically
func (apiClient *GraphqlAsyncClient) updateRateRemaining(rateRemaining int, resetAt *time.Time) {
	apiClient.rateRemaining = rateRemaining
	if rateRemaining > 0 {
		apiClient.rateExhaustCond.Signal()
	}
	go func() {
		nextDuring := 3 * time.Minute
		if resetAt != nil && resetAt.After(time.Now()) {
			nextDuring = time.Until(*resetAt)
		}
		select {
		case <-apiClient.ctx.Done():
			return
		case <-time.After(nextDuring):
			newRateRemaining, newResetAt, err := apiClient.getRateRemaining(apiClient.ctx, apiClient.client, apiClient.logger)
			if err != nil {
				panic(err)
			}
			apiClient.updateRateRemaining(newRateRemaining, newResetAt)
		}
	}()
}

// SetGetRateCost to calculate how many rate cost
// if not set, all query just cost 1
func (apiClient *GraphqlAsyncClient) SetGetRateCost(getRateCost func(q interface{}) int) {
	apiClient.getRateCost = getRateCost
}

// Query send a graphql request when get lock
// []graphql.DataError are the errors returned in response body
// errors.Error is other error
func (apiClient *GraphqlAsyncClient) Query(q interface{}, variables map[string]interface{}) ([]graphql.DataError, errors.Error) {
	apiClient.waitGroup.Add(1)
	defer apiClient.waitGroup.Done()
	apiClient.mu.Lock()
	defer apiClient.mu.Unlock()

	apiClient.rateExhaustCond.L.Lock()
	defer apiClient.rateExhaustCond.L.Unlock()
	for apiClient.rateRemaining <= 0 {
		apiClient.logger.Info(`rate limit remaining exhausted, waiting for next period.`)
		apiClient.rateExhaustCond.Wait()
	}

	retryTime := 0
	var err error
	//  if it needs retry, check and retry
	for retryTime < apiClient.maxRetry {
		select {
		case <-apiClient.ctx.Done():
			return nil, nil
		default:
			var dataErrors []graphql.DataError
			dataErrors, err := apiClient.client.Query(apiClient.ctx, q, variables)
			if err != nil {
				apiClient.logger.Warn(err, "retry #%d graphql calling after %ds", retryTime, apiClient.waitBeforeRetry/time.Second)
				retryTime++
				<-time.After(apiClient.waitBeforeRetry)
				continue
			}
			if dataErrors != nil {
				return dataErrors, nil
			}
			cost := 1
			if apiClient.getRateCost != nil {
				cost = apiClient.getRateCost(q)
			}
			apiClient.rateRemaining -= cost
			apiClient.logger.Debug(`query cost %d in %v`, cost, variables)
			return nil, nil
		}
	}
	return nil, errors.Default.Wrap(err, fmt.Sprintf("got error when querying GraphQL (from the %dth retry)", retryTime))
}

// NextTick to return the NextTick of scheduler
func (apiClient *GraphqlAsyncClient) NextTick(task func() errors.Error, taskErrorChecker func(err errors.Error)) {
	// to make sure task will be enqueued
	apiClient.waitGroup.Add(1)
	go func() {
		select {
		case <-apiClient.ctx.Done():
			return
		default:
			go func() {
				// if set waitGroup done here, a serial of goruntine will block until son goruntine finish.
				// But if done out of this go func, so task will run after waitGroup finish
				// I have no idea about this now...
				defer apiClient.waitGroup.Done()
			}()
		}
	}()
}

// Wait blocks until all async requests were done
func (apiClient *GraphqlAsyncClient) Wait() {
	apiClient.waitGroup.Wait()
}

// Release will release the ApiAsyncClient with scheduler
func (apiClient *GraphqlAsyncClient) Release() {
	apiClient.cancel()
}
