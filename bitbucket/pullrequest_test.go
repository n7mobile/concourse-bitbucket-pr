package bitbucket

import (
	"log"
	"testing"
)

func TestGetPullRequests(t *testing.T) {
	cli := NewBasicAuth("", "")

	qry := &PullRequestQuery{
		Workspace: "n7mobile",
		Slug:      "pr-test-repo",
	}

	prs, err := cli.GetPullRequests(qry)
	if err != nil {
		t.Error(err)
	}

	for _, pr := range prs {
		log.Println(pr)
	}
}
