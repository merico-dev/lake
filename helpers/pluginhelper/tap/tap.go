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
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/utils"
)

// Tap the abstract interface for Taps. Consumer code should not use concrete implementations directly.
type Tap interface {
	// Run runs the tap and returns a stream of results. Expected to be called after all the other Setters.
	Run() (<-chan *utils.ProcessResponse[Result], errors.Error)
	// GetName the name of this tap
	GetName() string
	// SetProperties enables the passed in required stream and may modify its properties at runtime. Returns a unique hash representing the properties object.
	SetProperties(requiredStream string) (uint64, errors.Error)
	// SetState sets state on this tap
	SetState(state interface{}) errors.Error
	// SetConfig sets the config of this tap
	SetConfig() errors.Error
}
