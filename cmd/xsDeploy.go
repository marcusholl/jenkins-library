package cmd

import (
	"fmt"
	"errors"
	"github.com/SAP/jenkins-library/pkg/command"
	"os"
)


func xsDeploy(myXsDeployOptions xsDeployOptions) error {
	c := command.Command{}
	return runXsDeploy(myXsDeployOptions, &c)
}

func runXsDeploy(XsDeployOptions xsDeployOptions, command execRunner) error {

	fmt.Println("[DEBUG] Inside xsDeploy")
	err := login()

	return err
}

func login() error {
	fmt.Println("[DEBUG] inside login")
	c := command.Command{}
	c.RunShell("/bin/bash", "echo Hello && touch .xyz")

	if ! fileExists(".xxyz") {
		return errors.New("File does not exist")
	}

	return nil
}

func fileExists(filename string) bool {
    f, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !f.IsDir()
}