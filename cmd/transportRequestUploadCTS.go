package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	transportrequest "github.com/SAP/jenkins-library/pkg/transportrequest/cts"
)

type transportRequestUploadUtils interface {
	command.ShellRunner

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The transportRequestUploadUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

// CTSUploadAction ...
type CTSUploadAction interface {
	Perform(command.ShellRunner) error
	WithConnection(transportrequest.CTSConnection)
	WithApplication(transportrequest.CTSApplication)
	WithNodeProperties(transportrequest.CTSNode)
	WithTransportRequestID(string)
	WithConfigFile(string)
	WithDeployUser(string)
}

type transportRequestUploadCTSUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to transportRequestUploadUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// transportRequestUploadUtilsBundle and forward to the implementation of the dependency.
}

func newTransportRequestUploadCTSUtils() transportRequestUploadUtils {
	utils := transportRequestUploadCTSUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func transportRequestUploadCTS(config transportRequestUploadCTSOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newTransportRequestUploadCTSUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runTransportRequestUploadCTS(&config, &transportrequest.CTSUploadAction{}, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runTransportRequestUploadCTS(
	config *transportRequestUploadCTSOptions,
	action CTSUploadAction,
	telemetryData *telemetry.CustomData,
	cmd command.ShellRunner) error {

	log.Entry().Debugf("Entering 'runTransportRequestUpload' with config: %v", config)

	action.WithConnection(transportrequest.CTSConnection{
		Endpoint: config.Endpoint,
		Client:   config.Client,
		User:     config.Username,
		Password: config.Password,
	})
	action.WithApplication(transportrequest.CTSApplication{
		Name: config.ApplicationName,
		Pack: config.AbapPackage,
		Desc: config.Description,
	})
	action.WithNodeProperties(transportrequest.CTSNode{
		DeployDependencies: config.DeployToolDependencies,
		InstallOpts:        config.NpmInstallOpts,
	})

	action.WithTransportRequestID(config.TransportRequestID)
	action.WithConfigFile(config.DeployConfigFile)
	action.WithDeployUser(config.OsDeployUser)

	return action.Perform(cmd)
}
