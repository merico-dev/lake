package tasks

import (
	"encoding/json"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ExtractCommitsMeta = core.SubTaskMeta{
	Name:             "extractApiCommits",
	EntryPoint:       ExtractApiCommits,
	EnabledByDefault: true,
	Description:      "Extract raw commit data into tool layer table GiteeCommit,GiteeUser and GiteeRepoCommit",
}

type GiteeCommit struct {
	Author struct {
		Date  helper.Iso8601Time `json:"date"`
		Email string             `json:"email"`
		Name  string             `json:"name"`
	}
	Committer struct {
		Date  helper.Iso8601Time `json:"date"`
		Email string             `json:"email"`
		Name  string             `json:"name"`
	}
	Message string `json:"message"`
}

type GiteeApiCommitResponse struct {
	Author      *models.GiteeUser `json:"author"`
	AuthorId    int
	CommentsUrl string            `json:"comments_url"`
	Commit      GiteeCommit       `json:"commit"`
	Committer   *models.GiteeUser `json:"committer"`
	HtmlUrl     string            `json:"html_url"`
	Sha         string            `json:"sha"`
	Url         string            `json:"url"`
}

func ExtractApiCommits(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_COMMIT_TABLE)

	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Extract: func(row *helper.RawData) ([]interface{}, error) {
			results := make([]interface{}, 0, 4)

			commit := &GiteeApiCommitResponse{}

			err := json.Unmarshal(row.Data, commit)

			if err != nil {
				return nil, err
			}

			if commit.Sha == "" {
				return nil, nil
			}

			giteeCommit, err := ConvertCommit(commit)

			if err != nil {
				return nil, err
			}

			if commit.Author != nil {
				giteeCommit.AuthorId = commit.Author.Id
				results = append(results, commit.Author)
			}
			if commit.Committer != nil {
				giteeCommit.CommitterId = commit.Committer.Id
				results = append(results, commit.Committer)

			}

			giteeRepoCommit := &models.GiteeRepoCommit{
				RepoId:    data.Repo.GiteeId,
				CommitSha: commit.Sha,
			}
			results = append(results, giteeCommit)
			results = append(results, giteeRepoCommit)
			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}

// ConvertCommit Convert the API response to our DB model instance
func ConvertCommit(commit *GiteeApiCommitResponse) (*models.GiteeCommit, error) {
	giteeCommit := &models.GiteeCommit{
		Sha:            commit.Sha,
		AuthorId:       commit.Author.Id,
		Message:        commit.Commit.Message,
		AuthorName:     commit.Commit.Author.Name,
		AuthorEmail:    commit.Commit.Author.Email,
		AuthoredDate:   commit.Commit.Author.Date.ToTime(),
		CommitterName:  commit.Commit.Author.Name,
		CommitterEmail: commit.Commit.Author.Email,
		CommittedDate:  commit.Commit.Author.Date.ToTime(),
		WebUrl:         commit.Url,
	}
	return giteeCommit, nil
}
