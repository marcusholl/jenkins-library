package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/godo.v2/glob"
	"os"
	"testing"
)

type manifestMock struct {
	manifestFileName string
	apps             []map[string]interface{}
}

func (m manifestMock) GetAppName(index int) (string, error) {
	val, err := m.GetApplicationProperty(index, "name")
	if err != nil {
		return "", err
	}
	if v, ok := val.(string); ok {
		return v, nil
	}
	return "", fmt.Errorf("Cannot resolve application name")
}
func (m manifestMock) ApplicationHasProperty(index int, name string) (bool, error) {
	_, exists := m.apps[index][name]
	return exists, nil
}
func (m manifestMock) GetApplicationProperty(index int, name string) (interface{}, error) {
	return m.apps[index][name], nil
}
func (m manifestMock) GetFileName() string {
	return m.manifestFileName
}
func (m manifestMock) Transform() error {
	return nil
}
func (m manifestMock) IsModified() bool {
	return false
}
func (m manifestMock) GetApplications() ([]map[string]interface{}, error) {
	return m.apps, nil
}
func (m manifestMock) WriteManifest() error {
	return nil
}

func TestCfDeployment(t *testing.T) {

	// everything below in the config map annotated with '//default' is a default in the metadata
	// since we don't get injected these values during the tests we set it here.
	defaultConfig := cloudFoundryDeployOptions{
		Org:                 "myOrg",
		Space:               "mySpace",
		Username:            "me",
		Password:            "******",
		APIEndpoint:         "https://examples.sap.com/cf",
		SmokeTestStatusCode: "200",          // default
		Manifest:            "manifest.yml", //default
		MtaDeployParameters: "-f",           // default
	}

	config := defaultConfig

	successfulLogin := cloudfoundry.LoginOptions{
		CfAPIEndpoint: "https://examples.sap.com/cf",
		CfOrg:         "myOrg",
		CfSpace:       "mySpace",
		Username:      "me",
		Password:      "******",
	}

	var loginOpts cloudfoundry.LoginOptions
	var logoutCalled, mtarFileRetrieved bool

	noopCfAPICalls := func(t *testing.T, s mock.ExecMockRunner) {
		assert.Empty(t, s.Calls)   // --> in case of an invalid deploy tool there must be no cf api calls
		assert.Empty(t, loginOpts) // no login options: login has not been called
		assert.False(t, logoutCalled)
	}

	withLoginAndLogout := func(t *testing.T, asserts func(t *testing.T)) {

		assert.Equal(t, loginOpts, successfulLogin)
		asserts(t)
		assert.True(t, logoutCalled)
	}

	var cleanup = func() {
		loginOpts = cloudfoundry.LoginOptions{}
		logoutCalled = false
		mtarFileRetrieved = false
		config = defaultConfig
	}

	defer func() {
		_glob = glob.Glob
		_cfLogin = cloudfoundry.Login
		_cfLogout = cloudfoundry.Logout
	}()

	_glob = func(patterns []string) ([]*glob.FileAsset, []*glob.RegexpInfo, error) {
		mtarFileRetrieved = true
		return []*glob.FileAsset{&glob.FileAsset{Path: "x.mtar"}}, nil, nil
	}

	_cfLogin = func(opts cloudfoundry.LoginOptions) error {
		loginOpts = opts
		return nil
	}

	_cfLogout = func() error {
		logoutCalled = true
		return nil
	}

	t.Run("Invalid deploytool", func(t *testing.T) {

		defer cleanup()

		s := mock.ExecMockRunner{}

		config.DeployTool = "invalid"

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {
			noopCfAPICalls(t, s)
		}
	})

	t.Run("deploytool cf native", func(t *testing.T) {

		defer cleanup()

		defer func() {
			_getWd = os.Getwd
			_fileExists = piperutils.FileExists
			_getManifest = getManifest
		}()

		_getWd = func() (string, error) {
			return "/home/me", nil
		}

		_fileExists = func(name string) (bool, error) {
			return name == "manifest.yml", nil
		}

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "manifest.yml",
					apps:             []map[string]interface{}{map[string]interface{}{"name": "testAppName"}}},
				nil
		}

		config.DeployTool = "cf_native"

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check cf api calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {
					assert.Equal(t, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{"push", "-f", "manifest.yml"}},
					}, s.Calls)
				})
			})
		}

		t.Run("check environment variables", func(t *testing.T) {
			assert.Contains(t, s.Env, "CF_HOME=/home/me")        // REVISIT: cross check if that variable should point to the user home dir
			assert.Contains(t, s.Env, "CF_PLUGIN_HOME=/home/me") // REVISIT: cross check if that variable should point to the user home dir
			assert.Contains(t, s.Env, "STATUS_CODE=200")
		})
	})

	t.Run("deploy cf native with docker image and docker username", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployDockerImage = "repo/image:tag"
		config.DockerUsername = "me"
		config.AppName = "testAppName"

		config.Manifest = ""

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			withLoginAndLogout(t, func(t *testing.T) {
				assert.Equal(t, []mock.ExecCall{
					mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
					mock.ExecCall{Exec: "cf", Params: []string{"push",
						"testAppName",
						"--docker-image",
						"repo/image:tag",
						"--docker-username",
						"me",
					}},
				}, s.Calls)
			})
		}
	})

	t.Run("deploy_cf_native with manifest and docker credentials", func(t *testing.T) {

		defer cleanup()

		// Docker image can be done via manifest.yml.
		// if a private Docker registry is used, --docker-username and DOCKER_PASSWORD
		// must be set; this is checked by this test

		config.DeployTool = "cf_native"
		config.DeployDockerImage = "repo/image:tag"
		config.DockerUsername = "test_cf_docker"
		config.DockerPassword = "********"
		config.AppName = "testAppName"

		config.Manifest = ""

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {
			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{"push",
							"testAppName",
							"--docker-image",
							"repo/image:tag",
							"--docker-username",
							"test_cf_docker",
						}},
					}, s.Calls)
				})
			})

			t.Run("check environment variables", func(t *testing.T) {
				//REVISIT: in the corresponding groovy test we checked for "${'********'}"
				// I don't understand why, but we should discuss ...
				assert.Contains(t, s.Env, "CF_DOCKER_PASSWORD=********")
			})
		}
	})

	t.Run("deploy cf native blue green with manifest and docker credentials", func(t *testing.T) {

		defer cleanup()

		// Blue Green Deploy cf cli plugin does not support --docker-username and --docker-image parameters
		// docker username and docker image have to be set in the manifest file
		// if a private docker repository is used the CF_DOCKER_PASSWORD env variable must be set

		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.DockerUsername = "test_cf_docker"
		config.DockerPassword = "********"
		config.AppName = "testAppName"

		defer func() {
			_fileExists = piperutils.FileExists
			_getManifest = getManifest
		}()

		_fileExists = func(name string) (bool, error) {
			return name == "manifest.yml", nil
		}

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "manifest.yml",
					apps:             []map[string]interface{}{map[string]interface{}{"name": "testAppName"}}},
				nil
		}

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
							"blue-green-deploy",
							"testAppName",
							"--delete-old-apps",
							"-f",
							"manifest.yml",
						}},
					}, s.Calls)
				})
			})

			t.Run("check environment variables", func(t *testing.T) {
				//REVISIT: in the corresponding groovy test we checked for "${'********'}"
				// I don't understand why, but we should discuss ...
				assert.Contains(t, s.Env, "CF_DOCKER_PASSWORD=********")
			})
		}
	})

	t.Run("deploy cf native app name from manifest", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.Manifest = "test-manifest.yml"

		defer func() {
			_fileExists = piperutils.FileExists
			_getManifest = getManifest
		}()

		_fileExists = func(name string) (bool, error) {
			return name == "test-manifest.yml", nil
		}

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					// app name is not asserted since it does not appear in the cf calls
					// but it is checked that an app name is present, hence we need it here.

					apps: []map[string]interface{}{map[string]interface{}{"name": "dummyApp"}}},
				nil
		}

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
							"push",
							"-f",
							"test-manifest.yml",
						}},
					}, s.Calls)

				})
			})
		}
	})

	t.Run("deploy cf native without app name", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.Manifest = "test-manifest.yml"

		defer func() {
			_fileExists = piperutils.FileExists
			_getManifest = getManifest
		}()

		_fileExists = func(name string) (bool, error) {
			return name == "test-manifest.yml", nil
		}

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					// Here we don't provide an application name from the mock. To make that
					// more explicit we provide the empty string default explicitly.
					apps: []map[string]interface{}{map[string]interface{}{"name": ""}}},
				nil
		}

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.EqualError(t, err, "No appName available in manifest 'test-manifest.yml'") {

			t.Run("check shell calls", func(t *testing.T) {
				noopCfAPICalls(t, s)
			})
		}
	})

	// tests from groovy checking for keep old instances are already contained above. Search for '--delete-old-apps'

	t.Run("deploy cf native blue green keep old instance", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"
		config.KeepOldInstance = true

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
							"blue-green-deploy",
							"myTestApp",
							"-f",
							"test-manifest.yml",
						}},
						mock.ExecCall{Exec: "cf", Params: []string{
							"stop",
							"myTestApp-old",
							// MIGRATE FFROM GROOVY: in contrast to groovy there is not redirect of everything &> to a file since we
							// read the stream directly now.
						}},
					}, s.Calls)
				})
			})
		}
	})

	t.Run("cf deploy blue green multiple applications", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"

		defer func() {
			_fileExists = piperutils.FileExists
			_getManifest = getManifest
		}()

		_fileExists = func(name string) (bool, error) {
			return name == "test-manifest.yml", nil
		}

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						map[string]interface{}{"name": "app1"},
						map[string]interface{}{"name": "app2"},
					},
				},
				nil
		}

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.EqualError(t, err, "Your manifest contains more than one application. For blue green deployments your manifest file may contain only one application") {
			t.Run("check shell calls", func(t *testing.T) {
				noopCfAPICalls(t, s)
			})
		}
	})

	t.Run("cf native deploy blue green with no route", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"

		defer func() {
			_fileExists = piperutils.FileExists
			_getManifest = getManifest
		}()

		_fileExists = func(name string) (bool, error) {
			return name == "test-manifest.yml", nil
		}

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						map[string]interface{}{
							"name":     "app1",
							"no-route": true,
						},
					},
				},
				nil
		}

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
							"push",
							"myTestApp",
							"-f",
							"test-manifest.yml",
						}},
					}, s.Calls)
				})
			})
		}
	})

	t.Run("cf native deployment failure", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"

		defer func() {
			_fileExists = piperutils.FileExists
			_getManifest = getManifest
		}()

		_fileExists = func(name string) (bool, error) {
			return name == "test-manifest.yml", nil
		}

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						map[string]interface{}{
							"name":     "app1",
							"no-route": true,
						},
					},
				},
				nil
		}

		s := mock.ExecMockRunner{}

		s.ShouldFailOnCommand = map[string]error{"cf.*": fmt.Errorf("cf deploy failed")}
		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.EqualError(t, err, "cf deploy failed") {
			t.Run("check shell calls", func(t *testing.T) {

				// we should try to logout in this case
				assert.True(t, logoutCalled)
			})
		}
	})

	t.Run("cf native deployment failure when logging in", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"

		defer func() {
			_fileExists = piperutils.FileExists
			_getManifest = getManifest

			_cfLogin = func(opts cloudfoundry.LoginOptions) error {
				loginOpts = opts
				return nil
			}
		}()

		_cfLogin = func(opts cloudfoundry.LoginOptions) error {
			loginOpts = opts
			return fmt.Errorf("Unable to login")
		}

		_fileExists = func(name string) (bool, error) {
			return name == "test-manifest.yml", nil
		}

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						map[string]interface{}{
							"name":     "app1",
							"no-route": true,
						},
					},
				},
				nil
		}

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.EqualError(t, err, "Unable to login") {
			t.Run("check shell calls", func(t *testing.T) {

				// no calls to the cf client in this case
				assert.Empty(t, s.Calls)
				// no logout
				assert.False(t, logoutCalled)
			})
		}
	})

	// TODO testCfNativeBlueGreenKeepOldInstanceShouldThrowErrorOnStopError

	t.Run("cf native deploy standard should not stop instance", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "standard"
		config.Manifest = "test-manifest.yml"
		config.AppName = "myTestApp"
		config.KeepOldInstance = true

		defer func() {
			_fileExists = piperutils.FileExists
			_getManifest = getManifest
		}()

		_fileExists = func(name string) (bool, error) {
			return name == "test-manifest.yml", nil
		}

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						map[string]interface{}{
							"name":     "app1",
							"no-route": true,
						},
					},
				},
				nil
		}

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
							"push",
							"myTestApp",
							"-f",
							"test-manifest.yml",
						}},

						//
						// There is no cf stop
						//

					}, s.Calls)
				})
			})
		}
	})

	t.Run("testCfNativeWithoutAppNameBlueGreen", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.DeployType = "blue-green"
		config.Manifest = "test-manifest.yml"

		defer func() {
			_fileExists = piperutils.FileExists
			_getManifest = getManifest
		}()

		_fileExists = func(name string) (bool, error) {
			return name == "test-manifest.yml", nil
		}

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						map[string]interface{}{
							"there-is": "no-app-name",
						},
					},
				},
				nil
		}

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.EqualError(t, err, "Blue-green plugin requires app name to be passed (see https://github.com/bluemixgaragelondon/cf-blue-green-deploy/issues/27)") {

			t.Run("check shell calls", func(t *testing.T) {
				noopCfAPICalls(t, s)
			})
		}
	})

	// TODO add test for testCfNativeFailureInShellCall

	t.Run("deploytool mtaDeployPlugin blue green", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "mtaDeployPlugin"
		config.DeployType = "blue-green"
		config.MtarPath = "target/test.mtar"

		defer func() {
			_fileExists = piperutils.FileExists
		}()

		_fileExists = func(name string) (bool, error) {
			return name == "target/test.mtar", nil
		}

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"api", "https://examples.sap.com/cf"}},
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
							"bg-deploy",
							"target/test.mtar",
							"-f",
							"--no-confirm",
						}},

						//
						// There is no cf stop
						//

					}, s.Calls)
				})
			})
		}
	})

	// TODO: add test for influx reporting (influx reporting is missing at the moment)

	t.Run("cf push with variables from file and as list", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.Manifest = "test-manifest.yml"
		config.ManifestVariablesFiles = []string{"vars.yaml"}
		config.ManifestVariables = []string{"appName=testApplicationFromVarsList"}
		config.AppName = "testAppName"

		defer func() {
			_fileExists = piperutils.FileExists
			_getManifest = getManifest
		}()

		_fileExists = func(name string) (bool, error) {
			return name == "test-manifest.yml" || name == "vars.yaml", nil
		}

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						map[string]interface{}{
							"name": "myApp",
						},
					},
				},
				nil
		}

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {

					// Revisit: we don't verify a log message in case of a non existing vars file

					assert.Equal(t, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
							"push",
							"testAppName",
							"--var",
							"appName=testApplicationFromVarsList",
							"--vars-file",
							"vars.yaml",
							"-f",
							"test-manifest.yml",
						}},
					}, s.Calls)
				})
			})
		}
	})

	t.Run("cf push with variables from file which does not exist", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "cf_native"
		config.Manifest = "test-manifest.yml"
		config.ManifestVariablesFiles = []string{"vars.yaml", "vars-does-not-exist.yaml"}
		config.AppName = "testAppName"

		defer func() {
			_fileExists = piperutils.FileExists
			_getManifest = getManifest
		}()

		_fileExists = func(name string) (bool, error) {
			return name == "test-manifest.yml" || name == "vars.yaml", nil
		}

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						map[string]interface{}{
							"name": "myApp",
						},
					},
				},
				nil
		}

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {

				withLoginAndLogout(t, func(t *testing.T) {
					// Revisit: we don't verify a log message in case of a non existing vars file

					assert.Equal(t, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{
							"push",
							"testAppName",
							"--vars-file",
							"vars.yaml",
							"-f",
							"test-manifest.yml",
						}},
					}, s.Calls)
				})
			})
		}
	})

	// TODO: testCfPushDeploymentWithoutVariableSubstitution is already handled above (?)

	// TODO: testCfBlueGreenDeploymentWithVariableSubstitution variable substitution is not handled at the moment (pr pending).
	// but anyway we should not test the full cycle here, but only that the variables substitution tool is called in the appropriate way.
	// variable substitution should be tested at the variables substitution tool itself (yaml util)

	t.Run("deploytool mtaDeployPlugin", func(t *testing.T) {

		defer cleanup()

		config.DeployTool = "mtaDeployPlugin"
		config.MtaDeployParameters = "-f"

		t.Run("mta config file from project sources", func(t *testing.T) {

			s := mock.ExecMockRunner{}
			err := runCloudFoundryDeploy(&config, nil, &s)

			if assert.NoError(t, err) {

				withLoginAndLogout(t, func(t *testing.T) {

					assert.Equal(t, s.Calls, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"api", "https://examples.sap.com/cf"}},
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{"deploy", "x.mtar", "-f"}}})

				})

				t.Run("mtar retrieved", func(t *testing.T) {
					assert.True(t, mtarFileRetrieved)
				})
			}
		})

		t.Run("mta config file from project config does not exist", func(t *testing.T) {
			defer func() { config.MtarPath = "" }()
			config.MtarPath = "my.mtar"
			s := mock.ExecMockRunner{}
			err := runCloudFoundryDeploy(&config, nil, &s)
			assert.EqualError(t, err, "mtar file 'my.mtar' retrieved from configuration does not exist")
		})

		// TODO: add test for mtar file from project config which does exist in project sources
	})
}

func TestManifestVariableFiles(t *testing.T) {

	defer func() {
		_fileExists = piperutils.FileExists
	}()

	_fileExists = func(name string) (bool, error) {
		return name == "a/varsA.txt" || name == "varsB.txt", nil
	}

	t.Run("straight forward", func(t *testing.T) {
		varOpts, err := getVarFileOptions([]string{"a/varsA.txt", "varsB.txt"})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{"--vars-file", "a/varsA.txt", "--vars-file", "varsB.txt"}, varOpts)
		}
	})

	t.Run("no var filesprovided", func(t *testing.T) {
		varOpts, err := getVarFileOptions([]string{})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{}, varOpts)
		}
	})

	t.Run("one var file does not exist", func(t *testing.T) {
		varOpts, err := getVarFileOptions([]string{"a/varsA.txt", "doesNotExist.txt"})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{"--vars-file", "a/varsA.txt"}, varOpts)
		}
	})
}

func TestManifestVariables(t *testing.T) {
	t.Run("straight forward", func(t *testing.T) {
		varOpts, err := getVarOptions([]string{"a=b", "c=d"})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{"--var", "a=b", "--var", "c=d"}, varOpts)
		}
	})

	t.Run("empty variabls list", func(t *testing.T) {
		varOpts, err := getVarOptions([]string{})
		if assert.NoError(t, err) {
			assert.Equal(t, []string{}, varOpts)
		}
	})

	t.Run("no equal sign in variable", func(t *testing.T) {
		_, err := getVarOptions([]string{"ab"})
		assert.EqualError(t, err, "Invalid parameter provided (expected format <key>=<val>: 'ab'")
	})
}

func TestMtarLookup(t *testing.T) {
	t.Run("One MTAR", func(t *testing.T) {

		defer func() {
			_glob = glob.Glob
		}()

		_glob = func(patterns []string) ([]*glob.FileAsset, []*glob.RegexpInfo, error) {
			return []*glob.FileAsset{&glob.FileAsset{Path: "x.mtar"}}, nil, nil
		}

		path, err := findMtar()

		if assert.NoError(t, err) {
			assert.Equal(t, "x.mtar", path)
		}
	})

	t.Run("No MTAR", func(t *testing.T) {

		defer func() {
			_glob = glob.Glob
		}()

		_glob = func(patterns []string) ([]*glob.FileAsset, []*glob.RegexpInfo, error) {
			return []*glob.FileAsset{}, nil, nil
		}

		_, err := findMtar()

		assert.EqualError(t, err, "No mtar file matching pattern '**/*.mtar' found")
	})

	t.Run("Several MTARs", func(t *testing.T) {

		defer func() {
			_glob = glob.Glob
		}()

		_glob = func(patterns []string) ([]*glob.FileAsset, []*glob.RegexpInfo, error) {
			return []*glob.FileAsset{&glob.FileAsset{Path: "x.mtar"}, &glob.FileAsset{Path: "y.mtar"}}, nil, nil
		}

		_, err := findMtar()
		assert.EqualError(t, err, "Found multiple mtar files matching pattern '**/*.mtar' (x.mtar,y.mtar), please specify file via mtaPath parameter 'mtarPath'")
	})
}
