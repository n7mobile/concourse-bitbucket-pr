package models

import (
	"errors"
)

type Version struct {
	Commit string `json:"commit"`
	ID     string `json:"id"`
	Title  string `json:"title"`
	Branch string `json:"branch"`
}

func (s Version) Validate() error {
	if len(s.Commit) == 0 {
		return errors.New("commit reference is empty")
	}

	if len(s.ID) == 0 {
		return errors.New("pullrequest identifier is empty")
	}

	if len(s.Branch) == 0 {
		return errors.New("branch namse is empty")
	}

	return nil
}

type InRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type InResponse struct {
	Version  Version  `json:"version"`
	Metadata Metadata `json:"metadata"`
}

type OutRequest struct {
	Source Source `json:"source"`
	Params Params `json:"params"`
}

type OutResponse struct {
	Version  Version  `json:"version"`
	Metadata Metadata `json:"metadata"`
}

type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type CheckResponse []Version

type Source struct {
	Workspace string `json:"workspace"`
	Slug      string `json:"slug"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Debug     bool   `json:"debug"`
}

type Params struct {
	VersionPath string `json:"version_path"`
	Status      string `json:"status"`
}

func (s Source) Validate() error {
	if len(s.Workspace) == 0 || len(s.Slug) == 0 {
		return errors.New("BitBucket workspace name and repo slug has to be set")
	}

	if len(s.Username) == 0 || len(s.Username) == 0 {
		return errors.New("basic auth credentails for BitBucket has to be set")
	}

	return nil
}

type Metadata []MetadataField

type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
