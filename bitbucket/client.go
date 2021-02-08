package bitbucket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type endpoint string

const (
	pullRequestsEndpoint endpoint = "/pullrequests"
	commitEndpoint       endpoint = "/commit"
	commitsEndpoint      endpoint = "/commits"
)

const (
	apiBaseURL  string = "https://api.bitbucket.org/2.0/repositories"
	repoBaseURL string = "https://bitbucket.org"
)

type Client struct {
	auth       *Auth
	repoPath   string
	httpClient *http.Client
}

type Auth struct {
	Username string
	Password string
}

type PagedResponse struct {
	Size   int             `json:"size"`
	Next   string          `json:"next"`
	Values json.RawMessage `json:"values"`
}

func NewClient(workspace, slug string, auth *Auth) *Client {
	return &Client{
		auth:       auth,
		repoPath:   fmt.Sprintf("/%s/%s", workspace, slug),
		httpClient: &http.Client{},
	}
}

func (c Client) APIURL(endpoint endpoint, components ...string) string {
	url := apiBaseURL + c.repoPath + string(endpoint)

	if len(components) > 0 {
		url = url + "/" + strings.Join(components, "/")
	}

	return url
}

func (c Client) RepoURL() string {
	return repoBaseURL + c.repoPath + ".git"
}
