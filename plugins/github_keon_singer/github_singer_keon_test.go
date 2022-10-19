package main

import (
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/migration"
	"github.com/apache/incubator-devlake/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/core"
	migrationscripts2 "github.com/apache/incubator-devlake/plugins/github/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/github_keon_singer/impl"
	"github.com/apache/incubator-devlake/runner"
	"github.com/spf13/cobra"
	"testing"
	"time"
)

type TestPlugin struct {
	*impl.GithubSinger
}

func (plugin TestPlugin) MigrationScripts() []core.MigrationScript {
	scripts := migrationscripts.GetCoreMigrations()
	return scripts
}

func registerPluginMigrations() {
	db, err := runner.NewGormDb(config.GetConfig(), nil)
	if err != nil {
		panic(err)
	}
	migration.Init(db)
	scripts := migrationscripts2.All()
	migration.Register(scripts, "some comment")
}

func TestGitHubPlugin(t *testing.T) {
	cfg := config.GetConfig()
	cfg.Set("TAP_GITHUB", "/home/keon/dev/python/venv-dev/bin/tap-github")
	cfg.Set("SINGER_PROPERTIES_DIR", "/home/keon/dev/lake/config/singer")
	start, err := time.Parse("2006-01-02T15:04:05Z", "2022-06-01T15:04:05Z")
	if err != nil {
		panic(err)
	}
	//os.Stdin.Write([]byte{'A'})
	//t.SkipNow()
	//main()
	githubCmd := &cobra.Command{Use: "github_singer"}
	//connectionId := githubCmd.Flags().Uint64P("connectionId", "c", 1, "github connection id")
	//owner := githubCmd.Flags().StringP("owner", "o", "apache", "github owner")
	//repo := githubCmd.Flags().StringP("repo", "r", "incubator-devlake", "github repo")
	//_ = githubCmd.MarkFlagRequired("owner")
	//_ = githubCmd.MarkFlagRequired("repo")

	pluginEntry := &TestPlugin{&impl.GithubSinger{}}

	//overrides...
	config.GetConfig().Set("DB_URL", "mysql://merico:merico@localhost:3307/lake?charset=utf8mb4&parseTime=True")
	config.GetConfig().Set("FORCE_MIGRATION", true)

	registerPluginMigrations()

	githubCmd.Run = func(cmd *cobra.Command, args []string) {
		runner.DirectRun(cmd, args, pluginEntry, map[string]interface{}{
			"connectionId": 1,
			"owner":        "apache",
			"repo":         "apache/incubator-devlake",
			"start_date":   start,
		})
	}
	runner.RunCmd(githubCmd)
}
