package store

import (
	"github.com/merico-dev/lake/models/domainlayer/code"
	"github.com/merico-dev/lake/plugins/helper"
	"gorm.io/gorm"
	"reflect"
)

const BathSize = 100

type Database struct {
	db     *gorm.DB
	driver *helper.BatchSaveDivider
}

func NewDatabase(db *gorm.DB) *Database {
	database := new(Database)
	database.driver = helper.NewBatchSaveDivider(db, BathSize)
	return database
}

func (d *Database) RepoCommits(repoCommit *code.RepoCommit) error {
	batch, err := d.driver.ForType(reflect.TypeOf(repoCommit))
	if err != nil {
		return err
	}
	return batch.Add(repoCommit)
}

func (d *Database) Commits(commit *code.Commit) error {
	batch, err := d.driver.ForType(reflect.TypeOf(commit))
	if err != nil {
		return err
	}
	return batch.Add(commit)
}

func (d *Database) Refs(ref *code.Ref) error {
	batch, err := d.driver.ForType(reflect.TypeOf(ref))
	if err != nil {
		return err
	}
	return batch.Add(ref)
}

func (d *Database) CommitFiles(file *code.CommitFile) error {
	batch, err := d.driver.ForType(reflect.TypeOf(file))
	if err != nil {
		return err
	}
	return batch.Add(file)
}

func (d *Database) CommitParents(pp []*code.CommitParent) error {
	if len(pp) == 0 {
		return nil
	}
	batch, err := d.driver.ForType(reflect.TypeOf(pp[0]))
	if err != nil {
		return err
	}
	for _, cp := range pp {
		err = batch.Add(cp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Database) Close() error {
	return d.driver.Close()
}
