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
func (cts *CTS)Upload(command command.ExecRunner) error {
	command.RunExecutable("fiori", []string{"deploy"} ...)
	return nil
}