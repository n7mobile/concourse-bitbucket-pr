package bitbucket

import (
	"fmt"
	"net/http"
)

type endpoint string

const (
	pullRequestsEndpoint endpoint = "/pullrequests"
)

const (
	apiBaseUrl  string = "https://api.bitbucket.org/2.0/repositories"
	repoBaseUrl string = "https://bitbucket.org"
)

type Client struct {
	username   string
	password   string
	httpClient *http.Client
}

func NewBasicAuth(username, password string) *Client {
	return &Client{username, password, &http.Client{}}
}

func (c Client) apiURL(workspace, slug string, endpoint endpoint) string {
	return fmt.Sprintf("%s/%s/%s%v", apiBaseUrl, workspace, slug, endpoint)
}

func RepositoryURL(workspace, slug string) string {
	return fmt.Sprintf("%s/%s/%s.git", repoBaseUrl, workspace, slug)
}
