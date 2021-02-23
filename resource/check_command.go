package resource

import (
	"fmt"
	"os"
	"sort"

	git "github.com/libgit2/git2go/v31"
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

	destination := "/tmp/" + req.Source.Slug

	url := bitbucket.NewClient(req.Source.Workspace, req.Source.Slug, &auth).RepoURL()
	repo, err := cmd.gitBareClone(req.Source.Username, req.Source.Password, url, destination)
	if err != nil {
		return nil, fmt.Errorf("resource/check: repo clone: %w", err)
	}

	commits := []*git.Commit{}
	versions := []models.Version{}

	if commit, err := cmd.getCommit(repo, req.Version.Ref); err == nil {
		commits = append(commits, commit)
	} else if len(req.Version.Ref) > 0 {
		versions = append(versions, req.Version)
	}

	for _, preq := range preqs {
		if commit, err := cmd.getCommit(repo, preq.Source.Commit.Hash); err == nil {
			commits = append(commits, commit)
		}
	}

	sort.Slice(commits, func(i int, j int) bool {
		timeI := commits[i].Committer().When
		timeJ := commits[j].Committer().When
		return timeI.Before(timeJ)
	})

	for _, comm := range commits {
		versions = append(versions, models.Version{
			Ref: comm.AsObject().Id().String(),
		})
	}

	err = os.RemoveAll(destination)
	if err != nil {
		return nil, fmt.Errorf("resource/check: remove tmp dir: %s", destination)
	}

	return versions, nil
}

func (cmd CheckCommand) gitBareClone(user, pass string, url string, destination string) (*git.Repository, error) {
	creds, err := git.NewCredentialUserpassPlaintext(user, pass)
	if err != nil {
		return nil, fmt.Errorf("resource/check: git creds: %w", err)
	}

	cmd.Logger.Debugf("resource/check: \tClone history from repo %s", url)

	opts := git.CloneOptions{
		FetchOptions: &git.FetchOptions{
			RemoteCallbacks: git.RemoteCallbacks{
				CredentialsCallback: func(url, username_from_url string, allowed_types git.CredentialType) (*git.Credential, error) {
					return creds, nil
				},
				CertificateCheckCallback: func(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
					return git.ErrorCodeOK
				},
			},
		},
		Bare: true,
	}

	repo, err := git.Clone(url, destination, &opts)
	if err != nil {
		return nil, fmt.Errorf("resource/in: cloning: %w", err)
	}

	return repo, nil
}

func (cmd CheckCommand) getCommit(repo *git.Repository, ref string) (*git.Commit, error) {
	cmd.Logger.Debugf("resource/in: \tReverse parsing of a shorthand ref '%s'", ref)

	refObj, err := repo.RevparseSingle(ref)
	if err != nil {
		return nil, fmt.Errorf("resource/in: RevparseSingle: %w", err)
	}

	cmd.Logger.Debugf("resource/in: \tLooking up for a commit with id '%s'", refObj.Id().String())

	commit, err := repo.LookupCommit(refObj.Id())
	if err != nil {
		return nil, fmt.Errorf("resource/in: LookupCommit: %w", err)
	}

	return commit, nil
}
