package resource

import (
	"strconv"

	"github.com/n7mobile/ci-bitbucket-pr/bitbucket"
	"github.com/n7mobile/ci-bitbucket-pr/resource/models"
)

type CheckCommand struct {
	Logger *Logger
}

func (cmd *CheckCommand) Run(req models.CheckRequest) ([]models.Version, error) {
	err := req.Source.Validate()
	if err != nil {
		return nil, err
	}

	qry := bitbucket.PullRequestQuery{
		Workspace: req.Source.Workspace,
		Slug:      req.Source.Slug,
	}

	client := bitbucket.NewBasicAuth(req.Source.Username, req.Source.Password)
	preqs, err := client.GetPullRequests(&qry)
	if err != nil {
		return nil, err
	}

	versions := []models.Version{}

	if len(req.Version.Commit) > 0 {
		versions = append(versions, req.Version)
	}

	for _, preq := range preqs {
		versions = append(versions, models.Version{
			Commit: preq.Source.Commit.Hash,
			ID:     strconv.Itoa(preq.ID),
			Title:  preq.Title,
			Branch: preq.Source.Branch.Name,
		})
	}

	return versions, nil
}
