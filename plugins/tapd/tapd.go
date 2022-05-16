package main

import (
	"fmt"
	"time"

	"github.com/merico-dev/lake/migration"
	"github.com/merico-dev/lake/models/domainlayer/didgen"
	"github.com/merico-dev/lake/plugins/core"
	"github.com/merico-dev/lake/plugins/tapd/api"
	"github.com/merico-dev/lake/plugins/tapd/models"
	"github.com/merico-dev/lake/plugins/tapd/models/migrationscripts"
	"github.com/merico-dev/lake/plugins/tapd/tasks"
	"github.com/merico-dev/lake/runner"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var _ core.PluginMeta = (*Tapd)(nil)
var _ core.PluginInit = (*Tapd)(nil)
var _ core.PluginTask = (*Tapd)(nil)
var _ core.PluginApi = (*Tapd)(nil)
var _ core.Migratable = (*Tapd)(nil)

type Tapd struct{}

func (plugin Tapd) Init(config *viper.Viper, logger core.Logger, db *gorm.DB) error {
	api.Init(config, logger, db)
	return nil
}

func (plugin Tapd) Description() string {
	return "To collect and enrich data from Tapd"
}

func (plugin Tapd) SubTaskMetas() []core.SubTaskMeta {
	return []core.SubTaskMeta{
		tasks.CollectWorkspaceMeta,
		tasks.ExtractWorkspaceMeta,
		tasks.CollectBugStatusMeta,
		tasks.ExtractBugStatusMeta,
		tasks.CollectUserMeta,
		tasks.ExtractUserMeta,
		tasks.CollectIterationMeta,
		tasks.ExtractIterationMeta,
		tasks.CollectStoryMeta,
		tasks.CollectBugMeta,
		tasks.CollectTaskMeta,
		tasks.ExtractStoryMeta,
		tasks.ExtractBugMeta,
		tasks.ExtractTaskMeta,
		tasks.CollectBugChangelogMeta,
		tasks.ExtractBugChangelogMeta,
		tasks.CollectStoryChangelogMeta,
		tasks.ExtractStoryChangelogMeta,
		tasks.CollectTaskChangelogMeta,
		tasks.ExtractTaskChangelogMeta,
		tasks.CollectWorklogMeta,
		tasks.ExtractWorklogMeta,
		tasks.CollectBugCommitMeta,
		tasks.ExtractBugCommitMeta,
		tasks.CollectStoryCommitMeta,
		tasks.ExtractStoryCommitMeta,
		tasks.CollectTaskCommitMeta,
		tasks.ExtractTaskCommitMeta,
		tasks.ConvertWorkspaceMeta,
		tasks.ConvertUserMeta,
		tasks.ConvertIterationMeta,
		tasks.ConvertStoryMeta,
		tasks.ConvertBugMeta,
		tasks.ConvertTaskMeta,
		tasks.ConvertWorklogMeta,
		tasks.ConvertBugChangelogMeta,
		tasks.ConvertStoryChangelogMeta,
		tasks.ConvertTaskChangelogMeta,
		tasks.ConvertBugCommitMeta,
		tasks.ConvertStoryCommitMeta,
		tasks.ConvertTaskCommitMeta,
		tasks.ConvertStoryLabelsMeta,
		tasks.ConvertTaskLabelsMeta,
		tasks.ConvertBugLabelsMeta,
	}
}

func (plugin Tapd) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, error) {
	db := taskCtx.GetDb()
	var op tasks.TapdOptions
	err := mapstructure.Decode(options, &op)
	if err != nil {
		return nil, err
	}
	if op.ConnectionId == 0 {
		return nil, fmt.Errorf("ConnectionId is required for Tapd execution")
	}
	connection := &models.TapdConnection{}
	err = db.First(connection, op.ConnectionId).Error
	if err != nil {
		return nil, err
	}
	var since time.Time
	if op.Since != "" {
		since, err = time.Parse("2006-01-02T15:04:05Z", op.Since)
		if err != nil {
			return nil, fmt.Errorf("invalid value for `since`: %w", err)
		}
	}
	tapdApiClient, err := tasks.NewTapdApiClient(taskCtx, connection)
	if err != nil {
		return nil, fmt.Errorf("failed to create tapd api client: %w", err)
	}
	taskData := &tasks.TapdTaskData{
		Options:    &op,
		ApiClient:  tapdApiClient,
		Connection: connection,
	}
	if !since.IsZero() {
		taskData.Since = &since
	}
	tasks.UserIdGen = didgen.NewDomainIdGenerator(&models.TapdUser{})
	tasks.WorkspaceIdGen = didgen.NewDomainIdGenerator(&models.TapdWorkspace{})
	tasks.IssueIdGen = didgen.NewDomainIdGenerator(&models.TapdIssue{})
	tasks.IterIdGen = didgen.NewDomainIdGenerator(&models.TapdIteration{})

	return taskData, nil
}

func (plugin Tapd) RootPkgPath() string {
	return "github.com/merico-dev/lake/plugins/tapd"
}

func (plugin Tapd) MigrationScripts() []migration.Script {
	return []migration.Script{new(migrationscripts.InitSchemas)}
}

func (plugin Tapd) ApiResources() map[string]map[string]core.ApiResourceHandler {
	return map[string]map[string]core.ApiResourceHandler{
		"test": {
			"POST": api.TestConnection,
		},
		"connections": {
			"POST": api.PostConnections,
			"GET":  api.ListConnections,
		},
		"connections/:connectionId": {
			"PATCH":  api.PatchConnection,
			"DELETE": api.DeleteConnection,
			"GET":    api.GetConnection,
		},
		"connections/:connectionId/boards": {
			"GET": api.GetBoardsByConnectionId,
		},
		"connections/:connectionId/proxy/rest/*path": {
			"GET": api.Proxy,
		},
	}
}

// Export a variable named PluginEntry for Framework to search and load
var PluginEntry Tapd //nolint

// standalone mode for debugging
func main() {
	cmd := &cobra.Command{Use: "tapd"}
	connectionId := cmd.Flags().Uint64P("connection", "c", 0, "tapd connection id")
	workspaceId := cmd.Flags().Uint64P("workspace", "w", 0, "tapd workspace id")
	err := cmd.MarkFlagRequired("connection")
	if err != nil {
		panic(err)
	}
	//err = cmd.MarkFlagRequired("company")
	//if err != nil {
	//	panic(err)
	//}
	err = cmd.MarkFlagRequired("workspace")
	if err != nil {
		panic(err)
	}
	cmd.Run = func(c *cobra.Command, args []string) {
		runner.DirectRun(c, args, PluginEntry, map[string]interface{}{
			"connectionId": *connectionId,
			"workspaceId":  *workspaceId,
		})
	}
	runner.RunCmd(cmd)
}
