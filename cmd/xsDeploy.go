package cmd

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"io"
	"os"
	"strings"
	"sync"
)

func xsDeploy(myXsDeployOptions xsDeployOptions) error {
	c := command.Command{}
	return runXsDeploy(myXsDeployOptions, &c)
}

func runXsDeploy(XsDeployOptions xsDeployOptions, s shellRunner) error {

	prOut, pwOut := io.Pipe()
	prErr, pwErr := io.Pipe()

	s.Stdout(pwOut)
	s.Stderr(pwErr)

	var e, o string

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		buf := new(bytes.Buffer)
		io.Copy(buf, prOut)
		o = buf.String()
		wg.Done()
	}()

	go func() {
		buf := new(bytes.Buffer)
		io.Copy(buf, prErr)
		e = buf.String()
		wg.Done()
	}()

	err := xsLogin(XsDeployOptions, s, nil)
	pwOut.Close()
	pwErr.Close()

	wg.Wait()

	fmt.Printf("STDOUT: %v\n", o)
	fmt.Printf("STDERR: %v\n", e)

	return err
}

func xsLogin(XsDeployOptions xsDeployOptions, s shellRunner, fExists func(string) bool) error {

	log.Entry().Debugf("Performing xs login. api-url: '%s', org: '%s', space: '%s'",
		XsDeployOptions.APIURL, XsDeployOptions.Org, XsDeployOptions.Space)


	if fExists == nil {
		fExists = fileExists
	}

	xsSessionFile := ".xsconfig"
	if len(XsDeployOptions.XsSessionFile) > 0 {
		xsSessionFile = XsDeployOptions.XsSessionFile
	}

	loginScript := `#!/bin/bash
		xs login -a $API_URL -u $USERNAME -p '$PASSWORD' -o $ORG -s $SPACE $LOGIN_OPTS
		RC=$?
		[[ $RC == 0 && -f "${HOME}/$XS_SESSION_FILE" ]]  && cp "${HOME}/$XS_SESSION_FILE" .
		exit $RC
	`

	r := strings.NewReplacer(
		"$API_URL", XsDeployOptions.APIURL,
		"$USERNAME", XsDeployOptions.User,
		"$PASSWORD", XsDeployOptions.Password,
		"$ORG", XsDeployOptions.Org,
		"$SPACE", XsDeployOptions.Space,
		"$LOGIN_OPTS", XsDeployOptions.LoginOpts,
		"$XS_SESSION_FILE", xsSessionFile)

	loginScript = r.Replace(loginScript)

	if e := s.RunShell("/bin/bash", loginScript); e != nil {
		return e
	}

	if !fExists(xsSessionFile) {
		return fmt.Errorf("xs session file does not exist (%s)", xsSessionFile)
	}

	log.Entry().Infof("xs login has been performed. api-url: '%s', org: '%s', space: '%s'",
		XsDeployOptions.APIURL, XsDeployOptions.Org, XsDeployOptions.Space)

	return nil
}

func fileExists(filename string) bool {
	f, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !f.IsDir()
}
