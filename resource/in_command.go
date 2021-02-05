package resource

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	git "github.com/libgit2/git2go/v31"
	"github.com/n7mobile/ci-bitbucket-pr/bitbucket"
	"github.com/n7mobile/ci-bitbucket-pr/resource/models"
)

type InCommand struct {
	Logger *Logger
}

func (cmd *InCommand) Run(destination string, req models.InRequest) (models.InResponse, error) {
	err := req.Source.Validate()
	if err != nil {
		return models.InResponse{}, fmt.Errorf("source validation: %w", err)
	}

	err = req.Version.Validate()
	if err != nil {
		return models.InResponse{}, fmt.Errorf("version validation: %w", err)
	}

	cmd.Logger.Debugf("Creating destination directory at %s", destination)

	path, _ := filepath.Split(destination)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return models.InResponse{}, fmt.Errorf("creating destination: %w", err)
	}

	cmd.Logger.Debugf("Repository checkout...")

	url := bitbucket.RepositoryURL(req.Source.Workspace, req.Source.Slug)

	commit, err := cmd.gitCheckoutRef(req.Source.Username, req.Source.Password, url, req.Version.Branch, req.Version.Commit, destination)
	if err != nil {
		return models.InResponse{}, fmt.Errorf("gitCheckoutRef: %w", err)
	}

	cmd.Logger.Debugf("Checkout succeeded")

	cmd.Logger.Debugf("Version write as file...")

	err = cmd.writeVersion(&req.Version, destination)
	if err != nil {
		return models.InResponse{}, fmt.Errorf("version write: %w", err)
	}

	cmd.Logger.Debugf("Version write succeeded")

	return models.InResponse{
		Version: req.Version,
		Metadata: models.Metadata{
			{Name: "author", Value: commit.Author().Name},
			{Name: "commit", Value: commit.Id().String()},
			{Name: "message", Value: commit.Message()},
		},
	}, nil
}

func (cmd *InCommand) gitCheckoutRef(user, pass string, url, branch, ref string, destination string) (*git.Commit, error) {
	creds, err := git.NewCredentialUserpassPlaintext(user, pass)
	if err != nil {
		return nil, fmt.Errorf("git creds: %w", err)
	}

	cmd.Logger.Debugf("\tClone branch '%s' from repo %s", branch, url)

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
		CheckoutBranch: branch,
	}

	repo, err := git.Clone(url, destination, &opts)
	if err != nil {
		return nil, fmt.Errorf("cloning: %w", err)
	}

	cmd.Logger.Debugf("\tReverse parsing of a shorthand ref '%s'", ref)

	refObj, err := repo.RevparseSingle(ref)
	if err != nil {
		return nil, fmt.Errorf("RevparseSingle: %w", err)
	}

	cmd.Logger.Debugf("\tLooking up for a commit with id '%s'", refObj.Id().String())

	commit, err := repo.LookupCommit(refObj.Id())
	if err != nil {
		return nil, fmt.Errorf("LookupCommit: %w", err)
	}

	cmd.Logger.Debugf("\tHard reset of the repo to commit '%s'", commit.Summary())

	err = repo.ResetToCommit(commit, git.ResetHard, &git.CheckoutOptions{})
	if err != nil {
		return nil, fmt.Errorf("ResetToCommit: %w", err)
	}

	cmd.Logger.Debugf("\tSetting head detached")

	err = repo.SetHeadDetached(refObj.Id())
	if err != nil {
		return nil, fmt.Errorf("SetHeadDetached: %w", err)
	}

	return commit, nil
}

func (cmd *InCommand) writeVersion(version *models.Version, destination string) error {
	bytes, err := json.Marshal(version)
	if err != nil {
		return fmt.Errorf("marshaling current version: %w", err)
	}

	path := fmt.Sprintf("%s/version", destination)

	cmd.Logger.Debugf("\tWriting at path '%s' version data %s", path, string(bytes))

	err = ioutil.WriteFile(path, bytes, 0644)
	if err != nil {
		return fmt.Errorf("WriteFile: %w", err)
	}

	return nil
}
