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
	"github.com/mitchellh/mapstructure"
)

type (
	CallResult struct {
		results []map[string]any
		err     errors.Error
	}
	StreamResult = CallResult
	MethodStream struct {
		// send (not supported for now)
		outbound chan<- any
		// receive
		inbound <-chan *StreamResult
	}
)

func NewCallResult(results []map[string]any, err errors.Error) *CallResult {
	return &CallResult{
		results: results,
		err:     err,
	}
}

func (m *CallResult) Get(targets ...any) errors.Error {
	if m.err != nil {
		return m.err
	}
	if len(targets) != len(m.results) {
		// if everything came back as nil, consider it good
		for _, result := range m.results {
			if result != nil {
				return errors.Default.New("unequal results and targets length")
			}
		}
		return nil
	}
	for i, target := range targets {
		err := errors.Convert(mapstructure.Decode(m.results[i], target))
		if err != nil {
			return err
		}
	}
	return nil
}

func NewStreamResult(results []map[string]any, err errors.Error) *StreamResult {
	return &StreamResult{
		results: results,
		err:     err,
	}
}

func (m *MethodStream) Receive() <-chan *StreamResult {
	return m.inbound
}
