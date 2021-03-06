package resource

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	git "github.com/libgit2/git2go/v31"
	"github.com/n7mobile/concourse-bitbucket-pr/bitbucket"
	"github.com/n7mobile/concourse-bitbucket-pr/concourse"
	"github.com/n7mobile/concourse-bitbucket-pr/resource/models"
)

// OutCommand sets PullRequest status and metadata to passed one in OutRequest struct
type OutCommand struct {
	Logger *concourse.Logger
}

// Run OutCommand processing.
// Full SHA1 is fetched from HEAD of previous checkout step
func (cmd *OutCommand) Run(req models.OutRequest, destination string) (*models.OutResponse, error) {
	err := req.Source.Validate()
	if err != nil {
		return nil, fmt.Errorf("resource/out: source invalid: %w", err)
	}

	err = req.Params.Validate()
	if err != nil {
		return nil, fmt.Errorf("resource/out: params invalid: %w", err)
	}

	versionPath := filepath.Join(req.Params.RepoPath, string(concourse.VersionStorageFilename))

	cmd.Logger.Debugf("resource/out: reading version from %s", versionPath)

	var version models.Version

	err = concourse.NewStorage(destination, versionPath).Read(&version)
	if err != nil {
		return nil, fmt.Errorf("resource/out: version read: %w", err)
	}

	cmd.Logger.Debugf("resource/out: version with commit %s, id %s", version.Ref, version.ID)

	path := filepath.Join(destination, req.Params.RepoPath)

	hash, err := cmd.gitGetHeadHash(path)
	if err != nil {
		return nil, fmt.Errorf("resource/out: get git head: %w", err)
	}

	cmd.Logger.Debugf("resource/out: got commit SHA1: %s", hash)

	auth := bitbucket.Auth{
		Username: req.Source.Username,
		Password: req.Source.Password,
	}

	client := bitbucket.NewClient(req.Source.Workspace, req.Source.Slug, &auth)

	statReq := bitbucket.CommitBuildStatusRequest{
		Key:         substituteEnvs(req.Params.Key),
		Name:        substituteEnvs(req.Params.Name),
		Description: substituteEnvs(req.Params.Description),
		URL:         substituteEnvs(req.Params.URL),
		State:       bitbucket.CommitBuildStatus(req.Params.Status),
	}

	cmd.Logger.Debugf("resource/out: set status %s", statReq.State)

	err = client.SetCommitBuildStatus(hash, &statReq)
	if err != nil {
		return nil, fmt.Errorf("resource/out: set build status %w", err)
	}

	return &models.OutResponse{
		Version: version,
	}, nil
}

func (cmd *OutCommand) gitGetHeadHash(path string) (string, error) {
	repo, err := git.OpenRepository(path)
	if err != nil {
		return "", fmt.Errorf("resource/out: open repo at path %s: %w", path, err)
	}

	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("resource/out: get head: %w", err)
	}

	return head.Target().String(), nil
}

func substituteEnvs(s string) string {
	for _, item := range os.Environ() {
		kv := strings.Split(item, "=")
		s = strings.ReplaceAll(s, "$"+kv[0], kv[1])
	}

	return s
}
