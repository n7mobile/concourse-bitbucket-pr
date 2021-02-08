package models

import (
	"errors"
)

/*
	Version object schema
*/

type Version struct {
	Commit string `json:"commit"`
	ID     string `json:"id"`
	Title  string `json:"title"`
	Branch string `json:"branch"`
}

func (s Version) Validate() error {
	if len(s.Commit) == 0 {
		return errors.New("resource/model: commit ref is empty")
	}

	if len(s.ID) == 0 {
		return errors.New("resource/model: PR ID is empty")
	}

	if len(s.Branch) == 0 {
		return errors.New("resource/model: branch is empty")
	}

	return nil
}

/*
	In, Out, Check schema for request and response
*/

type InRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
	Params  Params  `json:"params"`
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

/*
	Source object schema
*/

type Source struct {
	Workspace string `json:"workspace"`
	Slug      string `json:"slug"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Debug     bool   `json:"debug"`
}

func (s Source) Validate() error {
	if len(s.Workspace) == 0 || len(s.Slug) == 0 {
		return errors.New("resource/model: workspace name and/or repo slug is empty")
	}

	if len(s.Username) == 0 || len(s.Username) == 0 {
		return errors.New("resource/model: basic auth is empty")
	}

	return nil
}

/*
	Params object schema
*/

type Params struct {
	VersionFilename string `json:"version_filename"`
	VersionPath     string `json:"version_path"`
	Status          string `json:"status"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	URL             string `json:"url"`
}

func (p Params) Validate() error {
	if len(p.VersionFilename) == 0 && len(p.VersionPath) == 0 {
		return errors.New("resource/model: version path or name is empty")
	}

	if len(p.Status) == 0 {
		return errors.New("resource/model: status is empty")
	}

	if len(p.VersionPath) == 0 {
		return errors.New("resource/model: repo path is empty")
	}

	if len(p.URL) == 0 {
		return errors.New("resource/model: urls is empty")
	}

	return nil
}

/*
	Metadata object schema
*/

type Metadata []MetadataField

type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
