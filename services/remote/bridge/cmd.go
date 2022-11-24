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
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/utils"
	"os/exec"
)

type CmdInvoker struct {
	resolveCmd  func(methodName string, args ...string) (string, []string)
	cancelled   bool
	workingPath string
}

func NewCmdInvoker(workingPath string, resolveCmd func(methodName string, args ...string) (string, []string)) *CmdInvoker {
	return &CmdInvoker{
		resolveCmd:  resolveCmd,
		workingPath: workingPath,
	}
}

func (c *CmdInvoker) Call(methodName string, ctx core.ExecContext, args ...any) *CallResult {
	allArgs := []any{ctx}
	allArgs = append(allArgs, args...)
	serializedArgs, err := serialize(allArgs...)
	if err != nil {
		return &CallResult{
			err: err,
		}
	}
	executable, inputArgs := c.resolveCmd(methodName, serializedArgs...)
	cmdCtx := DefaultContext.GetContext()
	cmd := exec.CommandContext(cmdCtx, executable, inputArgs...)
	if c.workingPath != "" {
		cmd.Dir = c.workingPath
	}
	response, err := utils.RunProcess(cmd, &utils.RunProcessOptions{
		OnStdout: func(b []byte) {
			msg := string(b)
			c.logRemoteMessage(ctx.GetLogger(), msg)
		},
		OnStderr: func(b []byte) {
			msg := string(b)
			c.logRemoteError(ctx.GetLogger(), msg)
		},
		UseFdOut: true,
	})
	if err != nil {
		return NewCallResult(nil, err)
	}
	err = response.GetError()
	if err != nil {
		return &CallResult{
			err: errors.Default.Wrap(err, fmt.Sprintf("failed to invoke remote function \"%s\"", methodName)),
		}
	}
	results, err := deserialize(response.GetFdOut().([]byte))
	return NewCallResult(results, errors.Convert(err))
}

func (c *CmdInvoker) Stream(methodName string, ctx core.ExecContext, args ...any) *MethodStream {
	recvChannel := make(chan *StreamResult)
	stream := &MethodStream{
		outbound: nil,
		inbound:  recvChannel,
	}
	allArgs := []any{ctx}
	allArgs = append(allArgs, args...)
	serializedArgs, err := serialize(allArgs...)
	if err != nil {
		recvChannel <- NewStreamResult(nil, err)
		return stream
	}
	executable, inputArgs := c.resolveCmd(methodName, serializedArgs...)
	cmdCtx := DefaultContext.GetContext() // grabbing context off of ctx kills the cmd after a couple of seconds... why?
	cmd := exec.CommandContext(cmdCtx, executable, inputArgs...)
	if c.workingPath != "" {
		cmd.Dir = c.workingPath
	}
	processHandle, err := utils.StreamProcess(cmd, &utils.StreamProcessOptions{
		OnStdout: func(b []byte) (any, errors.Error) {
			msg := string(b)
			c.logRemoteMessage(ctx.GetLogger(), msg)
			return b, nil
		},
		OnStderr: func(b []byte) (any, errors.Error) {
			msg := string(b)
			c.logRemoteError(ctx.GetLogger(), msg)
			return b, nil
		},
		UseFdOut: true,
		OnFdOut:  utils.NoopConverter,
	})
	if err != nil {
		recvChannel <- NewStreamResult(nil, err)
		return stream
	}
	go func() {
		defer close(recvChannel)
		for msg := range processHandle.Receive() {
			if err = msg.GetError(); err != nil {
				recvChannel <- NewStreamResult(nil, err)
			}
			if !c.cancelled {
				select {
				case <-ctx.GetContext().Done():
					err = processHandle.Cancel()
					if err != nil {
						recvChannel <- NewStreamResult(nil, errors.Default.Wrap(err, "error cancelling python target"))
						return
					}
					c.cancelled = true
					// continue until the stream gets closed by the child
				default:
				}
			}
			response := msg.GetFdOut()
			if response != nil {
				results, err := deserialize(response.([]byte))
				if err != nil {
					recvChannel <- NewStreamResult(nil, err)
				} else {
					recvChannel <- NewStreamResult(results, nil)
				}
			}
		}
	}()
	return stream
}

func (c *CmdInvoker) logRemoteMessage(logger core.Logger, msg string) {
	logger.Info(msg)
}

func (c *CmdInvoker) logRemoteError(logger core.Logger, msg string) {
	logger.Error(nil, msg)
}

var _ Invoker = (*CmdInvoker)(nil)
