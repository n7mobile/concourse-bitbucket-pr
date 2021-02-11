package bitbucket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PullRequestState indicatees state of PullRequest
// in contrast to state of single commit and it's CI build flow
type PullRequestState string

const (
	// OpenPullRequestState as PR is open
	OpenPullRequestState PullRequestState = "OPEN"

	// MergedPullRequestState as PR is already merged
	MergedPullRequestState PullRequestState = "MERGED"

	// SupersededPullRequestState as PR is superseded
	SupersededPullRequestState PullRequestState = "SUPERSEDED"

	// DeclinedPullRequestState as PR is declined
	DeclinedPullRequestState PullRequestState = "DECLINED"
)

type PullRequestEntity struct {
	ID              int              `json:"id"`
	Title           string           `json:"title"`
	State           PullRequestState `json:"state"`
	CloseAfterMerge bool             `json:"close_source_branch"`
	Author          GitAuthor        `json:"author"`
	Source          GitReference     `json:"source"`
	Dest            GitReference     `json:"destination"`
	UpdatedOn       string           `json:"updated_on"`
	CreatedOn       string           `json:"created_on"`
}

type GitAuthor struct {
	Name string `json:"display_name"`
}

type GitReference struct {
	Commit     GitCommit     `json:"commit"`
	Repository GitRepository `json:"repository"`
	Branch     GitBranch     `json:"branch"`
}

type GitCommit struct {
	Hash string `json:"hash"`
	Type string `json:"type"`
}

type GitRepository struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

type GitBranch struct {
	Name string `json:"name"`
}

// GetPullRequestsPaged fetches list of PR for given repository.
// Results are autopaged
func (c Client) GetPullRequestsPaged() ([]PullRequestEntity, error) {
	values := make([]PullRequestEntity, 0)

	url := c.APIURL(pullRequestsEndpoint)
	url = fmt.Sprintf("%s?pagelen=%d", url, 50)

	for ok := true; ok; ok = len(url) > 0 {
		resp, err := c.getPullRequestsSinglePage(url)
		if err != nil {
			return nil, err
		}

		var valuesPage []PullRequestEntity

		err = json.Unmarshal([]byte(resp.Values), &valuesPage)
		if err != nil {
			return nil, fmt.Errorf("bitbucket/client: unmarshal paged values: %w", err)
		}

		values = append(values, valuesPage...)
		url = resp.Next
	}

	return values, nil
}

func (c Client) getPullRequestsSinglePage(url string) (*PagedResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("bitbucket/client: http req creation: %w", err)
	}

	req.SetBasicAuth(c.auth.Username, c.auth.Password)
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bitbucket/client: http req to %s: %w", url, err)
	}
	defer res.Body.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return nil, fmt.Errorf("bitbucket/client: read body: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("bitbucket/client: %s: %s, %s", res.Status, url, buf.Bytes())
	}

	var pullRequestResponse PagedResponse

	err = json.NewDecoder(buf).Decode(&pullRequestResponse)
	if err != nil {
		return nil, fmt.Errorf("bitbucket/client: decode %s: %w", buf.Bytes(), err)
	}

	return &pullRequestResponse, nil
}
