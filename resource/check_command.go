package resource

import (
	"fmt"
	"sort"
	"strconv"

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

	sort.Slice(versions, func(i int, j int) bool {
		numI, _ := strconv.Atoi(versions[i].ID)
		numJ, _ := strconv.Atoi(versions[j].ID)
		return numI < numJ
	})

	return versions, nil
}
