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
	"sync"
	"text/template"
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
xs login -a {{.APIURL}} -u {{.User}} -p '{{.Password}}' -o {{.Org}} -s {{.Space}} {{.LoginOpts}}
`

const logoutScript = `#!/bin/bash
xs logout`

const deployScript = `#!/bin/bash
xs {{.Mode}} {{.MtaPath}} {{.DeployOpts}}`


func xsDeploy(myXsDeployOptions xsDeployOptions) error {
	c := command.Command{}
	return runXsDeploy(myXsDeployOptions, &c)
}

func runXsDeploy(XsDeployOptions xsDeployOptions, s shellRunner) error {

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

	// TODO: check: for action NONE --> deployable must exist.
	// Should be done before even trying to login

	var loginErr error

	if performLogin {
		loginErr = xsLogin(XsDeployOptions, s, nil, nil)
	} else {
		// TODO: check: session file must exist in case we do not perform a login
	}

	if loginErr == nil && (action == Resume || action == Abort || action == Retry) {
		err = complete(mode, XsDeployOptions, s)
	} else {
		err = deploy(mode, XsDeployOptions, s, nil)
	}

	if loginErr == nil && (performLogout || err != nil) {
		if logoutErr := xsLogout(XsDeployOptions, s, nil, nil, nil); err != nil {
			if err == nil {
				err = logoutErr
			}
		}
	}

	if err == nil {
		err = loginErr
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

	log.Entry().Debugf("Performing xs login. api-url: '%s', org: '%s', space: '%s'",
	XsDeployOptions.APIURL, XsDeployOptions.Org, XsDeployOptions.Space)

	if e := executeCmd("login", loginScript, XsDeployOptions, s); e != nil {
		return e
	}

	log.Entry().Infof("xs login has been performed. api-url: '%s', org: '%s', space: '%s'",
		XsDeployOptions.APIURL, XsDeployOptions.Org, XsDeployOptions.Space)

	if !fExists(xsSessionFile) {
		return fmt.Errorf("xs session file does not exist (%s)", xsSessionFile)
	}

	src, dest := fmt.Sprintf("%s/%s", os.Getenv("HOME"), xsSessionFile), fmt.Sprintf("%s", xsSessionFile)
	log.Entry().Debugf("Copying xs session file from '%s' to '%s'", src, dest)
	if _, err := fCopy(src, dest); err != nil {
		return errors.Wrapf(err, "Cannot copy xssession file from '%s' to '%s'", src, dest)
	}

	log.Entry().Debugf("xs session file copied from '%s' to '%s'", src, dest)

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

	if e := executeCmd("logout", logoutScript, XsDeployOptions, s); e != nil {
		return e
	}
	log.Entry().Info("xs logout has been performed")

	if e := fRemove(xsSessionFile); e != nil {
		return e
	}

	log.Entry().Debugf("xs session file '%s' has been deleted", xsSessionFile)

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

	var deployCommand string

	switch mode {
		case Deploy: deployCommand= "deploy"
		case BGDeploy: deployCommand = "bg-deploy"
		default: errors.New(fmt.Sprintf("Invalid deploy mode: '%s'.", mode))
	}

	type deployProperties struct {
		xsDeployOptions
		Mode string
	}

	src, dest := fmt.Sprintf("./%s", xsSessionFile), fmt.Sprintf("%s/%s", os.Getenv("HOME"), xsSessionFile)
	if _, err := fCopy(src, dest); err != nil {
		return errors.Wrapf(err, "Cannot copy xssession file from '%s' to '%s'", src, dest)
	}

	log.Entry().Debugf("Performing xs %s.", deployCommand)
	if e := executeCmd("deploy", deployScript, deployProperties{xsDeployOptions: XsDeployOptions, Mode: deployCommand}, s); e != nil {
		return e
	}
	log.Entry().Infof("... xs %s performed.", deployCommand)

	// TODO: in case of bg-deploy and successful deployment: read deployment id from log

	return nil

}

func complete(mode DeployMode, XsDeployOptions xsDeployOptions, s shellRunner) error {
	log.Entry().Debugf("Performing xs complete.")
	return nil
}

func executeCmd(templateID string, commandPattern string, properties interface{}, s shellRunner) error {

	tmpl, e := template.New(templateID).Parse(commandPattern)
	if(e != nil) {
		return e
	}

	var script bytes.Buffer
	tmpl.Execute(&script, properties)
	if e := s.RunShell("/bin/bash", script.String()); e != nil {
		return e
	}

	return nil
}
