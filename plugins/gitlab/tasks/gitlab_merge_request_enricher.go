package tasks

import (
	"context"
	lakeModels "github.com/merico-dev/lake/models"
	gitlabModels "github.com/merico-dev/lake/plugins/gitlab/models"
	"gorm.io/gorm/clause"
	"time"
)

func EnrichMergeRequests(ctx context.Context, projectId int) error {
	// get mrs from theDB
	cursor, err := lakeModels.Db.Model(&gitlabModels.GitlabMergeRequest{}).Where("project_id = ?", projectId).Rows()
	if err != nil {
		return err
	}
	defer cursor.Close()

	gitlabMr := &gitlabModels.GitlabMergeRequest{}
	for cursor.Next() {
		err = lakeModels.Db.ScanRows(cursor, gitlabMr)
		if err != nil {
			return err
		}
		// enrich first_comment_time field
		notes := make([]gitlabModels.GitlabMergeRequestNote, 0)
		// `system` = 0 is needed since we only care about human comments
		lakeModels.Db.Where("merge_request_id = ? AND `system` = 0", gitlabMr.GitlabId).
			Order("gitlab_created_at asc").Find(&notes)
		commits := make([]gitlabModels.GitlabCommit, 0)
		lakeModels.Db.Joins("join gitlab_merge_request_commits gmrc on gmrc.commit_sha = gitlab_commits.sha").
			Where("merge_request_id = ?", gitlabMr.GitlabId).Order("authored_date asc").Find(&commits)
		// calculate reviewRounds from commits and notes
		reviewRounds := getReviewRounds(commits, notes)
		gitlabMr.ReviewRounds = reviewRounds

		if len(notes) > 0 {
			earliestNote, err := findEarliestNote(notes)
			if err != nil {
				return err
			}
			if earliestNote != nil {
				gitlabMr.FirstCommentTime = &earliestNote.GitlabCreatedAt
			}
		}

		err = lakeModels.Db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(gitlabMr).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func findEarliestNote(notes []gitlabModels.GitlabMergeRequestNote) (*gitlabModels.GitlabMergeRequestNote, error) {
	var earliestNote *gitlabModels.GitlabMergeRequestNote
	earliestTime := time.Now()
	for i, _ := range notes {
		if !notes[i].Resolvable {
			continue
		}
		noteTime := notes[i].GitlabCreatedAt
		if noteTime.Before(earliestTime) {
			earliestTime = noteTime
			earliestNote = &notes[i]
		}
	}
	return earliestNote, nil
}

func getReviewRounds(commits []gitlabModels.GitlabCommit, notes []gitlabModels.GitlabMergeRequestNote) int {
	i := 0
	j := 0
	reviewRounds := 0
	if len(commits) == 0 && len(notes) == 0 {
		return 1
	}
	// state is used to keep track of previous activity
	// 0: init, 1: commit, 2: comment
	// whenever state is switched to comment, we increment reviewRounds by 1
	state := 0 // 0, 1, 2
	for i < len(commits) && j < len(notes) {
		if commits[i].AuthoredDate.Before(notes[j].GitlabCreatedAt) {
			i++
			state = 1
		} else {
			j++
			if state != 2 {
				reviewRounds++
			}
			state = 2
		}
	}
	// There's another implicit round of review in 2 scenarios
	// One: the last state is commit (state == 1)
	// Two: the last state is comment but there're still commits left
	if state == 1 || i < len(commits) {
		reviewRounds++
	}
	return reviewRounds
}
