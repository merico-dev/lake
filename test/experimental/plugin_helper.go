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

package experimental

import (
	"fmt"
	"github.com/apache/incubator-devlake/errors"
	"github.com/apache/incubator-devlake/runner"
	"github.com/apache/incubator-devlake/utils"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func dynamicallyLoad(sb *DevlakeSandbox, forceCompile bool, path string, pluginNames ...string) {
	if len(pluginNames) == 0 {
		return
	}
	var err errors.Error
	for _, pluginName := range pluginNames {
		newDestination := filepath.Join(throwawayDir, pluginName+".so")
		if forceCompile {
			rootPath := ""
			rootPath, err = getRootDir(path)
			require.NoError(sb.testCtx, err)
			err = compile(sb.testCtx, rootPath, fmt.Sprintf("plugins/%s", pluginName), newDestination)
		} else {
			pluginFilePath := filepath.Join(path, pluginName, pluginName+".so")
			err = errors.Convert(os.Symlink(pluginFilePath, newDestination))
		}
		require.NoError(sb.testCtx, err)
	}
	err = runner.LoadPlugins(throwawayDir, sb.cfg, sb.log, sb.db)
	require.NoError(sb.testCtx, err)
}

func compile(t *testing.T, rootPath string, mainPluginPkgDir string, outputPluginFile string) errors.Error {
	flags := ""
	if isUsingDebuggingSymbols() {
		flags = "-gcflags=\\'all=-N -l\\'"
	}
	args := []string{
		"/home/keon/go/bin/go",
		"build",
		//"-mod",
		//"vendor",
		"-v",
		//"-trimpath",
		flags,
		"-buildmode=plugin",
		"-o",
		outputPluginFile,
		fmt.Sprintf("./%s", mainPluginPkgDir),
	}
	fmt.Printf("calling: %+v\n", args)
	cmd := exec.Command(args[0], sanitize(args[1:]...)...)
	cmd.Dir = rootPath
	run, err := utils.StreamProcess(cmd, &utils.StreamProcessOptions{
		OnStdout: func(b []byte) (any, errors.Error) {
			fmt.Print(string(b))
			return nil, nil
		},
		OnStderr: func(b []byte) (any, errors.Error) {
			fmt.Print(string(b))
			return nil, nil
		},
	})
	if err != nil {
		return err
	}
	t.Cleanup(func() {
		fmt.Println("removing generated .so file: ", outputPluginFile)
		_ = os.Remove(outputPluginFile)
	})
	for response := range run.Receive() {
		if err = response.GetError(); err != nil {
			return err
		}
	}
	return nil
}

func getRootDir(path string) (string, errors.Error) {
	parts := strings.Split(path, "lake")
	if len(parts) < 2 {
		return "", errors.Default.New(fmt.Sprintf("invalid path: %s", path))
	}
	return filepath.Join(parts[0], "lake"), nil
}

func sanitize(args ...string) []string {
	cleaned := []string{}
	for _, arg := range args {
		if arg != "" {
			cleaned = append(cleaned, arg)
		}
	}
	return cleaned
}

func isUsingDebuggingSymbols() bool {
	// TODO figure this out later
	return false
}
