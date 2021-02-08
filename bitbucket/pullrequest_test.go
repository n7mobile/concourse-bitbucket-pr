package bitbucket

import (
	"log"
	"testing"
)

func TestGetPullRequests(t *testing.T) {
	auth := Auth{
		Username: "",
		Password: "",
	}

	cli := NewClient("n7mobile", "n7mobile", &auth)

	prs, err := cli.GetPullRequestsPaged()
	if err != nil {
		t.Error(err)
	}

	for _, pr := range prs {
		log.Println(pr)
	}
}
