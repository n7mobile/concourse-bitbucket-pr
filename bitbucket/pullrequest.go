package bitbucket

import (
	"encoding/json"
	"net/http"
)

type PullRequestQuery struct {
	Workspace string `json:"workspace"`
	Slug      string `json:"slug"`
}

type PullRequestState string

const (
	OpenPullRequestState       PullRequestState = "OPEN"
	MergedPullRequestState     PullRequestState = "MERGED"
	SupersededPullRequestState PullRequestState = "SUPERSEDED"
	DeclinedPullRequestState   PullRequestState = "DECLINED"
)

type PullRequestResponse struct {
	Size   int                 `json:"size"`
	Next   string              `json:"next"`
	Values []PullRequestEntity `json:"values"`
}

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

// GetPullRequests fetches list of PR for given repository.
// Results are autopaged
func (c Client) GetPullRequests(q *PullRequestQuery) ([]PullRequestEntity, error) {
	values := make([]PullRequestEntity, 0)

	url := c.apiURL(q.Workspace, q.Slug, pullRequestsEndpoint)

	for ok := true; ok; ok = len(url) > 0 {
		resp, err := c.getPullRequests(url)
		if err != nil {
			return nil, err
		}

		values = append(values, resp.Values...)
		url = resp.Next
	}

	return values, nil
}

func (c Client) getPullRequests(url string) (*PullRequestResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var pullRequestResponse PullRequestResponse

	err = json.NewDecoder(res.Body).Decode(&pullRequestResponse)
	if err != nil {
		return nil, err
	}

	return &pullRequestResponse, nil
}
