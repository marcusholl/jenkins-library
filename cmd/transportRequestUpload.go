package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/transportrequest"
)

type transportRequestUploadUtils interface {
	command.ShellRunner
	GetAction() *transportrequest.CTSUploadAction

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The transportRequestUploadUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

// ActionProvider ...
type ActionProvider struct {
	action transportrequest.CTSUploadAction
}

// GetAction ...
func (provider *ActionProvider) GetAction() *transportrequest.CTSUploadAction {
	return &provider.action
}

type transportRequestUploadUtilsBundle struct {
	*command.Command
	*ActionProvider

	// Embed more structs as necessary to implement methods or interfaces you add to transportRequestUploadUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// transportRequestUploadUtilsBundle and forward to the implementation of the dependency.
}

func newTransportRequestUploadUtils() transportRequestUploadUtils {
	utils := transportRequestUploadUtilsBundle{
		Command:        &command.Command{},
		ActionProvider: &ActionProvider{action: transportrequest.CTSUploadAction{}},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func transportRequestUpload(config transportRequestUploadOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newTransportRequestUploadUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runTransportRequestUpload(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runTransportRequestUpload(config *transportRequestUploadOptions, telemetryData *telemetry.CustomData, utils transportRequestUploadUtils) error {

	log.Entry().Debugf("Entering 'runTransportRequestUpload' with config: %v", config)

	action := utils.GetAction()

	action.Connection = transportrequest.CTSConnection{
		Endpoint: config.Endpoint,
		Client:   config.Client,
		User:     config.Username,
		Password: config.Password,
	}
	action.Application = transportrequest.CTSApplication{
		Name: config.ApplicationName,
		Pack: config.AbapPackage,
		Desc: config.Description,
	}
	action.Node = transportrequest.CTSNode{
		DeployDependencies: config.DeployToolDependencies,
		InstallOpts:        config.NpmInstallOpts,
	}

	action.TransportRequestID = config.TransportRequestID
	action.ConfigFile = config.DeployConfigFile
	action.DeployUser = config.OsDeployUser

	return action.Perform(utils)
}
