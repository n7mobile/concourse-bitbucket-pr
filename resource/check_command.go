package resource

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	git "github.com/libgit2/git2go/v31"
	"github.com/n7mobile/concourse-bitbucket-pr/bitbucket"
	"github.com/n7mobile/concourse-bitbucket-pr/concourse"
	"github.com/n7mobile/concourse-bitbucket-pr/resource/models"
)

// CheckCommand fetches list of PullRequests form BitBucket API and traslates it list versions sorted by PR Identifier
type CheckCommand struct {
	Logger *concourse.Logger
}

type commitAttr struct {
	commit      *git.Commit
	pullRequest bitbucket.PullRequestEntity
}

type sortByCommitDate []commitAttr

func (a sortByCommitDate) Len() int      { return len(a) }
func (a sortByCommitDate) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a sortByCommitDate) Less(i, j int) bool {
	timeI := a[i].commit.Committer().When
	timeJ := a[j].commit.Committer().When
	return timeI.Before(timeJ)
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

	commits := []commitAttr{}

	for _, pr := range preqs {
		if commit, err := cmd.getCommit(repo, pr.Source.Commit.Hash); err == nil {
			commits = append(commits, commitAttr{
				commit:      commit,
				pullRequest: pr,
			})
		} else {
			cmd.Logger.Errorf("resource/check: commit %s not found: %w", pr.Source.Commit.Hash, err)
		}
	}

	sort.Sort(sortByCommitDate(commits))

	versions := []models.Version{}
	hasVersion := false

	for _, c := range commits {
		ref := c.commit.AsObject().Id().String()
		id := strconv.Itoa(c.pullRequest.ID)

		versions = append(versions, models.Version{
			Ref: ref,
			ID:  id,
		})

		hasVersion = hasVersion || strings.HasPrefix(ref, req.Version.Ref)
		cmd.Logger.Debugf("resource/check: append version (%s, %s)", id, ref)
	}

	if !hasVersion && req.Version.Validate() == nil {
		versions = append([]models.Version{req.Version}, versions...)
		cmd.Logger.Debugf("resource/check: passed version (%s, %s) valid but not present in git. Prepending", req.Version.ID, req.Version.Ref)
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

	cmd.Logger.Debugf("resource/check: clone history from repo %s", url)

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
	refObj, err := repo.RevparseSingle(ref)
	if err != nil {
		return nil, fmt.Errorf("resource/in: RevparseSingle: %w", err)
	}

	commit, err := repo.LookupCommit(refObj.Id())
	if err != nil {
		return nil, fmt.Errorf("resource/in: LookupCommit: %w", err)
	}

	return commit, nil
}
