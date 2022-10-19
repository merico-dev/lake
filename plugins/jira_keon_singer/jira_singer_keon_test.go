package main

import (
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/jira_keon_singer/impl"
	"github.com/apache/incubator-devlake/runner"
	"github.com/spf13/cobra"
	"testing"
	"time"
)

type TestPlugin struct {
	*impl.JiraSinger
}

func (plugin TestPlugin) MigrationScripts() []core.MigrationScript {
	scripts := migrationscripts.GetCoreMigrations()
	return scripts
}

func TestJiraPlugin(t *testing.T) {
	cfg := config.GetConfig()
	cfg.Set("TAP_JIRA", "/home/keon/dev/python/venv-dev/bin/tap-jira")
	cfg.Set("SINGER_PROPERTIES_DIR", "/home/keon/dev/lake/config/singer")
	start, err := time.Parse("2006-01-02T15:04:05Z", "2022-06-01T15:04:05Z")
	if err != nil {
		panic(err)
	}
	//os.Stdin.Write([]byte{'A'})
	//t.SkipNow()
	//main()
	githubCmd := &cobra.Command{Use: "jira_singer"}
	//connectionId := githubCmd.Flags().Uint64P("connectionId", "c", 1, "github connection id")
	//owner := githubCmd.Flags().StringP("owner", "o", "apache", "github owner")
	//repo := githubCmd.Flags().StringP("repo", "r", "incubator-devlake", "github repo")
	//_ = githubCmd.MarkFlagRequired("owner")
	//_ = githubCmd.MarkFlagRequired("repo")

	pluginEntry := &TestPlugin{&impl.JiraSinger{}}
	//overrides...
	config.GetConfig().Set("DB_URL", "mysql://merico:merico@localhost:3306/lake?charset=utf8mb4&parseTime=True")
	githubCmd.Run = func(cmd *cobra.Command, args []string) {
		runner.DirectRun(cmd, args, pluginEntry, map[string]interface{}{
			"connectionId": 1,
			"start_date":   start,
		})
	}
	runner.RunCmd(githubCmd)
}
