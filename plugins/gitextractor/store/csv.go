package store

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/merico-dev/lake/models/domainlayer/code"
)

type csvWriter struct {
	f *os.File
	w *csv.Writer
}

func newCsvWriter(path string, v interface{}) (*csvWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	// declare UTF-8 encoding
	_, err = f.WriteString("\xEF\xBB\xBF")
	if err != nil {
		return nil, err
	}
	w := csv.NewWriter(f)
	value := reflect.Indirect(reflect.ValueOf(v))
	var header []string
	for i := 0; i < value.NumField(); i++ {
		if value.Type().Field(i).Anonymous {
			continue
		}
		header = append(header, value.Type().Field(i).Name)
	}
	err = w.Write(header)
	if err != nil {
		return nil, err
	}
	return &csvWriter{f: f, w: w}, nil
}

func (w *csvWriter) Write(item interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(item))
	n := v.NumField()
	record := make([]string, 0, n)
	for i := 0; i < n; i++ {
		if v.Type().Field(i).Anonymous {
			continue
		}
		record = append(record, fmt.Sprint(v.Field(i).Interface()))
	}
	return w.w.Write(record)
}

func (w *csvWriter) Close() error {
	w.w.Flush()
	return w.f.Close()
}

type CsvStore struct {
	dir                string
	repoCommitWriter   *csvWriter
	commitWriter       *csvWriter
	refWriter          *csvWriter
	commitFileWriter   *csvWriter
	commitParentWriter *csvWriter
}

func NewCsvStore(dir string) (*CsvStore, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return nil, err
		}
	}
	var err error
	s := &CsvStore{dir: dir}
	s.repoCommitWriter, err = newCsvWriter(filepath.Join(dir, "repo_commits.csv"), code.RepoCommit{})
	if err != nil {
		return nil, err
	}
	s.commitWriter, err = newCsvWriter(filepath.Join(dir, "commits.csv"), code.Commit{})
	if err != nil {
		return nil, err
	}
	s.refWriter, err = newCsvWriter(filepath.Join(dir, "refs.csv"), code.Ref{})
	if err != nil {
		return nil, err
	}
	s.commitFileWriter, err = newCsvWriter(filepath.Join(dir, "commit_files.csv"), code.CommitFile{})
	if err != nil {
		return nil, err
	}
	s.commitParentWriter, err = newCsvWriter(filepath.Join(dir, "commit_parents.csv"), code.CommitParent{})
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (c *CsvStore) RepoCommits(repoCommit *code.RepoCommit) error {
	return c.repoCommitWriter.Write(repoCommit)
}

func (c *CsvStore) Commits(commit *code.Commit) error {
	return c.commitWriter.Write(commit)
}

func (c *CsvStore) Refs(ref *code.Ref) error {
	return c.refWriter.Write(ref)
}

func (c *CsvStore) CommitFiles(file *code.CommitFile) error {
	return c.commitFileWriter.Write(file)
}

func (c *CsvStore) CommitParents(pp []*code.CommitParent) error {
	var err error
	for _, p := range pp {
		err = c.commitParentWriter.Write(p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CsvStore) Close() error {
	if c.repoCommitWriter != nil {
		c.repoCommitWriter.Close()
	}
	if c.commitWriter != nil {
		c.commitWriter.Close()
	}
	if c.refWriter != nil {
		c.refWriter.Close()
	}
	if c.commitFileWriter != nil {
		c.commitFileWriter.Close()
	}
	if c.commitParentWriter != nil {
		c.commitParentWriter.Close()
	}
	return nil
}
