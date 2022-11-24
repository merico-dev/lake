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

package bridge

import (
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/services/remote/models"
)

type (
	Bridge struct {
		invoker Invoker
	}
	Invoker interface {
		Call(methodName string, ctx core.ExecContext, args ...any) *CallResult
		Stream(methodName string, ctx core.ExecContext, args ...any) *MethodStream
	}
)

func NewBridge(invoker Invoker) *Bridge {
	return &Bridge{invoker: invoker}
}

func (b *Bridge) RemoteAPIHandler(ep *models.Endpoint) core.ApiResourceHandler {
	endpoint := *ep
	return func(input *core.ApiResourceInput) (*core.ApiResourceOutput, errors.Error) {
		output := core.ApiResourceOutput{}
		err := b.invoker.Call(endpoint.Handler, DefaultContext, input).Get(&output)
		if err != nil {
			return nil, err
		}
		return &output, err
	}
}

func (b *Bridge) RemoteSubtaskEntrypointHandler(subtaskMeta *models.SubtaskMeta) core.SubTaskEntryPoint {
	subtask := *subtaskMeta
	args := make([]any, len(subtask.Arguments))
	for i, arg := range subtask.Arguments {
		args[i] = arg
	}
	return func(ctx core.SubTaskContext) errors.Error {
		stream := b.invoker.Stream(subtask.EntryPointName, NewChildRemoteContext(ctx), args...)
		for recv := range stream.Receive() {
			if recv.err != nil {
				return recv.err
			}
			progress := RemoteProgress{}
			err := recv.Get(&progress)
			if err != nil {
				return err
			}
			if progress.Current != 0 {
				ctx.SetProgress(progress.Current, progress.Total)
			} else if progress.Increment != 0 {
				ctx.IncProgress(progress.Increment)
			}
		}
		return nil
	}
}
