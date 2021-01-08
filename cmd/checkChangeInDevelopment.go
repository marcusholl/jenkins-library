package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type checkChangeInDevelopmentUtils interface {
	command.ExecRunner

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The checkChangeInDevelopmentUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type checkChangeInDevelopmentUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to checkChangeInDevelopmentUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// checkChangeInDevelopmentUtilsBundle and forward to the implementation of the dependency.
}

func newCheckChangeInDevelopmentUtils() checkChangeInDevelopmentUtils {
	utils := checkChangeInDevelopmentUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func checkChangeInDevelopment(config checkChangeInDevelopmentOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newCheckChangeInDevelopmentUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runCheckChangeInDevelopment(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runCheckChangeInDevelopment(config *checkChangeInDevelopmentOptions, telemetryData *telemetry.CustomData, utils checkChangeInDevelopmentUtils) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")
	utils.RunExecutable("cmclient",
		"--endpoint", config.Endpoint,
		"--user", config.Username,
		"--password", config.Password,
		"--backend-type", "SOLMAN",
		"is-change-in-development",
		"--change-id", config.ChangeDocumentID,
		"--return-code")
	return nil
}
