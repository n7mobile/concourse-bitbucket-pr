package models

import (
	"errors"
)

/*
	Version object schema
*/

// Version object used to uniquely identify an instance of the resource by ConcourseCI
type Version struct {
	Commit  string `json:"commit"`
	ID      string `json:"id"`
	Title   string `json:"title"`
	Branch  string `json:"branch"`
	Updated string `json:"updated"`
}

// Validate Version object against required fields
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

// InRequest input for In stage
type InRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
	Params  Params  `json:"params"`
}

// InResponse output for In stage
type InResponse struct {
	Version  Version  `json:"version"`
	Metadata Metadata `json:"metadata"`
}

// OutRequest input for Out stage
type OutRequest struct {
	Source Source `json:"source"`
	Params Params `json:"params"`
}

// OutResponse output for Out stage
type OutResponse struct {
	Version  Version  `json:"version"`
	Metadata Metadata `json:"metadata"`
}

// CheckRequest input for Check stage
type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

// CheckResponse ouput for Check stage
type CheckResponse []Version

/*
	Source object schema
*/

// Source object with configuration of whole resource instance
type Source struct {
	Workspace string `json:"workspace"`
	Slug      string `json:"slug"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Debug     bool   `json:"debug"`
}

// Validate Source object against required fields
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

// ParamsOutAction enumerates types of the action performed during Out stage of resource
type ParamsOutAction string

const (
	// CommitBuildStatusSetParamsOutAction updates/creates build status for HEAD of current resource version
	CommitBuildStatusSetParamsOutAction ParamsOutAction = "set:commit.build.status"
)

// Params object containing configuration of single resource invocation
type Params struct {
	VersionFilename string          `json:"version_filename"`
	VersionPath     string          `json:"version_path"`
	Action          ParamsOutAction `json:"action"`
	Status          string          `json:"status"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	URL             string          `json:"url"`
}

// Validate Params object against required fields
func (p Params) Validate() error {
	if len(p.VersionFilename) == 0 && len(p.VersionPath) == 0 {
		return errors.New("resource/model: version path or name is empty")
	}

	if len(p.Action) == 0 && len(p.VersionPath) != 0 {
		return errors.New("resource/model: action is empty or invalid")
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

// Metadata object for presenting additional info in Concourse
type Metadata []MetadataField

// MetadataField as single entity of additional info in Concourse
type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
