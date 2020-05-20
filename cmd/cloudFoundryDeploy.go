package cmd

import (
	"bufio"
	"fmt"
	"bytes"
	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
	"gopkg.in/godo.v2/glob"
	"io"
	"os"
	"strings"
)

var _glob = glob.Glob // func(patterns []string) ([]*glob.FileAsset, []*glob.RegexpInfo, error)
var _cfLogin = cloudfoundry.Login
var _cfLogout = cloudfoundry.Logout
var fileUtils = piperutils.Files{}

const smokeTestScript = `#!/usr/bin/env bash
# this is simply testing if the application root returns HTTP STATUS_CODE
curl -so /dev/null -w '%{response_code}' https://$1 | grep $STATUS_CODE`

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
		log.Entry().WithError(err).Fatalf("step execution failed: %s", err)
	}
}

func runCloudFoundryDeploy(config *cloudFoundryDeployOptions, telemetryData *telemetry.CustomData, command execRunner) error {
	log.Entry().Infof("General parameters: deployTool='%s', cfApiEndpoint='%s'", config.DeployTool, config.APIEndpoint)

	var err error

	if config.DeployTool == "mtaDeployPlugin" {
		err = handleMTADeployment(config, command)
	} else if config.DeployTool == "cf_native" {
		err = handleCFNativeDeployment(config, command)
	} else {
		log.Entry().Warningf("Found unsupported deployTool ('%s'). Skipping deployment. Supported deploy tools: 'mtaDeployPlugin', 'cf_native'", config.DeployTool)
	}

	return err
}

func handleMTADeployment(config *cloudFoundryDeployOptions, command execRunner) error {

	mtarFilePath := config.MtarPath

	if len(mtarFilePath) == 0 {

		var err error
		mtarFilePath, err = findMtar()

		if err != nil {
			return err
		}

		log.Entry().Debugf("Using mtar file '%s' found in workspace", mtarFilePath)

	} else {

		exists, err := fileUtils.FileExists(mtarFilePath)

		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("mtar file '%s' retrieved from configuration does not exist", mtarFilePath)
		}

		log.Entry().Debugf("Using mtar file '%s' from configuration", mtarFilePath)
	}

	return deployMta(config, mtarFilePath, command)
}

type deployConfig struct {
	DeployCommand   string
	DeployOptions   []string
	AppName         string
	ManifestFile    string
	SmokeTestScript []string
}

func handleCFNativeDeployment(config *cloudFoundryDeployOptions, command execRunner) error {

	manifestFile := config.Manifest

	if len(manifestFile) == 0 {
		manifestFile = "manifest.yml"
	}

	deployType, err := checkAndUpdateDeployTypeForNotSupportedManifest(config, manifestFile)

	if err != nil {
		return err
	}

	var deployCommand string
	var smokeTestScript []string
	var deployOptions []string

	// deploy command will be provided by the prepare functions below

	if deployType == "blue-green" {
		deployCommand, deployOptions, smokeTestScript, err = prepareBlueGreenCfNativeDeploy(config)
		if err != nil {
			return errors.Wrapf(err, "Cannot prepare cf native deployment. DeployType '%s'", deployType)
		}
	} else {
		deployCommand, deployOptions, smokeTestScript, err = prepareCfPushCfNativeDeploy(config)
		if err != nil {
			return errors.Wrapf(err, "Cannot prepare cf push native deployment. DeployType '%s'", deployType)
		}

	}

	appName, err := getAppNameOrFail(config, manifestFile)

	if err != nil {
		return err
	}

	// somehow the WithField statements does not trigger the fields to be contained in the log
	log.Entry().WithField("appName", appName).WithField("deployType", deployType).WithField("manifest", config.Manifest).Errorf("CF native deployment")
	//.withField("appName", config.App)

	//.withField("") //[${STEP_NAME}] CF native deployment (${config.deployType}) with:"
	/*
	   	echo "[${STEP_NAME}] - cfManifestVariables=${config.cloudFoundry.manifestVariables ?: 'none specified'}"
	       echo "[${STEP_NAME}] - cfManifestVariablesFiles=${config.cloudFoundry.manifestVariablesFiles ?: 'none specified'}"
	       echo "[${STEP_NAME}] - cfdeployDockerImage=${config.deployDockerImage ?: 'none specified'}"
	       echo "[${STEP_NAME}] - cfdockerCredentialsId=${config.dockerCredentialsId ?: 'none specified'}"
	       echo "[${STEP_NAME}] - smokeTestScript=${config.smokeTestScript}"
	*/

	log.Entry().Infof("AppName: '%s'", appName)

	// TODO: some environment variables needs to be set

	myDeployConfig := deployConfig{
		DeployCommand:   deployCommand,
		DeployOptions:   deployOptions,
		AppName:         appName,
		ManifestFile:    manifestFile,
		SmokeTestScript: smokeTestScript,
	}

	log.Entry().Infof("DeployConfig: %v", myDeployConfig)

	//return nil
	return deployCfNative(myDeployConfig, config, command)
}

func deployCfNative(deployConfig deployConfig, config *cloudFoundryDeployOptions, command execRunner) error {

	// the deployStatement is complex and has lot of options; using a list and findAll allows to put each option
	// as a single list element; if a option is not set (= null or '') this removed before every element is joined
	// via a single whitespace; results in a single line deploy statement
	deployStatement := []string{
		"cf",
		deployConfig.DeployCommand,
		deployConfig.AppName,
	}

	if len(deployConfig.DeployOptions) > 0 {
		deployStatement = append(deployStatement, deployConfig.DeployOptions...)
	}

	if len(deployConfig.ManifestFile) > 0 {
		deployStatement = append(deployStatement, "-f")
		deployStatement = append(deployStatement, deployConfig.ManifestFile)
	}

	if len(deployConfig.SmokeTestScript) > 0 {
		deployStatement = append(deployStatement, deployConfig.SmokeTestScript...)
	}

	if len(config.CfNativeDeployParameters) > 0 {
		deployStatement = append(deployStatement, strings.FieldsFunc(config.CfNativeDeployParameters, func(c rune) bool {
			return c == ' '
		})...)
	}

	stopOldAppIfRunning := func(command execRunner) error {

		if config.KeepOldInstance && config.DeployType == "blue-green" {
			oldAppName := deployConfig.AppName + "-old"
			_ = oldAppName

			var buff bytes.Buffer

			oldOut := command.Stdout
			_ = oldOut
			command.Stdout(&buff)

			defer func() {
				command.Stdout(log.Writer())
			}()

			err := command.RunExecutable("cf", "stop", oldAppName)

			if err != nil {

				cfStopLog := buff.String()

				if ! strings.Contains(cfStopLog, oldAppName + " not found") {
					return fmt.Errorf("Could not stop application %s. Error: %s", oldAppName, cfStopLog)
				} else {
					log.Entry().Infof("Cannot stop application '%s': %s", oldAppName, cfStopLog)
				}
			} else {
				log.Entry().Infof("Old application '%s' has been stopped.", oldAppName)
			}
		}

		return nil
	}

	return cfDeploy(config, nil, deployStatement, stopOldAppIfRunning, command)
}

func getAppNameOrFail(config *cloudFoundryDeployOptions, manifestFile string) (string, error) {

	if len(config.AppName) > 0 {
		return config.AppName, nil
	} else {
		if config.DeployType == "blue-green" {
			return "", fmt.Errorf("Blue-green plugin requires app name to be passed (see https://github.com/bluemixgaragelondon/cf-blue-green-deploy/issues/27)")
		}
		if fileExists(manifestFile) {
			m, err := cloudfoundry.ReadManifest(manifestFile)
			if err != nil {
				return "", err
			}
			if len(m.Applications) > 0 && len(m.Applications[0].Name) != 0 {
				return m.Applications[0].Name, nil
			} else {
				return "", fmt.Errorf("No appName available in manifest '%s'", manifestFile)
			}
		} else {
			return "", fmt.Errorf("Manifest file '%s' not found", manifestFile)
		}
	}
	return "", fmt.Errorf("Cannot resolve app name")
}

func prepareBlueGreenCfNativeDeploy(config *cloudFoundryDeployOptions) (string, []string, []string, error) {

	if config.SmokeTestScript == "blueGreenCheckScript.sh" {
		// what should we do if there is already a script with the given name? Should we really overwrite ...
		err := fileUtils.FileWrite(config.SmokeTestScript, []byte(smokeTestScript), 0755)
		if err != nil {
			return "", []string{}, []string{}, err
		}
		log.Entry().Debugf("smoke test script '%s' has been written.", config.SmokeTestScript)
	}

	if len(config.SmokeTestScript) > 0 {
		err := os.Chmod(config.SmokeTestScript, 0755)
		if err != nil {
			return "", []string{}, []string{}, err
		}
	}
	pwd, err := os.Getwd()
	if err != nil {
		return "", []string{}, []string{}, err
	}

	var deployOptions = []string{}

	if config.KeepOldInstance {
		deployOptions = append(deployOptions, "--delete-old-apps")
	}

	handleLegacyCfManifest(config)

	return "blue-green-deploy", deployOptions, []string{"--smoke-test", fmt.Sprintf("%s/%s", pwd, config.SmokeTestScript)}, nil
}

func handleLegacyCfManifest(config *cloudFoundryDeployOptions) error {
	manifest, err := cloudfoundry.ReadManifest("manifest.yml")
	if err != nil {
		return err
	}

	changed, err := cloudfoundry.Transform(&manifest)
	if err != nil {
		return err
	}

	_ = changed
	return nil
}

func prepareCfPushCfNativeDeploy(config *cloudFoundryDeployOptions) (string, []string, []string, error) {
	return "push", []string{}, []string{}, nil
}

func checkAndUpdateDeployTypeForNotSupportedManifest(config *cloudFoundryDeployOptions, manifestFile string) (string, error) {

	manifestFileExists, err := piperutils.FileExists(manifestFile)

	if err != nil {
		return "", err
	}

	if config.DeployType == "blue-green" && manifestFileExists {

		m, _ := cloudfoundry.ReadManifest(manifestFile)

		if len(m.Applications) > 1 {
			fmt.Errorf("Your manifest contains more than one application. For blue green deployments your manifest file may contain only one application")
		}
		if len(m.Applications) == 1 && m.NoRoute {

			const deployTypeStandard = "standard"
			log.Entry().Warningf("Blue green deployment is not possible for application without route. Using deployment type '%s' instead.", deployTypeStandard)
			return deployTypeStandard, nil
		}
	}

	return config.DeployType, nil
}

func deployMta(config *cloudFoundryDeployOptions, mtarFilePath string, command execRunner) error {

	deployCommand := "deploy"
	deployParams := []string{}

	if len(config.MtaDeployParameters) > 0 {
		deployParams = append(deployParams, strings.Split(config.MtaDeployParameters, " ")...)
	}

	if config.DeployType == "bg-deploy" {

		deployCommand = "bg-deploy"

		const noConfirmFlag = "--no-confirm"
		if !contains(deployParams, noConfirmFlag) {
			deployParams = append(deployParams, noConfirmFlag)
		}
	}

	cfAPIParams := []string{
		"api",
		config.APIEndpoint,
	}

	if len(config.APIParameters) > 0 {
		cfAPIParams = append(cfAPIParams, strings.Split(config.APIParameters, " ")...)
	}

	cfDeployParams := []string{
		deployCommand,
		mtarFilePath,
	}

	if len(deployParams) > 0 {
		cfDeployParams = append(cfDeployParams, deployParams...)
	}

	{
		var mtaExtensionDescriptor string

		if len(config.MtaExtensionDescriptor) > 0 {
			if !strings.HasPrefix(config.MtaExtensionDescriptor, "-e") {
				mtaExtensionDescriptor = "-e " + config.MtaExtensionDescriptor
			} else {
				mtaExtensionDescriptor = config.MtaExtensionDescriptor
			}
		}

		if len(mtaExtensionDescriptor) > 0 {
			cfDeployParams = append(cfDeployParams, mtaExtensionDescriptor)
		}
	}

	return cfDeploy(config, cfAPIParams, cfDeployParams, nil, command)
}

// would make sense to have that method in some kind of helper instead having it here
func contains(collection []string, key string) bool {

	for _, v := range collection {
		if v == key {
			return true
		}
	}
	return false
}

func cfDeploy(config *cloudFoundryDeployOptions, cfAPIParams, cfDeployParams []string, postDeployAction func(command execRunner) error, command execRunner) error {

	const cfLogFile = "cf.log"
	var err error
	var loginPerformed bool

	// TODO set HOME to config.DockerWorkspace
	command.SetEnv([]string{"CF_TRACE=" + cfLogFile})

	if cfAPIParams != nil {
		err = command.RunExecutable("cf", cfAPIParams...)
		if err != nil {
			log.Entry().WithError(err).Errorf("Command '%s' failed.", cfAPIParams)
		}
	}

	err = _cfLogin(cloudfoundry.LoginOptions{
		CfAPIEndpoint: config.APIEndpoint,
		CfOrg:         config.Org,
		CfSpace:       config.Space,
		Username:      config.Username,
		Password:      config.Password,
	})

	if err == nil {
		loginPerformed = true
	}

	if err == nil {
		err = command.RunExecutable("cf", []string{"plugins"}...)
		if err != nil {
			log.Entry().WithError(err).Errorf("Command '%s' failed.", []string{"plugins"})
		}
	}

	if err == nil {
		err = command.RunExecutable("cf", cfDeployParams...)
		if err != nil {
			log.Entry().WithError(err).Errorf("Command '%s' failed.", cfDeployParams)
		}
	}

	if err == nil && postDeployAction != nil {
		err = postDeployAction(command)
	}

	if loginPerformed {

		logoutErr := _cfLogout()

		if logoutErr != nil {
			log.Entry().WithError(logoutErr).Errorf("Cannot perform cf logout")
			if err == nil {
				err = logoutErr
			}
		}
	}

	if err != nil || GeneralConfig.Verbose {
		e := handleCfCliLog(cfLogFile)
		if e != nil {
			log.Entry().WithError(err).Errorf("Error reading cf log file '%s'.", cfLogFile)
		}
	}

	return err
}

func findMtar() (string, error) {

	const pattern = "**/*.mtar"

	mtars, _, err := _glob([]string{pattern})

	if err != nil {
		return "", err
	}

	if len(mtars) == 0 {
		return "", fmt.Errorf("No mtar file matching pattern '%s' found", pattern)
	}

	if len(mtars) > 1 {
		sMtars := []string{}
		for _, mtar := range mtars {
			sMtars = append(sMtars, mtar.Path)
		}
		return "", fmt.Errorf("Found multiple mtar files matching pattern '%s' (%s), please specify file via mtaPath parameter 'TODO-PARAMETER'", pattern, strings.Join(sMtars, ","))
	}

	return mtars[0].Path, nil
}

func handleCfCliLog(logFile string) error {

	fExists, err := fileUtils.FileExists(logFile)

	if err != nil {
		return err
	}

	log.Entry().Info("### START OF CF CLI TRACE OUTPUT ###")

	if fExists {

		f, err := os.Open(logFile)

		if err != nil {
			return err
		}

		bReader := bufio.NewReader(f)

		var done bool

		for {
			line, err := bReader.ReadString('\n')
			if err == io.EOF {

				// maybe inappropriate to log as info. Maybe the line from the
				// log indicates an error, but that is something like a project
				// standard.
				done = true
			} else if err != nil {

				break
			}

			log.Entry().Info(line)

			if done {
				break
			}
		}

	} else {
		log.Entry().Warningf("No trace file found at '%s'", logFile)
	}

	log.Entry().Info("### END OF CF CLI TRACE OUTPUT ###")

	return err
}
