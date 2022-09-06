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
	"github.com/apache/incubator-devlake/plugins/Zentao/impl"
	"github.com/apache/incubator-devlake/runner"
	"github.com/spf13/cobra"
)

// Export a variable named PluginEntry for Framework to search and load
var PluginEntry impl.Zentao //nolint

// standalone mode for debugging
func main() {
	cmd := &cobra.Command{Use: "zentao"}

	connectionId := cmd.Flags().Uint64P("connectionId", "c", 0, "zentao connection id")
	executionId := cmd.Flags().IntP("executionId", "e", 8, "execution id")
	productId := cmd.Flags().IntP("productId", "o", 8, "product id")
	projectId := cmd.Flags().IntP("projectId", "p", 8, "project id")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		runner.DirectRun(cmd, args, PluginEntry, map[string]interface{}{
			"connectionId": *connectionId,
			"executionId":  *executionId,
			"productId":    *productId,
			"projectId":    *projectId,
		})
	}
	runner.RunCmd(cmd)
}
