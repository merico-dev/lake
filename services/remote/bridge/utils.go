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
	"encoding/json"
	"github.com/apache/incubator-devlake/errors"
)

func serialize(args ...any) ([]string, errors.Error) {
	var serializedArgs []string
	for _, arg := range args {
		serializedArg, err := json.Marshal(arg)
		if err != nil {
			return nil, errors.Convert(err)
		}
		serializedArgs = append(serializedArgs, string(serializedArg))
	}
	return serializedArgs, nil
}

func deserialize(bytes json.RawMessage) ([]map[string]any, errors.Error) {
	if len(bytes) == 0 {
		return nil, nil
	}
	var result []map[string]any
	if bytes[0] == '{' {
		single := make(map[string]any)
		if err := json.Unmarshal(bytes, &single); err != nil {
			return nil, errors.Convert(err)
		}
		result = append(result, single)
	} else if bytes[0] == '[' {
		if err := json.Unmarshal(bytes, &result); err != nil {
			return nil, errors.Convert(err)
		}
	} else {
		return nil, errors.Default.New("malformed JSON from remote call")
	}
	return result, nil
}
