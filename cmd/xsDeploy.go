package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
)


func xsDeploy(myXsDeployOptions xsDeployOptions) error {
	c := command.Command{}
	return runXsDeploy(myXsDeployOptions, &c)
}

func runXsDeploy(XsDeployOptions xsDeployOptions, command execRunner) error {

	fmt.Println("[DEBUG] Inside xsDeploy")

	return nil
}
