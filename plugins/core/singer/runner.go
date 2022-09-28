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

package singer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var propertiesCache = map[string]any{}

type TapProperties[Schema any] struct {
	Streams []*Stream[Schema] `json:"streams"`
}

type Tap[Schema any] struct {
	cmd            string
	tempLocation   string
	propertiesFile *fileData[TapProperties[Schema]]
	stateFile      *fileData[[]byte]
	configFile     *fileData[[]byte]
	cfg            *Config
}

type fileData[Content any] struct {
	path    string
	content Content
}

func NewTap[Schema any](cfg *Config) *Tap[Schema] {
	tempDir, err := os.MkdirTemp("", "singer"+"_*")
	if err != nil {
		panic(err)
	}
	_ = tempDir
	return &Tap[Schema]{
		cmd:          cfg.Cmd,
		tempLocation: tempDir,
		cfg:          cfg,
	}
}

func (t *Tap[Schema]) WriteProperties() {
	file, err := os.OpenFile(t.propertiesFile.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	b, err := json.Marshal(t.propertiesFile.content)
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(file)
	if _, err = writer.Write(b); err != nil {
		panic(err)
	}
}

func (t *Tap[Schema]) WriteConfig() {
	b, err := json.Marshal(t.cfg.Mappings)
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile(filepath.Join(t.tempLocation, "config.json"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	_, err = file.Write(b)
	if err != nil {
		panic(err)
	}
	t.configFile = &fileData[[]byte]{
		path:    file.Name(),
		content: b,
	}
}

func (t *Tap[Schema]) WriteState(state interface{}) {
	b, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile(filepath.Join(t.tempLocation, "state.json"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	_, err = file.Write(b)
	if err != nil {
		panic(err)
	}
	t.stateFile = &fileData[[]byte]{
		path:    file.Name(),
		content: b,
	}
}

func (t *Tap[Schema]) discoverProperties() *TapProperties[Schema] {
	args := []string{"--config", t.configFile.path, "--discover"}
	cmd := exec.Command(t.cmd, args...)
	response, err := RunProcess(cmd)
	if err != nil {
		panic(err)
	}
	if response.Err != nil {
		panic(response.Err)
	}
	var properties TapProperties[Schema]
	if err = json.Unmarshal(response.Data, &properties); err != nil {
		panic(err)
	}
	return &properties
}

func (t *Tap[Schema]) DiscoverProperties() {
	propsRaw, ok := propertiesCache[t.cfg.TapType]
	var props *TapProperties[Schema]
	if !ok {
		props = t.discoverProperties()
		propertiesCache[t.cfg.TapType] = props
	} else {
		props = propsRaw.(*TapProperties[Schema])
	}
	t.propertiesFile = &fileData[TapProperties[Schema]]{
		path:    filepath.Join(t.tempLocation, "properties.json"),
		content: *props,
	}
}

func (t *Tap[Schema]) SetProperties(f func(stream *Stream[Schema]) bool) {
	err := t.modifyStreams(f)
	if err != nil {
		panic(err)
	}
}

func (t *Tap[Schema]) Run() <-chan *ProcessResponse[TapResponse] {
	t.WriteProperties()
	args := []string{"--config", t.configFile.path, "--properties", t.propertiesFile.path}
	if t.stateFile != nil {
		args = append(args, []string{"--state", t.stateFile.path}...)
	}
	cmd := exec.Command(t.cmd, args...)
	stream, err := StreamProcess[TapResponse](cmd, func(b []byte) (TapResponse, error) {
		result := TapResponse{}
		if err := json.Unmarshal(b, &result); err != nil {
			return result, errors.Default.WrapRaw(err)
		}
		return result, nil
	})
	if err != nil {
		panic(err)
	}
	return stream
}

type ProcessResponse[T any] struct {
	Data T
	Err  error
}

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
			remoteErrorMsg.Write(scanner.Bytes())
			remoteErrorMsg.WriteString("\n")
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
			remoteErrorMsg.Write(scanner.Bytes())
			remoteErrorMsg.WriteString("\n")
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

func (t *Tap[Schema]) modifyStreams(modifier func(stream *Stream[Schema]) bool) error {
	var err error
	properties := t.propertiesFile.content
	filteredStreams := []*Stream[Schema]{}
	for i := 0; i < len(properties.Streams); i++ {
		stream := properties.Streams[i]
		if modifier(stream) {
			filteredStreams = append(filteredStreams, stream)
		}
	}
	properties.Streams = filteredStreams
	var encodedJson []byte
	if encodedJson, err = json.Marshal(&properties); err != nil {
		return err
	}
	f, err := os.OpenFile(t.propertiesFile.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(f)
	_, err = writer.Write(encodedJson)
	if err != nil {
		return err
	}
	_ = writer.Flush()
	return nil
}
