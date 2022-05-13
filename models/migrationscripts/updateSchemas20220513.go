package migrationscripts

import (
	"context"
	"fmt"

	"github.com/merico-dev/lake/models/migrationscripts/archived"
	"gorm.io/gorm"
)

type RefsIssuesDiffs20220513 struct {
	NewRefId        string `gorm:"primaryKey;type:varchar(255)"`
	OldRefId        string `gorm:"primaryKey;type:varchar(255)"`
	NewRefCommitSha string `gorm:"type:varchar(40)"`
	OldRefCommitSha string `gorm:"type:varchar(40)"`
	IssueNumber     string `gorm:"type:varchar(255)"`
	IssueId         string `gorm:"primaryKey;type:varchar(255)"`
	archived.NoPKModel
}

func (RefsIssuesDiffs20220513) TableName() string {
	return "refs_issues_diffs"
}

type updateSchemas20220513 struct{}

func (*updateSchemas20220513) Up(ctx context.Context, db *gorm.DB) error {
	cursor, err := db.Model(archived.RefsIssuesDiffs{}).Rows()
	if err != nil {
		return err
	}
	defer cursor.Close()

	duplicateData := make(map[string]interface{})
	for cursor.Next() {
		inputRow := archived.RefsIssuesDiffs{}
		err := db.ScanRows(cursor, &inputRow)
		if err != nil {
			return err

		}

		inputRowkey := fmt.Sprintf("%s%s%s", inputRow.NewRefId, inputRow.OldRefId, inputRow.IssueId)
		if inputRowkey != "" {
			duplicateData[inputRowkey] = inputRow
		}
	}

	err = db.Migrator().DropTable(archived.RefsIssuesDiffs{})
	if err != nil {
		return err
	}

	err = db.Migrator().CreateTable(RefsIssuesDiffs20220513{})
	if err != nil {
		return err
	}

	for _, v := range duplicateData {
		db.Create(v)
	}

	return nil
}

func (*updateSchemas20220513) Version() uint64 {
	return 20220513212319
}

func (*updateSchemas20220513) Name() string {
	return "refs_issues_diffs add primary key"
}
