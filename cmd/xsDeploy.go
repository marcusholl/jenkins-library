package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"os"
	"strings"
)

func xsDeploy(myXsDeployOptions xsDeployOptions) error {
	c := command.Command{}
	return runXsDeploy(myXsDeployOptions, &c)
}

func runXsDeploy(XsDeployOptions xsDeployOptions, s shellRunner) error {

	fmt.Println("[DEBUG] Inside xsDeploy")
	err := xsLogin(XsDeployOptions, s, nil)

	return err
}

func xsLogin(XsDeployOptions xsDeployOptions, s shellRunner, fExists func(string) bool) error {

	if fExists == nil {
		fExists = fileExists
	}

	loginScript := `#!/bin/bash
		xs login -a $API_URL -u $USERNAME -p '$PASSWORD' -o $ORG -s $SPACE $LOGIN_OPTS
		RC=$?
		[ $RC == 0 ]  && cp "${HOME}/$XS_SESSION_FILE" .
		exit $RC
	`

	r := strings.NewReplacer(
		"$API_URL", XsDeployOptions.APIURL,
		"$USERNAME", XsDeployOptions.User,
		"$PASSWORD", XsDeployOptions.Password,
		"$ORG", XsDeployOptions.Org,
		"$SPACE", XsDeployOptions.Space,
		"$LOGIN_OPTS", XsDeployOptions.LoginOpts,
		"$XS_SESSION_FILE", XsDeployOptions.XsSessionFile)

	loginScript = r.Replace(loginScript)

	s.RunShell("/bin/bash", loginScript)

	if !fExists(XsDeployOptions.XsSessionFile) {
		return fmt.Errorf("file does not exist (%s)", XsDeployOptions.XsSessionFile)
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
