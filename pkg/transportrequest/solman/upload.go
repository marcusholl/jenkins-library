package solman

import (
	"github.com/SAP/jenkins-library/pkg/command"
)

// SOLMANUploadAction Collects all the properties we need for the deployment
type SOLMANUploadAction struct {
	DeployUser         string
}

func (a *SOLMANUploadAction)Perform(command command.ExecRunner) error {

	err := command.RunExecutable("cmclient",
		"--endpoint", "x",//config.Endpoint,
		"--user", "x", //config.Username,
		"--password", "x", //config.Password,
		"--backend-type", "SOLMAN",
		"is-change-in-development",
		"--change-id", "x", //config.ChangeDocumentID,
		"--return-code")
	return err
}