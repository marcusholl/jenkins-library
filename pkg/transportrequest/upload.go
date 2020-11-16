package transportrequest

import (
	"github.com/SAP/jenkins-library/pkg/command"
)

// CTS ...
type CTS struct {
	endpoint string
	client string
	user string
	password string
}

// Upload ...
func (cts *CTS)Upload(command command.ExecRunner, transportRequestID string, abapPackage string, applicationName string) error {


	params := []string{
		"deploy",
		"-f", // failfast --> provide return code != 0 in case of any failure
        "-y", // autoconfirm --> no need to press 'y' key in order to confirm the params and trigger the deployment
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
	if len(abapPackage) > 0 {
		params = append(params, "-p", abapPackage)
	}
	if len(applicationName) > 0 {
		params = append(params, "-n", applicationName)
	}

	command.RunExecutable("fiori", params ...)
	return nil
}