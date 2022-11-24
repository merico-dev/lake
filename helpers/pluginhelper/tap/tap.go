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

package tap

import (
	"context"
	"encoding/json"
	"github.com/apache/incubator-devlake/errors"
)

// Tap the abstract interface for Taps. Consumer code should not use concrete implementations directly.
type Tap[Stream any] interface {
	// Run runs the tap and returns a stream of results. Expected to be called after all the other Setters. The Stdout of the response should
	// be of type Output[json.RawMessage]
	Run(ctx context.Context) (<-chan *Response, errors.Error)
	// GetName the name of this tap
	GetName() string
	// SetProperties Sets the properties of the tap and allows you to modify the properties at runtime.
	// Returns a unique hash representing the properties object.
	SetProperties(streamName string, propsModifier func(props *Stream) bool) (uint64, errors.Error)
	// SetState sets state on this tap
	SetState(state any) errors.Error
	// SetConfig sets the config of this tap
	SetConfig(config any) errors.Error
}

type Response struct {
	Out Output[json.RawMessage]
	Err errors.Error
}
