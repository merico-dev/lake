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
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

// NoopConverter no-op converter
var NoopConverter = func(b []byte) (any, errors.Error) {
	return b, nil
}

// ProcessResponse wraps output of a process
type ProcessResponse struct {
	stdout any
	stderr any
	fdOut  any
	err    errors.Error
}

// ProcessStream wraps output of a process
type ProcessStream struct {
	receiveChannel <-chan *ProcessResponse
	process        *os.Process
	cancelled      bool
}

// StreamProcessOptions options for streaming a process
type StreamProcessOptions struct {
	OnStdout func(b []byte) (any, errors.Error)
	OnStderr func(b []byte) (any, errors.Error)
	// UseFdOut if true, it'll open this fd to be used by the child process. Useful to isolate stdout and custom outputs
	UseFdOut bool
	OnFdOut  func(b []byte) (any, errors.Error)
}

// RunProcessOptions options for running a process
type RunProcessOptions struct {
	OnStdout func(b []byte)
	OnStderr func(b []byte)
	UseFdOut bool
	OnFdOut  func(b []byte)
}

type processPipes struct {
	stdout io.ReadCloser
	stderr io.ReadCloser
	fdOut  io.ReadCloser
}

func (p *processPipes) close() {
	_ = p.stderr.Close()
	_ = p.stdout.Close()
	if p.fdOut != nil {
		_ = p.fdOut.Close()
	}
}

// Receive listens to the process retrieval channel
func (p *ProcessStream) Receive() <-chan *ProcessResponse {
	return p.receiveChannel
}

// Cancel cancels the stream by sending a termination signal to the target.
func (p *ProcessStream) Cancel() errors.Error {
	err := errors.Convert(p.process.Signal(syscall.SIGTERM))
	if err != nil {
		return err
	}
	p.cancelled = true
	return nil
}

// GetStdout gets the stdout. The type depends on the conversion defined by StreamProcessOptions.OnStdout, otherwise it is []byte
func (resp *ProcessResponse) GetStdout() any {
	return resp.stdout
}

// GetStderr gets the stderr. The type depends on the conversion defined by StreamProcessOptions.OnStderr, otherwise it is []byte
func (resp *ProcessResponse) GetStderr() any {
	return resp.stderr
}

// GetFdOut gets the custom fd output. The type depends on the conversion defined by StreamProcessOptions.OnFdOut, otherwise it is []byte
func (resp *ProcessResponse) GetFdOut() any {
	return resp.fdOut
}

// GetError gets the error on the response
func (resp *ProcessResponse) GetError() errors.Error {
	return resp.err
}

// RunProcess runs the cmd and blocks until its completion. All returned results will have type []byte.
func RunProcess(cmd *exec.Cmd, opts *RunProcessOptions) (*ProcessResponse, errors.Error) {
	stream, err := StreamProcess(cmd, &StreamProcessOptions{
		OnStdout: func(b []byte) (any, errors.Error) {
			if opts.OnStdout != nil {
				opts.OnStdout(b)
			}
			return NoopConverter(b)
		},
		OnStderr: func(b []byte) (any, errors.Error) {
			if opts.OnStderr != nil {
				opts.OnStderr(b)
			}
			return NoopConverter(b)
		},
		UseFdOut: opts.UseFdOut,
		OnFdOut: func(b []byte) (any, errors.Error) {
			if opts.OnFdOut != nil {
				opts.OnFdOut(b)
			}
			return NoopConverter(b)
		},
	})
	if err != nil {
		return nil, err
	}
	var stdout []byte
	var stderr []byte
	var fdOut []byte
	for result := range stream.Receive() {
		if result.err != nil {
			err = result.err
			break
		}
		if result.stdout != nil {
			stdout = append(stdout, result.stdout.([]byte)...)
		}
		if result.stderr != nil {
			stderr = append(stderr, result.stderr.([]byte)...)
		}
		if result.fdOut != nil {
			fdOut = append(fdOut, result.fdOut.([]byte)...)
		}
	}
	return &ProcessResponse{
		stdout: stdout,
		stderr: stderr,
		fdOut:  fdOut,
		err:    err,
	}, nil
}

// StreamProcess runs the cmd and returns its output on a line-by-line basis, on a channel. The converter functor will allow you
// to convert the incoming raw to your custom data type T. This is a nonblocking function.
func StreamProcess(cmd *exec.Cmd, opts *StreamProcessOptions) (*ProcessStream, errors.Error) {
	if opts == nil {
		opts = &StreamProcessOptions{}
	}
	cmd.Env = append(cmd.Env, os.Environ()...)
	pipes, err := getPipes(cmd, opts)
	if err != nil {
		return nil, err
	}
	if err = errors.Convert(cmd.Start()); err != nil {
		return nil, err
	}
	receiveStream := make(chan *ProcessResponse, 32)
	wg := &sync.WaitGroup{}
	stdScanner := scanOutputPipe(pipes.stdout, wg, opts.OnStdout, func(result any) *ProcessResponse {
		return &ProcessResponse{stdout: result}
	}, receiveStream)
	errScanner, remoteErrorMsg := scanErrorPipe(pipes.stderr, receiveStream)
	fdOutScanner := scanOutputPipe(pipes.fdOut, wg, opts.OnFdOut, func(result any) *ProcessResponse {
		return &ProcessResponse{fdOut: result}
	}, receiveStream)
	wg.Add(2)
	if pipes.fdOut != nil {
		wg.Add(1)
	}
	go stdScanner()
	go errScanner()
	if pipes.fdOut != nil {
		go fdOutScanner()
	}
	processStream := &ProcessStream{
		process:        cmd.Process,
		receiveChannel: receiveStream,
	}
	go func() {
		defer pipes.close()
		if err = errors.Convert(cmd.Wait()); err != nil {
			if !processStream.cancelled {
				receiveStream <- &ProcessResponse{err: errors.Default.Wrap(err, fmt.Sprintf("remote error response:\n%s", remoteErrorMsg))}
			}
		}
		wg.Done()
	}()
	go func() {
		defer close(receiveStream)
		wg.Wait()
	}()
	return processStream, nil
}

func getPipes(cmd *exec.Cmd, opts *StreamProcessOptions) (*processPipes, errors.Error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Convert(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, errors.Convert(err)
	}
	var fdOut *os.File
	if opts.UseFdOut {
		fdReader, fdOutWriter, err := os.Pipe()
		if err != nil {
			return nil, errors.Convert(err)
		}
		cmd.ExtraFiles = []*os.File{fdOutWriter}
		fdOut = fdReader
	}
	return &processPipes{
		stdout: stdout,
		stderr: stderr,
		fdOut:  fdOut,
	}, nil
}

func scanOutputPipe(pipe io.ReadCloser, wg *sync.WaitGroup, onReceive func([]byte) (any, errors.Error),
	responseCreator func(any) *ProcessResponse, outboundChannel chan<- *ProcessResponse) func() {
	return func() {
		scanner := bufio.NewScanner(pipe)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			src := scanner.Bytes()
			data := make([]byte, len(src))
			copy(data, src)
			if result, err := onReceive(data); err != nil {
				outboundChannel <- &ProcessResponse{err: err}
			} else {
				outboundChannel <- responseCreator(result)
			}
		}
		wg.Done()
	}
}

func scanErrorPipe(pipe io.ReadCloser, outboundChannel chan<- *ProcessResponse) (func(), *strings.Builder) {
	remoteErrorMsg := &strings.Builder{}
	return func() {
		scanner := bufio.NewScanner(pipe)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			src := scanner.Bytes()
			data := make([]byte, len(src))
			copy(data, src)
			outboundChannel <- &ProcessResponse{stderr: data}
			_, _ = remoteErrorMsg.Write(src)
			_, _ = remoteErrorMsg.WriteString("\n")
		}
	}, remoteErrorMsg
}

// CreateCmd wraps the args in "sh -c" for shell-level execution
func CreateCmd(args ...string) *exec.Cmd {
	if len(args) < 1 {
		panic("no cmd given")
	}
	cmd := "sh"
	cmdArgs := []string{"-c"}
	cmdBuilder := &strings.Builder{}
	for _, elem := range args {
		if elem != "" {
			_, _ = cmdBuilder.WriteString(elem)
			_, _ = cmdBuilder.WriteString(" ")
		}
	}
	cmdArgs = append(cmdArgs, cmdBuilder.String())
	return exec.Command(cmd, cmdArgs...)
}
