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
