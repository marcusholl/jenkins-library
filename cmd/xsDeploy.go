package cmd

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
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

	err := xsLogin(XsDeployOptions, s, nil, nil)

	if err == nil {
		err = xsLogout(XsDeployOptions, s, nil, nil, nil)
	}

	pwOut.Close()
	pwErr.Close()

	wg.Wait()

	fmt.Printf("STDOUT: %v\n", o)
	fmt.Printf("STDERR: %v\n", e)

	return err
}

func xsLogin(XsDeployOptions xsDeployOptions, s shellRunner,
	fExists func(string) bool,
	fCopy func(string, string) (int64, error)) error {

	log.Entry().Debugf("Performing xs login. api-url: '%s', org: '%s', space: '%s'",
		XsDeployOptions.APIURL, XsDeployOptions.Org, XsDeployOptions.Space)

	if fExists == nil {
		fExists = piperutils.FileExists
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

	src, dest := fmt.Sprintf("%s/%s", os.Getenv("HOME"), xsSessionFile), fmt.Sprintf("./%s", xsSessionFile)
	if _, err := fCopy(src, dest); err != nil {
		return  errors.Wrapf(err, "Cannot copy xssession file from '%s' to '%s'", src, dest)
	}

	log.Entry().Debugf("xs session file copied from '%s' to '%s'", src, dest)

	log.Entry().Infof("xs login has been performed. api-url: '%s', org: '%s', space: '%s'",
		XsDeployOptions.APIURL, XsDeployOptions.Org, XsDeployOptions.Space)

	return nil
}

func xsLogout(XsDeployOptions xsDeployOptions, s shellRunner,
	fExists func(string) bool,
	fCopy func(string, string) (int64, error),
	fRemove func(string) error) error {

	log.Entry().Debug("Performing xs logout.")

	xsSessionFile := ".xsconfig"
	if len(XsDeployOptions.XsSessionFile) > 0 {
		xsSessionFile = XsDeployOptions.XsSessionFile
	}

	if fRemove == nil {
		fRemove = os.Remove
	}

	if fCopy == nil {
		fCopy =  piperutils.Copy
	}

	if fExists == nil {
		fExists = piperutils.FileExists
	}

	if !fExists(xsSessionFile) {
		return fmt.Errorf("xs session file does not exist (%s)", xsSessionFile)
	}

	logoutScript := `#!/bin/bash
	cp $XS_SESSION_FILE ${HOME}
	xs logout`

	r := strings.NewReplacer(
		"$XS_SESSION_FILE", xsSessionFile)

	logoutScript = r.Replace(logoutScript)

	if e := s.RunShell("/bin/bash", logoutScript); e != nil {
		return e
	}

	if e := fRemove(xsSessionFile); e != nil {
		return e
	}

	log.Entry().Info("xs logout has been performed")

	return nil
}

