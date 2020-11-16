package transportrequest

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type fileUtils interface {
	FileExists(string) (bool, error)
}

var files fileUtils = piperutils.Files{}

// CTS ...
type CTS struct {
	endpoint string
	client   string
	user     string
	password string
}

// CTSApp ...
type CTSApp struct {
	name string
	pack string
	desc string
}

// Upload ...
func (cts *CTS) Upload(command command.ExecRunner, transportRequestID string, configFile string, app CTSApp) error {

	desc := app.desc
	if len(desc) == 0 {
		desc = "Deployed with Piper based on SAP Fiori tools"
	}

	useConfigFile, noConfig, err := handleConfigFile(configFile)
	if err != nil {
		return err
	}

	params := []string{
		"deploy",
		"-f", // failfast --> provide return code != 0 in case of any failure
		"-y", // autoconfirm --> no need to press 'y' key in order to confirm the params and trigger the deployment
		"-e", desc,
	}

	if noConfig {
		params = append(params, "--noConfig") // no config file, but we will provide our parameters
	}
	if useConfigFile {
		params = append(params, "-c", configFile)
	}
	if len(cts.endpoint) > 0 {
		params = append(params, "-u", cts.endpoint)
	}
	if len(cts.client) > 0 {
		params = append(params, "-l", cts.client)
	}
	if len(transportRequestID) > 0 {
		params = append(params, "-t", transportRequestID)
	}
	if len(app.pack) > 0 {
		params = append(params, "-p", app.pack)
	}
	if len(app.name) > 0 {
		params = append(params, "-n", app.name)
	}

	return command.RunExecutable("fiori", params...)
}

func handleConfigFile(path string) (bool, bool, error) {

	useConfigFile := true
	noConfig := false

	if len(path) == 0 {
		useConfigFile = false
		exists, err := files.FileExists("ui5-deploy.yaml")
		if err != nil {
			return false, false, err
		}
		noConfig = !exists
	} else {
		exists, err := files.FileExists(path)
		if err != nil {
			return false, false, err
		}
		if exists {
			useConfigFile = true
			noConfig = false
		} else {
			if path == "ui5-deploy.yaml" {
				useConfigFile = false
				noConfig = true
			} else {
				err = fmt.Errorf("Configured deploy config file '%s' does not exists", path)
				return false, false, err
			}
		}
	}
	return useConfigFile, noConfig, nil
}