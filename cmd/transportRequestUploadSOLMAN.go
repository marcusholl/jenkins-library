package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	transportrequest "github.com/SAP/jenkins-library/pkg/transportrequest/solman"
)

type transportRequestUploadSOLMANUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	GetExitCode() int

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The transportRequestUploadSOLMANUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type transportRequestUploadSOLMANUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to transportRequestUploadSOLMANUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// transportRequestUploadSOLMANUtilsBundle and forward to the implementation of the dependency.
}

func newTransportRequestUploadSOLMANUtils() transportRequestUploadSOLMANUtils {
	utils := transportRequestUploadSOLMANUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func transportRequestUploadSOLMAN(config transportRequestUploadSOLMANOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newTransportRequestUploadSOLMANUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runTransportRequestUploadSOLMAN(&config, &transportrequest.SOLMANUploadAction{}, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runTransportRequestUploadSOLMAN(config *transportRequestUploadSOLMANOptions, action transportrequest.Action, telemetryData *telemetry.CustomData, utils transportRequestUploadSOLMANUtils) error {
	action.WithConnection(transportrequest.SOLMANConnection {
		Endpoint: config.Endpoint,
		User:	config.Username,
		Password: config.Password,
	})
	action.WithChangeDocumentId(config.ChangeDocumentID)
	action.WithTransportRequestId(config.TransportRequestID)
	action.WithApplicationID(config.ApplicationID)
	action.WithFile(config.FilePath)
	action.WithCMOpts(config.Cmclientops)
	return action.Perform(utils, utils)
}
