package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func cloudFoundryDeploy(config cloudFoundryDeployOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runCloudFoundryDeploy(&config, telemetryData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runCloudFoundryDeploy(config *cloudFoundryDeployOptions, telemetryData *telemetry.CustomData, command execRunner) error {
	log.Entry().Infof("General parameters: deployTool='%s', cfApiEndpoint='%s'", config.DeployTool, config.APIEndpoint)

	var err error

	if config.DeployTool == "mtaDeployPlugin" {
		err = handleMTADeployment(config)
	} else if config.DeployTool == "cf_native" {
		err = handleCFNativeDeployment(config)
	} else {
		log.Entry().Warningf("Found unsupported deployTool ('%s'). Skipping deployment. Supported deploy tools: 'mtaDeployPlugin', 'cf_native'", config.DeployTool)
	}

	return err
}

func handleMTADeployment(config *cloudFoundryDeployOptions) error {

		if !exists {
			return fmt.Errorf("mtar file '%s' retrieved from configuration does not exist", mtarFilePath)
		}

		log.Entry().Debugf("Using mtar file '%s' from configuration", mtarFilePath)
	}

	return deployMta(command)
}

func handleCFNativeDeployment(config *cloudFoundryDeployOptions, command execRunner) error {
	return nil
}

func deployMta(command execRunner) error {
	log.Entry().Info("Inside deployMta")
	return _deploy(command)
}

func _deploy(command execRunner) error {
	// TODO set HOME to config.DockerWorkspace
	var err error
	command.SetEnv([]string{"CF_TRACE=cf.log"})
	err = command.RunExecutable("cf", "login")
	if err != nil {
		return err
	}
	fmt.Printf("PATH: %v\n", path)
	return nil
}

func handleCFNativeDeployment(config *cloudFoundryDeployOptions) error {
	return nil
}
