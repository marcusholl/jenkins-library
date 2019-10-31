package cmd

import (
	"strings"
	"fmt"
	"errors"
	"github.com/SAP/jenkins-library/pkg/command"
	"os"
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

	r := strings.NewReplacer("$API_URL", "https://example.org",
							"$USERNAME", "me",
							"$PASSWORD", "secret",
							"$ORG", "myOrg",
							"$SPACE", "mySpace",
							"$LOGIN_OPTS", "--skip-ssl-validation",
							"$XS_SESSION_FILE", ".xssession")

	loginScript = r.Replace(loginScript)


	s.RunShell("/bin/bash", loginScript)

	if ! fExists("xyz") {
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