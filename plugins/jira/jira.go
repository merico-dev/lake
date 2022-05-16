package main // must be main for plugin entry point

import (
	"fmt"
	"net/http"
	"time"

	"github.com/merico-dev/lake/migration"
	"github.com/merico-dev/lake/plugins/core"
	"github.com/merico-dev/lake/plugins/jira/api"
	"github.com/merico-dev/lake/plugins/jira/models"
	"github.com/merico-dev/lake/plugins/jira/models/migrationscripts"
	"github.com/merico-dev/lake/plugins/jira/tasks"
	"github.com/merico-dev/lake/runner"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var _ core.PluginMeta = (*Jira)(nil)
var _ core.PluginInit = (*Jira)(nil)
var _ core.PluginTask = (*Jira)(nil)
var _ core.PluginApi = (*Jira)(nil)
var _ core.Migratable = (*Jira)(nil)

type Jira struct{}

func (plugin Jira) Init(config *viper.Viper, logger core.Logger, db *gorm.DB) error {
	api.Init(config, logger, db)
	return nil
}

func (plugin Jira) Description() string {
	return "To collect and enrich data from JIRA"
}

func (plugin Jira) SubTaskMetas() []core.SubTaskMeta {
	return []core.SubTaskMeta{
		{Name: "collectProjects", EntryPoint: tasks.CollectProjects, EnabledByDefault: true, Description: "collect Jira projects"},
		{Name: "extractProjects", EntryPoint: tasks.ExtractProjects, EnabledByDefault: true, Description: "extract Jira projects"},

		{Name: "collectBoard", EntryPoint: tasks.CollectBoard, EnabledByDefault: true, Description: "collect Jira board"},
		{Name: "extractBoard", EntryPoint: tasks.ExtractBoard, EnabledByDefault: true, Description: "extract Jira board"},

		{Name: "collectIssues", EntryPoint: tasks.CollectIssues, EnabledByDefault: true, Description: "collect Jira issues"},
		{Name: "extractIssues", EntryPoint: tasks.ExtractIssues, EnabledByDefault: true, Description: "extract Jira issues"},

		{Name: "collectChangelogs", EntryPoint: tasks.CollectChangelogs, EnabledByDefault: true, Description: "collect Jira change logs"},
		{Name: "extractChangelogs", EntryPoint: tasks.ExtractChangelogs, EnabledByDefault: true, Description: "extract Jira change logs"},

		{Name: "collectWorklogs", EntryPoint: tasks.CollectWorklogs, EnabledByDefault: true, Description: "collect Jira work logs"},
		{Name: "extractWorklogs", EntryPoint: tasks.ExtractWorklogs, EnabledByDefault: true, Description: "extract Jira work logs"},

		{Name: "collectRemotelinks", EntryPoint: tasks.CollectRemotelinks, EnabledByDefault: true, Description: "collect Jira remote links"},
		{Name: "extractRemotelinks", EntryPoint: tasks.ExtractRemotelinks, EnabledByDefault: true, Description: "extract Jira remote links"},

		{Name: "collectSprints", EntryPoint: tasks.CollectSprints, EnabledByDefault: true, Description: "collect Jira sprints"},
		{Name: "extractSprints", EntryPoint: tasks.ExtractSprints, EnabledByDefault: true, Description: "extract Jira sprints"},

		{Name: "convertBoard", EntryPoint: tasks.ConvertBoard, EnabledByDefault: true, Description: "convert Jira board"},

		{Name: "convertIssues", EntryPoint: tasks.ConvertIssues, EnabledByDefault: true, Description: "convert Jira issues"},

		{Name: "convertWorklogs", EntryPoint: tasks.ConvertWorklogs, EnabledByDefault: true, Description: "convert Jira work logs"},

		{Name: "convertChangelogs", EntryPoint: tasks.ConvertChangelogs, EnabledByDefault: true, Description: "convert Jira change logs"},

		{Name: "convertSprints", EntryPoint: tasks.ConvertSprints, EnabledByDefault: true, Description: "convert Jira sprints"},

		{Name: "convertIssueCommits", EntryPoint: tasks.ConvertIssueCommits, EnabledByDefault: true, Description: "convert Jira issue commits"},
		{Name: "convertIssueRepoCommits", EntryPoint: tasks.ConvertIssueRepoCommits, EnabledByDefault: false, Description: "convert Jira issue repo commits"},

		{Name: "convertUsers", EntryPoint: tasks.ConvertUsers, EnabledByDefault: true, Description: "convert Jira users"},
	}
}

func (plugin Jira) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, error) {
	var op tasks.JiraOptions
	var err error
	db := taskCtx.GetDb()
	err = mapstructure.Decode(options, &op)
	if err != nil {
		return nil, err
	}
	if op.ConnectionId == 0 {
		return nil, fmt.Errorf("connectionId is invalid")
	}
	connection := &models.JiraConnection{}
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
	jiraApiClient, err := tasks.NewJiraApiClient(taskCtx, connection)
	if err != nil {
		return nil, fmt.Errorf("failed to create jira api client: %w", err)
	}
	info, code, err := tasks.GetJiraServerInfo(jiraApiClient)
	if err != nil || code != http.StatusOK || info == nil {
		return nil, fmt.Errorf("fail to get server info: error:[%s] code:[%d]", err, code)
	}
	taskData := &tasks.JiraTaskData{
		Options:        &op,
		ApiClient:      jiraApiClient,
		Connection:     connection,
		JiraServerInfo: *info,
	}
	if !since.IsZero() {
		taskData.Since = &since
	}
	return taskData, nil
}

func (plugin Jira) RootPkgPath() string {
	return "github.com/merico-dev/lake/plugins/jira"
}

func (plugin Jira) MigrationScripts() []migration.Script {
	return []migration.Script{new(migrationscripts.InitSchemas), new(migrationscripts.UpdateSchemas20220505), new(migrationscripts.UpdateSchemas20220507)}
}

func (plugin Jira) ApiResources() map[string]map[string]core.ApiResourceHandler {
	return map[string]map[string]core.ApiResourceHandler{
		"test": {
			"POST": api.TestConnection,
		},
		"echo": {
			"POST": func(input *core.ApiResourceInput) (*core.ApiResourceOutput, error) {
				return &core.ApiResourceOutput{Body: input.Body}, nil
			},
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
		"connections/:connectionId/epics": {
			"GET": api.GetEpicsByConnectionId,
		},
		"connections/:connectionId/granularities": {
			"GET": api.GetGranularitiesByConnectionId,
		},
		"connections/:connectionId/boards": {
			"GET": api.GetBoardsByConnectionId,
		},
		"connections/:connectionId/type-mappings": {
			"POST": api.PostIssueTypeMappings,
			"GET":  api.ListIssueTypeMappings,
		},
		"connections/:connectionId/type-mappings/:userType": {
			"PUT":    api.PutIssueTypeMapping,
			"DELETE": api.DeleteIssueTypeMapping,
		},
		"connections/:connectionId/type-mappings/:userType/status-mappings": {
			"POST": api.PostIssueStatusMappings,
			"GET":  api.ListIssueStatusMappings,
		},
		"connections/:connectionId/type-mappings/:userType/status-mappings/:userStatus": {
			"PUT":    api.PutIssueStatusMapping,
			"DELETE": api.DeleteIssueStatusMapping,
		},
		"connections/:connectionId/proxy/rest/*path": {
			"GET": api.Proxy,
		},
	}
}

// Export a variable named PluginEntry for Framework to search and load
var PluginEntry Jira //nolint

// standalone mode for debugging
func main() {
	cmd := &cobra.Command{Use: "jira"}
	connectionId := cmd.Flags().Uint64P("connection", "s", 0, "jira connection id")
	boardId := cmd.Flags().Uint64P("board", "b", 0, "jira board id")
	_ = cmd.MarkFlagRequired("connection")
	_ = cmd.MarkFlagRequired("board")
	cmd.Run = func(c *cobra.Command, args []string) {
		runner.DirectRun(c, args, PluginEntry, map[string]interface{}{
			"connectionId": *connectionId,
			"boardId":      *boardId,
		})
	}
	runner.RunCmd(cmd)
}
