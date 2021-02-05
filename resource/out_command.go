package resource

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/n7mobile/ci-bitbucket-pr/resource/models"
)

type OutCommand struct {
	Logger Logger
}

func (cmd *OutCommand) Run(req models.OutRequest, destination string) (models.OutResponse, error) {
	cmd.Logger.Debugf("\tReading version from %s", req.Params.VersionPath)

	version, err := cmd.readVersion(destination, req.Params.VersionPath)
	if err != nil {
		return models.OutResponse{}, fmt.Errorf("readVersion: %w", err)
	}

	return models.OutResponse{Version: *version}, nil
}

func (cmd *OutCommand) readVersion(destination, versionPath string) (*models.Version, error) {
	path := fmt.Sprintf("%s/%s", destination, versionPath)

	cmd.Logger.Debugf("\tReading from path '%s' version data", path)

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ReadFile: %w", err)
	}

	var version models.Version

	err = json.Unmarshal(bytes, &version)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal: %w", err)
	}

	return &version, nil
}
