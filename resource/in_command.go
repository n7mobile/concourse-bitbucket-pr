package resource

import (
	"fmt"
	"os"
	"path/filepath"

	git "github.com/libgit2/git2go/v31"
	"github.com/n7mobile/concourse-bitbucket-pr/bitbucket"
	"github.com/n7mobile/concourse-bitbucket-pr/concourse"
	"github.com/n7mobile/concourse-bitbucket-pr/resource/models"
)

// InCommand performs git checkout <commit_hash> into passed by Concourse destination directory
type InCommand struct {
	Logger *concourse.Logger
}

// Run InCommand processing.
func (cmd *InCommand) Run(destination string, req models.InRequest) (*models.InResponse, error) {
	err := req.Source.Validate()
	if err != nil {
		return nil, fmt.Errorf("resource/in: source validation: %w", err)
	}

	err = req.Version.Validate()
	if err != nil {
		return nil, fmt.Errorf("resource/in: version validation: %w", err)
	}

	cmd.Logger.Debugf("resource/in: Creating destination directory at %s", destination)

	path, _ := filepath.Split(destination)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return nil, fmt.Errorf("resource/in: creating destination: %w", err)
	}

	cmd.Logger.Debugf("resource/in: repo checkout...")

	auth := bitbucket.Auth{
		Username: req.Source.Username,
		Password: req.Source.Password,
	}

	url := bitbucket.NewClient(req.Source.Workspace, req.Source.Slug, &auth).RepoURL()

	commit, err := cmd.gitCheckoutRef(req.Source.Username, req.Source.Password, url, req.Version.Ref, destination)
	if err != nil {
		return nil, fmt.Errorf("resource/in: gitCheckoutRef: %w", err)
	}

	cmd.Logger.Debugf("resource/in: Checkout succeeded")

	return &models.InResponse{
		Version: req.Version,
		Metadata: models.Metadata{
			{Name: models.AuthorMetadataName, Value: commit.Author().Name},
			{Name: models.HashMetadataName, Value: commit.Id().String()},
			{Name: models.MessageMetadataName, Value: commit.Message()},
		},
	}, nil
}

func (cmd *InCommand) gitCheckoutRef(user, pass string, url, ref string, destination string) (*git.Commit, error) {
	creds, err := git.NewCredentialUserpassPlaintext(user, pass)
	if err != nil {
		return nil, fmt.Errorf("resource/in: git creds: %w", err)
	}

	cmd.Logger.Debugf("resource/in: \tClone from repo %s", url)

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
	}

	repo, err := git.Clone(url, destination, &opts)
	if err != nil {
		return nil, fmt.Errorf("resource/in: cloning: %w", err)
	}

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

	cmd.Logger.Debugf("resource/in: \tHard reset of the repo to commit '%s'", commit.Summary())

	err = repo.ResetToCommit(commit, git.ResetHard, &git.CheckoutOptions{})
	if err != nil {
		return nil, fmt.Errorf("resource/in: ResetToCommit: %w", err)
	}

	cmd.Logger.Debugf("resource/in: \tSetting head detached")

	err = repo.SetHeadDetached(refObj.Id())
	if err != nil {
		return nil, fmt.Errorf("resource/in: SetHeadDetached: %w", err)
	}

	return commit, nil
}
