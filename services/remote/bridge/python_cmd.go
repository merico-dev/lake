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

import "path/filepath"

const (
	pythonExec = "python"
	poetryExec = "poetry"
)

func NewPythonCmdInvoker(scriptPath string) *CmdInvoker {
	return NewCmdInvoker("", func(methodName string, args ...string) (string, []string) {
		allArgs := []string{scriptPath, methodName}
		allArgs = append(allArgs, args...)
		return pythonExec, allArgs
	})
}

func NewPythonPoetryCmdInvoker(scriptPath string) *CmdInvoker {
	tomlPath := filepath.Dir(filepath.Dir(scriptPath)) //the main entrypoint expected to be at toplevel
	return NewCmdInvoker(tomlPath, func(methodName string, args ...string) (string, []string) {
		allArgs := []string{"run", pythonExec, scriptPath, methodName}
		allArgs = append(allArgs, args...)
		return poetryExec, allArgs
	})
}
