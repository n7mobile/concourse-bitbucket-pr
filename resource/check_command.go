package resource

import (
	"fmt"
	"sort"

	"github.com/n7mobile/concourse-bitbucket-pr/bitbucket"
	"github.com/n7mobile/concourse-bitbucket-pr/concourse"
	"github.com/n7mobile/concourse-bitbucket-pr/resource/models"
)

// CheckCommand fetches list of PullRequests form BitBucket API and traslates it list versions sorted by PR Identifier
type CheckCommand struct {
	Logger *concourse.Logger
}

// Run CheckCommand processing.
func (cmd *CheckCommand) Run(req models.CheckRequest) ([]models.Version, error) {
	err := req.Source.Validate()
	if err != nil {
		return nil, fmt.Errorf("resource/check: source invalid: %w", err)
	}

	auth := bitbucket.Auth{
		Username: req.Source.Username,
		Password: req.Source.Password,
	}

	client := bitbucket.NewClient(req.Source.Workspace, req.Source.Slug, &auth)
	preqs, err := client.GetPullRequestsPaged()
	if err != nil {
		return nil, fmt.Errorf("resource/check: paged prs: %w", err)
	}

	sort.Slice(preqs, func(i int, j int) bool {
		return preqs[i].ID < preqs[j].ID
	})

	versions := []models.Version{}
	containsVer := false

	for _, preq := range preqs {
		versions = append(versions, models.Version{
			Commit: preq.Source.Commit.Hash,
		})

		containsVer = containsVer || preq.Source.Commit.Hash == req.Version.Commit
	}

	if len(req.Version.Commit) > 0 && !containsVer {
		versions = append([]models.Version{req.Version}, versions...)
	}

	return versions, nil
}
