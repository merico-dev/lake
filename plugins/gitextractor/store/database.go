package store

import (
	"github.com/merico-dev/lake/models/domainlayer/code"
	"github.com/merico-dev/lake/plugins/helper"
	"gorm.io/gorm"
	"reflect"
)

const BathSize = 100

type Database struct {
	db            *gorm.DB
	repoCommits   *helper.BatchSave
	commits       *helper.BatchSave
	refs          *helper.BatchSave
	commitFiles   *helper.BatchSave
	commitParents *helper.BatchSave
}

func NewDatabase(db *gorm.DB) (*Database, error) {
	var err error
	database := new(Database)
	database.repoCommits, err = helper.NewBatchSave(db, reflect.TypeOf(&code.RepoCommit{}), BathSize)
	if err != nil {
		return nil, err
	}
	database.commits, err = helper.NewBatchSave(db, reflect.TypeOf(&code.Commit{}), BathSize)
	if err != nil {
		return nil, err
	}
	database.refs, err = helper.NewBatchSave(db, reflect.TypeOf(&code.Ref{}), BathSize)
	if err != nil {
		return nil, err
	}
	database.commitFiles, err = helper.NewBatchSave(db, reflect.TypeOf(&code.CommitFile{}), BathSize)
	if err != nil {
		return nil, err
	}
	database.commitParents, err = helper.NewBatchSave(db, reflect.TypeOf(&code.CommitParent{}), BathSize)
	if err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) RepoCommits(repoCommit *code.RepoCommit) error {
	return d.repoCommits.Add(repoCommit)
}

func (d *Database) Commits(commit *code.Commit) error {
	return d.commits.Add(commit)
}

func (d *Database) Refs(ref *code.Ref) error {
	return d.refs.Add(ref)
}

func (d *Database) CommitFiles(file *code.CommitFile) error {
	return d.commitFiles.Add(file)
}

func (d *Database) CommitParents(pp []*code.CommitParent) error {
	for _, cp := range pp {
		err := d.commitParents.Add(cp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Database) Flush() error {
	var err error
	err = d.repoCommits.Flush()
	if err != nil {
		return err
	}
	err = d.commits.Flush()
	if err != nil {
		return err
	}
	err = d.refs.Flush()
	if err != nil {
		return err
	}
	err = d.commitParents.Flush()
	if err != nil {
		return err
	}
	err = d.commitFiles.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) Close() error {
	return nil
}
