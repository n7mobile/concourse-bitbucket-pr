package concourse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// Storage read and write for JSON data.
// Utility used to pass data from in to out step by filesystem
type Storage struct {
	path string
}

// NewStorage with given filename in given directory path
func NewStorage(dir, name string) *Storage {
	return &Storage{
		path: filepath.Join(dir, name),
	}
}

// StorageFilename enumerates constant names used in system
type StorageFilename string

const (
	// VersionStorageFilename points concourse Version object
	VersionStorageFilename StorageFilename = ".concourse.version.json"
)

// Write obj marshaled into JSON in file with given name in given directory
func (s Storage) Write(obj interface{}) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("concourse/storage: json marshal: %w", err)
	}

	err = ioutil.WriteFile(s.path, bytes, 0644)
	if err != nil {
		return fmt.Errorf("concourse/storage: write file: %w", err)
	}

	return nil
}

// Read file with given name in given dir and unmarshal it into obj
func (s Storage) Read(obj interface{}) error {
	bytes, err := ioutil.ReadFile(s.path)
	if err != nil {
		return fmt.Errorf("concourse/storage: read file: %w", err)
	}

	err = json.Unmarshal(bytes, obj)
	if err != nil {
		return fmt.Errorf("concourse/storage: json unmarshal: %w", err)
	}

	return nil
}
