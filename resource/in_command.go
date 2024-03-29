package resource

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	client := bitbucket.NewClient(req.Source.Workspace, req.Source.Slug, &auth)
	url := client.RepoURL()

	commit, err := cmd.gitCheckoutRef(req.Source.Username, req.Source.Password, url, req.Version.Ref, destination)
	if err != nil {
		return nil, fmt.Errorf("resource/in: gitCheckoutRef: %w", err)
	}

	cmd.Logger.Debugf("resource/in: Checkout succeeded")

	if req.Source.RecurseSubmodules {
		cmd.Logger.Debugf("resource/in: submodules update")

		cmd.gitUpdateSubmodules(req.Source.Username, req.Source.Password, commit.Owner())
	}

	branch, err := cmd.gitBranchOfCommit(commit)
	if err != nil {
		cmd.Logger.Errorf("resource/in: gitBranchOfCommit: %w", err)
	}

	cmd.Logger.Debugf("resource/in: version write to %s", concourse.VersionStorageFilename)

	err = concourse.NewStorage(destination, string(concourse.VersionStorageFilename)).Write(&req.Version)
	if err != nil {
		return nil, fmt.Errorf("resource/in: version write: %w", err)
	}

	response := models.InResponse{
		Version: req.Version,
		Metadata: models.Metadata{
			{Name: models.AuthorMetadataName, Value: commit.Author().Name},
			{Name: models.TimestampMetadataName, Value: commit.Author().When.String()},
			{Name: models.MessageMetadataName, Value: commit.Message()},
			{Name: models.CommitMetadataName, Value: commit.AsObject().Id().String()},
			{Name: models.PullrequestURLMetadataName, Value: client.PullrequestURL(req.Version.ID)},
		},
	}

	if branch != nil {
		response.Metadata = append(response.Metadata, models.MetadataField{
			Name:  models.BranchMetadataName,
			Value: *branch,
		})
	}

	return &response, nil
}

func (cmd *InCommand) gitCheckoutRef(user, pass string, url, ref string, destination string) (*git.Commit, error) {
	creds, err := git.NewCredentialUserpassPlaintext(user, pass)
	if err != nil {
		return nil, fmt.Errorf("resource/in: git creds: %w", err)
	}

	cmd.Logger.Debugf("resource/in: Clone from repo '%s'", url)

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
		return nil, fmt.Errorf("resource/in: Cloning: %w", err)
	}

	commit, err := cmd.gitCheckoutDetachedHead(repo, ref)
	if err != nil {
		return nil, fmt.Errorf("resource/in: Detach head at %s: %w", ref, err)
	}

	return commit, nil
}

func (cmd InCommand) gitUpdateSubmodules(user, pass string, repo *git.Repository) {
	opts := &git.SubmoduleUpdateOptions{
		FetchOptions: &git.FetchOptions{
			RemoteCallbacks: git.RemoteCallbacks{
				CredentialsCallback: func(url, username_from_url string, allowed_types git.CredentialType) (*git.Credential, error) {
					return git.NewCredentialUserpassPlaintext(user, pass)
				},
				CertificateCheckCallback: func(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
					return git.ErrorCodeOK
				},
			},
		},
	}

	repo.Submodules.Foreach(func(sub *git.Submodule, name string) int {
		cmd.Logger.Debugf("resource/in: Submodule '%s' update to commit %s", name, sub.HeadId().String())

		err := sub.Update(true, opts)

		if err != nil {
			cmd.Logger.Errorf("resource/in: Submodule '%s' update: %w", name, err)
		}

		repo, err := sub.Open()

		if err != nil {
			cmd.Logger.Errorf("resource/in: Submodule '%s' open: %w", name, err)
		}

		_, err = cmd.gitCheckoutDetachedHead(repo, sub.HeadId().String())

		if err != nil {
			cmd.Logger.Errorf("resource/in: Submodule '%s' detach head: %w", name, err)
		}

		return 0
	})
}

func (cmd InCommand) gitCheckoutDetachedHead(repo *git.Repository, ref string) (*git.Commit, error) {
	remote, err := repo.Remotes.Lookup("origin")

	if err != nil {
		return nil, fmt.Errorf("resource/in: Repo lookup origin remote: %w", err)
	}

	repoURL := remote.Url()

	cmd.Logger.Debugf("resource/in: %s: Reverse parsing of a shorthand ref '%s'", repoURL, ref)

	refObj, err := repo.RevparseSingle(ref)
	if err != nil {
		return nil, fmt.Errorf("resource/in: %s:  Repo revparseSingle: %w", repoURL, err)
	}

	cmd.Logger.Debugf("resource/in: %s: Looking up for a commit with id '%s'", repoURL, refObj.Id().String())

	commit, err := repo.LookupCommit(refObj.Id())
	if err != nil {
		return nil, fmt.Errorf("resource/in: %s: LookupCommit: %w", repoURL, err)
	}

	err = repo.SetHeadDetached(commit.Id())
	if err != nil {
		return nil, fmt.Errorf("resource/in: %s: SetHeadDetached: %w", repoURL, err)
	}

	err = repo.CheckoutHead(&git.CheckoutOptions{Strategy: git.CheckoutForce})
	if err != nil {
		return nil, fmt.Errorf("resource/in: %s: CheckoutHead: %w", repoURL, err)
	}

	return commit, nil
}

// gitBranchOfCommit iterates over commits of every remote branch and compares its refs
// First ref matches yields name of the branch
func (cmd InCommand) gitBranchOfCommit(commit *git.Commit) (*string, error) {
	repo := commit.Owner()
	if repo == nil {
		return nil, errors.New("resource/in: owner of commit is empty")
	}

	iterator, err := repo.NewBranchIterator(git.BranchRemote)
	if err != nil {
		return nil, fmt.Errorf("resource/in: iterator for remote branches: %w", err)
	}

	var branch *git.Branch

	err = iterator.ForEach(func(b *git.Branch, bt git.BranchType) error {
		head := b.Target()
		if head == nil {
			return nil
		}

		walk, err := repo.Walk()
		if err != nil {
			return fmt.Errorf("resource/in: branch walk: %w", err)
		}

		err = walk.Push(head)
		if err != nil {
			return fmt.Errorf("resource/in: walk push %s: %w", head.String(), err)
		}

		found := false

		walk.Sorting(git.SortTopological)
		walk.Iterate(func(c *git.Commit) bool {
			found = c.Id().Equal(commit.Id())
			return !found
		})

		if found {
			branch = b
		}

		return nil
	})

	if branch == nil {
		return nil, fmt.Errorf("resource/in: branch with commit not found")
	}

	refName, err := branch.Name()
	if err != nil {
		return nil, fmt.Errorf("resource/in: branch ref name: %w", err)
	}

	cmd.Logger.Debugf("resource/in: Checkout repo at ref '%s'", refName)

	branchName := refName

	if remotes, err := repo.Remotes.List(); err == nil && len(remotes) > 0 {
		branchName = strings.TrimPrefix(branchName, remotes[0]+"/")
	}

	return &branchName, nil
}
