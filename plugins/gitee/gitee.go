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

package main

import (
	"github.com/apache/incubator-devlake/plugins/gitee/impl"
	"github.com/apache/incubator-devlake/runner"
	"github.com/spf13/cobra"
)

var PluginEntry impl.Gitee //nolint

func main() {
	giteeCmd := &cobra.Command{Use: "gitee"}
	owner := giteeCmd.Flags().StringP("owner", "o", "", "gitee owner")
	repo := giteeCmd.Flags().StringP("repo", "r", "", "gitee repo")
	token := giteeCmd.Flags().StringP("auth", "a", "", "access token")
	_ = giteeCmd.MarkFlagRequired("owner")
	_ = giteeCmd.MarkFlagRequired("repo")

	giteeCmd.Run = func(cmd *cobra.Command, args []string) {
		runner.DirectRun(cmd, args, PluginEntry, map[string]interface{}{
			"owner": *owner,
			"repo":  *repo,
			"token": *token,
		})
	}
	runner.RunCmd(giteeCmd)
}
