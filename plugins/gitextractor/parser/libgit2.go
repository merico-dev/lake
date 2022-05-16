package parser

import (
	"context"
	"fmt"

	git "github.com/libgit2/git2go/v33"

	"github.com/merico-dev/lake/models/domainlayer"
	"github.com/merico-dev/lake/models/domainlayer/code"
	"github.com/merico-dev/lake/plugins/core"
	"github.com/merico-dev/lake/plugins/gitextractor/models"
)

const (
	BRANCH = "BRANCH"
	TAG    = "TAG"
)

type LibGit2 struct {
	store      models.Store
	logger     core.Logger
	ctx        context.Context     // for canceling
	subTaskCtx core.SubTaskContext // for updating progress
}

func NewLibGit2(store models.Store, subTaskCtx core.SubTaskContext) *LibGit2 {
	return &LibGit2{store: store,
		logger:     subTaskCtx.GetLogger(),
		ctx:        subTaskCtx.GetContext(),
		subTaskCtx: subTaskCtx}
}

func (l *LibGit2) LocalRepo(repoPath, repoId string) error {
	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return err
	}
	return l.run(repo, repoId)
}

func (l *LibGit2) run(repo *git.Repository, repoId string) error {
	defer l.store.Close()
	l.subTaskCtx.SetProgress(0, -1)

	// collect tags
	var err error
	err = repo.Tags.Foreach(func(name string, id *git.Oid) error {
		select {
		case <-l.ctx.Done():
			return l.ctx.Err()
		default:
			break
		}
		var err1 error
		var tag *git.Tag
		var tagCommit string
		tag, _ = repo.LookupTag(id)
		if tag != nil {
			tagCommit = tag.TargetId().String()
		} else {
			tagCommit = id.String()
		}
		l.logger.Info("tagCommit", tagCommit)
		if tagCommit != "" {
			ref := &code.Ref{
				DomainEntity: domainlayer.DomainEntity{Id: fmt.Sprintf("%s:%s", repoId, name)},
				RepoId:       repoId,
				Name:         name,
				CommitSha:    tagCommit,
				RefType:      TAG,
			}
			err1 = l.store.Refs(ref)
			if err1 != nil {
				return err1
			}
			l.subTaskCtx.IncProgress(1)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// collect branches
	var repoInter *git.BranchIterator
	repoInter, err = repo.NewBranchIterator(git.BranchAll)
	if err != nil {
		return err
	}
	err = repoInter.ForEach(func(branch *git.Branch, branchType git.BranchType) error {
		select {
		case <-l.ctx.Done():
			return l.ctx.Err()
		default:
			break
		}
		if branch.IsBranch() {
			name, err1 := branch.Name()
			if err1 != nil {
				return err1
			}
			var sha string
			if oid := branch.Target(); oid != nil {
				sha = oid.String()
			}
			ref := &code.Ref{
				DomainEntity: domainlayer.DomainEntity{Id: fmt.Sprintf("%s:%s", repoId, name)},
				RepoId:       repoId,
				Name:         name,
				CommitSha:    sha,
				RefType:      BRANCH,
			}
			ref.IsDefault, _ = branch.IsHead()
			err1 = l.store.Refs(ref)
			if err1 != nil {
				return err1
			}
			l.subTaskCtx.IncProgress(1)
			return nil
		}
		return nil
	})
	if err != nil {
		return err
	}

	// collect commits
	opts, err := git.DefaultDiffOptions()
	if err != nil {
		return err
	}
	opts.NotifyCallback = func(diffSoFar *git.Diff, delta git.DiffDelta, matchedPathSpec string) error {
		return nil
	}

	odb, err := repo.Odb()
	if err != nil {
		return err
	}
	err = odb.ForEach(func(id *git.Oid) error {
		select {
		case <-l.ctx.Done():
			return l.ctx.Err()
		default:
			break
		}
		commit, _ := repo.LookupCommit(id)
		if commit == nil {
			return nil
		}
		commitSha := commit.Id().String()
		l.logger.Info("process commit: %s", commitSha)
		c := &code.Commit{
			Sha:     commitSha,
			Message: commit.Message(),
		}
		author := commit.Author()
		if author != nil {
			c.AuthorName = author.Name
			c.AuthorEmail = author.Email
			c.AuthorId = author.Email
			c.AuthoredDate = author.When
		}
		committer := commit.Committer()
		if committer != nil {
			c.CommitterName = committer.Name
			c.CommitterEmail = committer.Email
			c.CommitterId = committer.Email
			c.CommittedDate = committer.When
		}
		var commitParents []*code.CommitParent
		for i := uint(0); i < commit.ParentCount(); i++ {
			parent := commit.Parent(i)
			if parent != nil {
				if parentId := parent.Id(); parentId != nil {
					commitParents = append(commitParents, &code.CommitParent{
						CommitSha:       c.Sha,
						ParentCommitSha: parentId.String(),
					})
				}
			}
		}
		err2 := l.store.CommitParents(commitParents)
		if err2 != nil {
			return err2
		}
		if commit.ParentCount() > 0 {
			parent := commit.Parent(0)
			if parent != nil {
				var parentTree, tree *git.Tree
				parentTree, err2 = parent.Tree()
				if err2 != nil {
					return err2
				}
				tree, err2 = commit.Tree()
				if err2 != nil {
					return err2
				}
				var diff *git.Diff
				diff, err2 = repo.DiffTreeToTree(parentTree, tree, &opts)
				if err2 != nil {
					return err2
				}
				var commitFile *code.CommitFile
				err2 = diff.ForEach(func(file git.DiffDelta, progress float64) (
					git.DiffForEachHunkCallback, error) {
					if commitFile != nil {
						err2 = l.store.CommitFiles(commitFile)
						if err2 != nil {
							l.logger.Error("CommitFiles error:", err)
							return nil, err2
						}
					}
					commitFile = new(code.CommitFile)
					commitFile.CommitSha = c.Sha
					commitFile.FilePath = file.NewFile.Path
					return func(hunk git.DiffHunk) (git.DiffForEachLineCallback, error) {
						return func(line git.DiffLine) error {
							if line.Origin == git.DiffLineAddition {
								commitFile.Additions += line.NumLines
							}
							if line.Origin == git.DiffLineDeletion {
								commitFile.Deletions += line.NumLines
							}
							return nil
						}, nil
					}, nil
				}, git.DiffDetailLines)
				if err2 != nil {
					return err2
				}
				if commitFile != nil {
					err2 = l.store.CommitFiles(commitFile)
					if err2 != nil {
						l.logger.Error("CommitFiles error:", err)
					}
				}
				var stats *git.DiffStats
				stats, err2 = diff.Stats()
				if err2 != nil {
					return err2
				}
				c.Additions += stats.Insertions()
				c.Deletions += stats.Deletions()
			}
		}
		err2 = l.store.Commits(c)
		if err2 != nil {
			return err2
		}
		repoCommit := &code.RepoCommit{
			RepoId:    repoId,
			CommitSha: c.Sha,
		}
		err2 = l.store.RepoCommits(repoCommit)
		if err2 != nil {
			return err2
		}
		l.subTaskCtx.IncProgress(1)
		return nil
	})
	return err
}
