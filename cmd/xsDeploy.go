package cmd

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"
	"sync"
)

//
// START DeployMode
type DeployMode int

const (
	//UnknownMode ...
	UnknownMode = iota
	// NoDeploy ...
	NoDeploy DeployMode = iota
	//Deploy ...
	Deploy DeployMode = iota
	//BGDeploy ...
	BGDeploy DeployMode = iota
)

//ValueOfMode ...
func ValueOfMode(str string) (DeployMode, error) {
	switch str {
	case "UnknownMode":
		return UnknownMode, nil
	case "NoDeploy":
		return NoDeploy, nil
	case "Deploy":
		return Deploy, nil
	case "BGDeploy":
		return BGDeploy, nil
	default:
		return UnknownMode, errors.New(fmt.Sprintf("Unknown DeployMode: '%s'", str))
	}
}

// String
func (m DeployMode) String() string {
	return [...]string{
			"UnknownMode",
			"None",
			"Deploy",
			"BGDeploy",
	}[m]
}

// END DeployMode
//

//
// START Action
type Action int

const (
	//None ...
	None Action = iota
	//Resume ...
	Resume Action = iota
	//Abort ...
	Abort Action = iota
	//Retry ...
	Retry Action = iota
)

//ValueOfAction ...
func ValueOfAction(str string) (Action, error) {
	switch str {
	case "None":
		return None, nil
	case "Resume":
		return Resume, nil
	case "Abort":
		return Abort, nil
	case "Retry":
		return Retry, nil

	default:
		return None, errors.New(fmt.Sprintf("Unknown Action: '%s'", str))
	}
}

// String
func (a Action) String() string {
	return [...]string{
			"None",
			"Resume",
			"Abort",
			"Retry",
	}[a]
}

// END Action
//

const loginScript = `#!/bin/bash
xs login -a $API_URL -u $USERNAME -p '$PASSWORD' -o $ORG -s $SPACE $LOGIN_OPTS
`

const logoutScript = `#!/bin/bash
cp $XS_SESSION_FILE ${HOME}
xs logout`

const deployScript = `#!/bin/bash
xs $d $MTA_PATH $DEPLOY_OPTS`


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

	mode, err := ValueOfMode(XsDeployOptions.Mode)
	if err != nil {
		fmt.Printf("Extracting mode failed: %v\n", err)
		return err
	}

	if mode == NoDeploy {
		log.Entry().Infof("Deployment skipped intentionally. Deploy mode '%s'", mode.String())
		return nil
	}

	action, err := ValueOfAction(XsDeployOptions.Action)
	if err != nil {
		fmt.Printf("Extracting action failed: %v\n", err)
		return err
	}

	if mode == Deploy && action != None {
		return errors.New(fmt.Sprintf("Cannot perform action '%s' in mode '%s'. Only action '%s' is allowed.", action, mode, None))
	}

	log.Entry().Debugf("Mode: '%s', Action: '%s'", mode, action)

	performLogin  := mode == Deploy || (mode == BGDeploy && ! (action == Resume || action == Abort))
	performLogout := mode == Deploy || (mode == BGDeploy && action != None)
	log.Entry().Debugf("performLogin: %t, performLogout: %t", performLogin, performLogout)

	// TODO: check: for action NONE --> deployable must exist.
	// Should be done before even trying to login

	if performLogin {
		if err = xsLogin(XsDeployOptions, s, nil, nil); err != nil {
			return err
		}
	} else {
		// TODO: check: session file must exist in case we do not perform a login
	}

	if action == Resume || action == Abort || action == Retry {
		complete(mode, XsDeployOptions, s)
	} else {
		deploy(mode, XsDeployOptions, s, nil)
	}

	if(performLogout) {
		if err = xsLogout(XsDeployOptions, s, nil, nil, nil); err != nil {
			return err
		}
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

	if fCopy == nil {
		fCopy = piperutils.Copy
	}

	xsSessionFile := ".xsconfig"
	if len(XsDeployOptions.XsSessionFile) > 0 {
		xsSessionFile = XsDeployOptions.XsSessionFile
	}

	r := strings.NewReplacer(
		"$API_URL", XsDeployOptions.APIURL,
		"$USERNAME", XsDeployOptions.User,
		"$PASSWORD", XsDeployOptions.Password,
		"$ORG", XsDeployOptions.Org,
		"$SPACE", XsDeployOptions.Space,
		"$LOGIN_OPTS", XsDeployOptions.LoginOpts,
		"$XS_SESSION_FILE", xsSessionFile)

	if e := s.RunShell("/bin/bash", r.Replace(loginScript)); e != nil {
		return e
	}

	if !fExists(xsSessionFile) {
		return fmt.Errorf("xs session file does not exist (%s)", xsSessionFile)
	}

	src, dest := fmt.Sprintf("%s/%s", os.Getenv("HOME"), xsSessionFile), fmt.Sprintf("./%s", xsSessionFile)
	log.Entry().Debugf("Copying xs session file from '%s' to '%s' (%v)", src, dest, fCopy)
	if _, err := fCopy(src, dest); err != nil {
		return errors.Wrapf(err, "Cannot copy xssession file from '%s' to '%s'", src, dest)
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
		fCopy = piperutils.Copy
	}

	if fExists == nil {
		fExists = piperutils.FileExists
	}

	if !fExists(xsSessionFile) {
		return fmt.Errorf("xs session file does not exist (%s)", xsSessionFile)
	}

	r := strings.NewReplacer(
		"$XS_SESSION_FILE", xsSessionFile)

	if e := s.RunShell("/bin/bash", r.Replace(logoutScript)); e != nil {
		return e
	}

	if e := fRemove(xsSessionFile); e != nil {
		return e
	}

	log.Entry().Debugf("xs session file '%s' has been deleted", xsSessionFile)

	log.Entry().Info("xs logout has been performed")

	return nil
}

func deploy(mode DeployMode, XsDeployOptions xsDeployOptions, s shellRunner,
	fCopy func(string, string) (int64, error)) error {

	xsSessionFile := ".xsconfig"
	if len(XsDeployOptions.XsSessionFile) > 0 {
		xsSessionFile = XsDeployOptions.XsSessionFile
	}

	if fCopy == nil {
		fCopy = piperutils.Copy
	}

	var d string

	switch mode {
		case Deploy: d = "deploy"
		case BGDeploy: d = "bg-deploy"
		default: errors.New(fmt.Sprintf("Invalid deploy mode: '%s'.", mode))
	}

	log.Entry().Debugf("Performing xs %s.", d)

	src, dest := fmt.Sprintf("./%s", xsSessionFile), fmt.Sprintf("%s/%s", os.Getenv("HOME"), xsSessionFile)
	if _, err := fCopy(src, dest); err != nil {
		return errors.Wrapf(err, "Cannot copy xssession file from '%s' to '%s'", src, dest)
	}

	r := strings.NewReplacer(
		"$d", d,
		"$MTA_PATH", XsDeployOptions.MtaPath,
		"$DEPLOY_OPTS", XsDeployOptions.DeployOpts)

	if e := s.RunShell("/bin/bash", r.Replace(deployScript)); e != nil {
		return errors.Wrapf(e, "Cannot perform xs %s", d)
	}

	log.Entry().Infof("... xs %s performed.", d)

	// TODO: in case of bg-deploy and successful deployment: read deployment id from log

	return nil

}

func complete(mode DeployMode, XsDeployOptions xsDeployOptions, s shellRunner) error {
	log.Entry().Debugf("Performing xs complete.")
	return nil
}
