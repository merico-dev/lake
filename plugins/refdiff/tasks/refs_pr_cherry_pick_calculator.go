package tasks

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/merico-dev/lake/models/domainlayer/code"
	"github.com/merico-dev/lake/plugins/core"
	"gorm.io/gorm/clause"
)

type cherryPick struct {
	RepoName             string `gorm:"type:varchar(255)"`
	ParentPrKey          int
	CherrypickBaseBranch string `gorm:"type:varchar(255)"`
	CherrypickPrKey      int
	ParentPrUrl          string `gorm:"type:varchar(255)"`
	ParentPrId           string `gorm:"type:varchar(255)"`
	CreatedDate          time.Time
}

func CalculatePrCherryPick(taskCtx core.SubTaskContext) error {
	data := taskCtx.GetData().(*RefdiffTaskData)
	repoId := data.Options.RepoId
	ctx := taskCtx.GetContext()
	db := taskCtx.GetDb()
	var prTitleRegex *regexp.Regexp
	prTitlePattern := taskCtx.GetConfig("GITHUB_PR_TITLE_PATTERN")

	if len(prTitlePattern) > 0 {
		fmt.Println(prTitlePattern)
		prTitleRegex = regexp.MustCompile(prTitlePattern)
	}

	cursor, err := db.Model(&code.PullRequest{}).
		Joins("left join repos on pull_requests.base_repo_id = repos.id").
		Where("repos.id = ?", repoId).Rows()
	if err != nil {
		return err
	}

	defer cursor.Close()

	pr := &code.PullRequest{}
	var parentPrKeyInt int
	taskCtx.SetProgress(0, -1)

	// iterate all rows
	for cursor.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err = db.ScanRows(cursor, pr)
		if err != nil {
			return err
		}

		parentPrKey := ""
		if prTitleRegex != nil {
			groups := prTitleRegex.FindStringSubmatch(pr.Title)
			if len(groups) > 0 {
				parentPrKey = groups[1]
			}
		}

		if parentPrKeyInt, err = strconv.Atoi(parentPrKey); err != nil {
			continue
		}

		var parentPrId string
		err = db.Model(&code.PullRequest{}).
			Where("key = ? and base_repo_id = ?", parentPrKeyInt, repoId).
			Pluck("id", &parentPrId).Error
		if err != nil {
			return err
		}
		if len(parentPrId) == 0 {
			continue
		}
		pr.ParentPrId = parentPrId

		err = db.Save(pr).Error
		if err != nil {
			return err
		}
		taskCtx.IncProgress(1)
	}

	cursor2, err := db.Raw(
		`
			SELECT pr2.KEY                              AS parent_pr_key,
			       pr1.parent_pr_id                     AS parent_pr_id,
			       pr1.base_ref                         AS cherrypick_base_branch,
			       pr1.KEY                              AS cherrypick_pr_key,
			       repos.NAME                           AS repo_name,
			       Concat(repos.url, '/pull/', pr2.KEY) AS parent_pr_url,
 				   pr2.created_date
			FROM   pull_requests pr1
			       LEFT JOIN pull_requests pr2
			              ON pr1.parent_pr_id = pr2.id
			       LEFT JOIN repos
			              ON pr2.base_repo_id = repos.id
			WHERE  pr1.parent_pr_id != ''
			ORDER  BY pr1.parent_pr_id,
			          pr2.created_date,
					  pr1.base_ref ASC
			`).Rows()
	if err != nil {
		return err
	}
	defer cursor2.Close()

	var refsPrCherryPick *code.RefsPrCherrypick
	var lastParentPrId string
	var lastCreatedDate time.Time
	var cherrypickBaseBranches []string
	var cherrypickPrKeys []string
	for cursor2.Next() {
		var item cherryPick
		err = db.ScanRows(cursor2, &item)
		if err != nil {
			return err
		}
		if item.ParentPrId == lastParentPrId && item.CreatedDate == lastCreatedDate {
			cherrypickBaseBranches = append(cherrypickBaseBranches, item.CherrypickBaseBranch)
			cherrypickPrKeys = append(cherrypickPrKeys, strconv.Itoa(item.CherrypickPrKey))
		} else {
			if refsPrCherryPick != nil {
				refsPrCherryPick.CherrypickBaseBranches = strings.Join(cherrypickBaseBranches, ",")
				refsPrCherryPick.CherrypickPrKeys = strings.Join(cherrypickPrKeys, ",")
				err = db.Clauses(clause.OnConflict{
					UpdateAll: true,
				}).Create(refsPrCherryPick).Error
				if err != nil {
					return err
				}
			}
			lastParentPrId = item.ParentPrId
			lastCreatedDate = item.CreatedDate
			cherrypickBaseBranches = []string{item.CherrypickBaseBranch}
			cherrypickPrKeys = []string{strconv.Itoa(item.CherrypickPrKey)}
			refsPrCherryPick = &code.RefsPrCherrypick{
				RepoName:    item.RepoName,
				ParentPrKey: item.ParentPrKey,
				ParentPrUrl: item.ParentPrUrl,
				ParentPrId:  item.ParentPrId,
			}
		}
	}
	err = db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(refsPrCherryPick).Error
	if err != nil {
		return err
	}

	return nil
}

var CalculatePrCherryPickMeta = core.SubTaskMeta{
	Name:             "calculatePrCherryPick",
	EntryPoint:       CalculatePrCherryPick,
	EnabledByDefault: true,
	Description:      "Calculate pr cherry pick",
}
