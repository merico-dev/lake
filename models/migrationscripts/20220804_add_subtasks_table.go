package migrationscripts

import (
	"context"
	commonArchived "github.com/apache/incubator-devlake/models/migrationscripts/archived"
	"gorm.io/gorm"
	"time"
)

type addSubtasksTable struct {
}

// Subtask20220804 DB snapshot model of models.Subtask
type Subtask20220804 struct {
	commonArchived.Model
	TaskID       uint64     `json:"task_id" gorm:"index"`
	SubtaskName  string     `json:"name" gorm:"column:name;index"`
	Number       int        `json:"number"`
	BeganAt      *time.Time `json:"beganAt"`
	FinishedAt   *time.Time `json:"finishedAt" gorm:"index"`
	SpentSeconds int64      `json:"spentSeconds"`
}

func (s Subtask20220804) TableName() string {
	return "_devlake_subtasks"
}

func (u addSubtasksTable) Up(_ context.Context, db *gorm.DB) error {
	err := db.Migrator().AutoMigrate(&Subtask20220804{})
	return err
}

func (u addSubtasksTable) Version() uint64 {
	return 20220804000001
}

func (u addSubtasksTable) Name() string {
	return "create subtask schema"
}
