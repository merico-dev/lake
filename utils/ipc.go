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

package utils

import (
	"bufio"
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// ProcessResponse wraps output of a process
type ProcessResponse[T any] struct {
	Data T
	Err  error
}

// RunProcess runs the cmd and returns its raw standard output. This is a blocking function.
func RunProcess(cmd *exec.Cmd) (*ProcessResponse[[]byte], error) {
	cmd.Env = append(cmd.Env, os.Environ()...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	remoteErrorMsg := &strings.Builder{}
	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			_, _ = remoteErrorMsg.Write(scanner.Bytes())
			_, _ = remoteErrorMsg.WriteString("\n")
		}
	}()
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Default.Wrap(err, fmt.Sprintf("remote error message:\n%s", remoteErrorMsg.String()))
	}
	return &ProcessResponse[[]byte]{
		Data: output,
	}, nil
}

// StreamProcess runs the cmd and returns its standard output on a line-by-line basis, on a channel. The converter functor will allow you
// to convert the incoming raw to your custom data type T. This is a nonblocking function.
func StreamProcess[T any](cmd *exec.Cmd, converter func(b []byte) (T, error)) (<-chan *ProcessResponse[T], error) {
	cmd.Env = append(cmd.Env, os.Environ()...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	stream := make(chan *ProcessResponse[T], 32)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			src := scanner.Bytes()
			data := make([]byte, len(src))
			copy(data, src)
			if result, err := converter(data); err != nil {
				stream <- &ProcessResponse[T]{Err: err}
			} else {
				stream <- &ProcessResponse[T]{Data: result}
			}
		}
		wg.Done()
	}()
	remoteErrorMsg := &strings.Builder{}
	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			_, _ = remoteErrorMsg.Write(scanner.Bytes())
			_, _ = remoteErrorMsg.WriteString("\n")
		}
	}()
	go func() {
		if err = cmd.Wait(); err != nil {
			stream <- &ProcessResponse[T]{Err: errors.Default.Wrap(err, fmt.Sprintf("remote error response:\n%s", remoteErrorMsg))}
		}
		wg.Done()
	}()
	go func() {
		defer close(stream)
		wg.Wait()
	}()
	return stream, nil
}
