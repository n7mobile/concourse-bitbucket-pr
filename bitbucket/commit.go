package bitbucket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type CommitReponse struct {
	Hash    string `json:"hash"`
	Date    string `json:"date"`
	Message string `json:"message"`
}

type CommitBuildStatus string

const (
	SuccessfullCommitBuildStatus CommitBuildStatus = "SUCCESSFUL"
	FailedCommitBuildStatus      CommitBuildStatus = "FAILED"
	InProgressCommitBuildStatus  CommitBuildStatus = "INPROGRESS"
	StoppedCommitBuildStatus     CommitBuildStatus = "STOPPED"
)

type CommitBuildStatusRequest struct {
	Key         string            `json:"key"`
	State       CommitBuildStatus `json:"state"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	URL         string            `json:"url"`
}

func (c Client) GetCommits(branch string) ([]CommitReponse, error) {
	url := c.APIURL(commitsEndpoint, branch)

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

	var pagedResponse PagedResponse

	err = json.NewDecoder(buf).Decode(&pagedResponse)
	if err != nil {
		return nil, fmt.Errorf("bitbucket/client: decode paged %s: %w", buf.Bytes(), err)
	}

	var commitsPage []CommitReponse

	err = json.Unmarshal([]byte(pagedResponse.Values), &commitsPage)
	if err != nil {
		return nil, fmt.Errorf("bitbucket/client: decode values %s: %w", pagedResponse.Values, err)
	}

	return commitsPage, nil
}

func (c Client) SetCommitBuildStatus(commitHash string, statReq *CommitBuildStatusRequest) error {
	url := c.APIURL(commitEndpoint, commitHash, "statuses", "build")

	data := new(bytes.Buffer)
	err := json.NewEncoder(data).Encode(statReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.auth.Username, c.auth.Password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		resData, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("setting commit build status failed with status %s and response \n%s", res.Status, resData)
	}

	return nil
}
