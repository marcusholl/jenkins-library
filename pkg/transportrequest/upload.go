package transportrequest

import (
	"github.com/SAP/jenkins-library/pkg/command"
)

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
func (cts *CTS) Upload(command command.ExecRunner, transportRequestID string, app CTSApp) error {

	desc := app.desc
	if len(desc) == 0 {
		desc = "Deployed with Piper based on SAP Fiori tools"
	}

	params := []string{
		"deploy",
		"-f", // failfast --> provide return code != 0 in case of any failure
		"-y", // autoconfirm --> no need to press 'y' key in order to confirm the params and trigger the deployment
		"-e", desc,
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

	command.RunExecutable("fiori", params...)
	return nil
}
