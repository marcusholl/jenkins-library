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

	var loginOpts cloudfoundry.LoginOptions
	var logoutCalled, mtarFileRetrieved bool

	var cleanup = func() {
		loginOpts = cloudfoundry.LoginOptions{}
		logoutCalled = false
		mtarFileRetrieved = false
		_getWd = os.Getwd
	}

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

		config := cloudFoundryDeployOptions{DeployTool: "invalid"}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {
			assert.Empty(t, s.Calls)                                // --> in case of an invalid deploy tool there must be no cf api calls
			assert.Equal(t, loginOpts, cloudfoundry.LoginOptions{}) // no login options: login has not been called
			assert.False(t, logoutCalled)
		}
	})

	t.Run("deploytool cf native", func(t *testing.T) {

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

		config := cloudFoundryDeployOptions{
			DeployTool:          "cf_native",
			Org:                 "myOrg",
			Space:               "mySpace",
			Username:            "me",
			Password:            "******",
			APIEndpoint:         "https://examples.sap.com/cf",
			SmokeTestStatusCode: "200",
			Manifest:            "manifest.yml", // the default, will be provided in the free wild from the metadata, but here we have to set it.
		}

		defer cleanup()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check cf api calls", func(t *testing.T) {

				assert.Equal(t, loginOpts,
					cloudfoundry.LoginOptions{
						CfAPIEndpoint: "https://examples.sap.com/cf",
						CfOrg:         "myOrg",
						CfSpace:       "mySpace",
						Username:      "me",
						Password:      "******",
					})

				// REVISIT: we have more the less the same test below (deploy cf native app name from manifest)
				// that other test has been transfered from groovy. But here we have some more checks for the
				// environment variables. --> check if we really need both tests and try to find out why there
				// was no need for asserting the environment variables on groovy.

				assert.Equal(t, []mock.ExecCall{
					mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
					mock.ExecCall{Exec: "cf", Params: []string{"push", "-f", "manifest.yml"}},
				}, s.Calls)
			})

			assert.True(t, logoutCalled)
		}

		t.Run("check environment variables", func(t *testing.T) {
			assert.Contains(t, s.Env, "CF_HOME=/home/me")        // REVISIT: cross check if that variable should point to the user home dir
			assert.Contains(t, s.Env, "CF_PLUGIN_HOME=/home/me") // REVISIT: cross check if that variable should point to the user home dir
			assert.Contains(t, s.Env, "STATUS_CODE=200")
		})
	})

	t.Run("deploy cf native with docker image and docker username", func(t *testing.T) {

		config := cloudFoundryDeployOptions{
			DeployTool:        "cf_native",
			Org:               "myOrg",
			Space:             "mySpace",
			Username:          "me",
			Password:          "******",
			APIEndpoint:       "https://examples.sap.com/cf",
			DeployDockerImage: "repo/image:tag",
			DockerUsername:    "me",
			AppName:           "testAppName",
		}

		defer cleanup()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			assert.Equal(t, loginOpts,
				cloudfoundry.LoginOptions{
					CfAPIEndpoint: "https://examples.sap.com/cf",
					CfOrg:         "myOrg",
					CfSpace:       "mySpace",
					Username:      "me",
					Password:      "******",
				})

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

			assert.True(t, logoutCalled)
		}
	})

	t.Run("deploy_cf_native with manifest and docker credentials", func(t *testing.T) {

		// Docker image can be done via manifest.yml.
		// if a private Docker registry is used, --docker-username and DOCKER_PASSWORD
		// must be set; this is checked by this test

		config := cloudFoundryDeployOptions{
			DeployTool:        "cf_native",
			Org:               "myOrg",
			Space:             "mySpace",
			Username:          "me",
			Password:          "******",
			APIEndpoint:       "https://examples.sap.com/cf",
			DeployDockerImage: "repo/image:tag",
			DockerUsername:    "test_cf_docker",
			DockerPassword:    "********",
			AppName:           "testAppName",
		}

		defer cleanup()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {
			t.Run("check shell calls", func(t *testing.T) {
				assert.Equal(t, loginOpts,
					cloudfoundry.LoginOptions{
						CfAPIEndpoint: "https://examples.sap.com/cf",
						CfOrg:         "myOrg",
						CfSpace:       "mySpace",
						Username:      "me",
						Password:      "******",
					})

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

				assert.True(t, logoutCalled)
			})

			t.Run("check environment variables", func(t *testing.T) {
				//REVISIT: in the corresponding groovy test we checked for "${'********'}"
				// I don't understand why, but we should discuss ...
				assert.Contains(t, s.Env, "CF_DOCKER_PASSWORD=********")
			})
		}
	})

	t.Run("deploy cf native blue green with manifest and docker credentials", func(t *testing.T) {

		// Blue Green Deploy cf cli plugin does not support --docker-username and --docker-image parameters
		// docker username and docker image have to be set in the manifest file
		// if a private docker repository is used the CF_DOCKER_PASSWORD env variable must be set

		config := cloudFoundryDeployOptions{
			DeployTool:     "cf_native",
			DeployType:     "blue-green",
			Org:            "myOrg",
			Space:          "mySpace",
			Username:       "me",
			Password:       "******",
			APIEndpoint:    "https://examples.sap.com/cf",
			DockerUsername: "test_cf_docker",
			DockerPassword: "********",
			AppName:        "testAppName",
			Manifest:       "manifest.yml",
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

		defer cleanup()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {
				assert.Equal(t, loginOpts,
					cloudfoundry.LoginOptions{
						CfAPIEndpoint: "https://examples.sap.com/cf",
						CfOrg:         "myOrg",
						CfSpace:       "mySpace",
						Username:      "me",
						Password:      "******",
					})

				assert.Equal(t, []mock.ExecCall{
					mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},

					//cf blue-green-deploy testAppName --delete-old-apps -f 'manifest.yml'

					mock.ExecCall{Exec: "cf", Params: []string{
						"blue-green-deploy",
						"testAppName",
						"--delete-old-apps",
						"-f",
						"manifest.yml",
					}},
				}, s.Calls)

				assert.True(t, logoutCalled)
			})

			t.Run("check environment variables", func(t *testing.T) {
				//REVISIT: in the corresponding groovy test we checked for "${'********'}"
				// I don't understand why, but we should discuss ...
				assert.Contains(t, s.Env, "CF_DOCKER_PASSWORD=********")
			})
		}
	})

	t.Run("deploy cf native app name from manifest", func(t *testing.T) {

		config := cloudFoundryDeployOptions{
			DeployTool:  "cf_native",
			Org:         "myOrg",
			Space:       "mySpace",
			Username:    "me",
			Password:    "******",
			APIEndpoint: "https://examples.sap.com/cf",
			Manifest:    "test-manifest.yml",
		}

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

		defer cleanup()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {
				assert.Equal(t, loginOpts,
					cloudfoundry.LoginOptions{
						CfAPIEndpoint: "https://examples.sap.com/cf",
						CfOrg:         "myOrg",
						CfSpace:       "mySpace",
						Username:      "me",
						Password:      "******",
					})

				assert.Equal(t, []mock.ExecCall{
					mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
					mock.ExecCall{Exec: "cf", Params: []string{
						"push",
						"-f",
						"test-manifest.yml",
					}},
				}, s.Calls)

				assert.True(t, logoutCalled)
			})
		}
	})

	t.Run("deploy cf native without app name", func(t *testing.T) {

		config := cloudFoundryDeployOptions{
			DeployTool:  "cf_native",
			Org:         "myOrg",
			Space:       "mySpace",
			Username:    "me",
			Password:    "******",
			APIEndpoint: "https://examples.sap.com/cf",
			Manifest:    "test-manifest.yml",
		}

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

		defer cleanup()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.EqualError(t, err, "No appName available in manifest 'test-manifest.yml'") {

			t.Run("check shell calls", func(t *testing.T) {

				// no login in this case
				assert.Equal(t, cloudfoundry.LoginOptions{}, loginOpts)
				// no calls to the cf client in this case
				assert.Empty(t, s.Calls)
				// no logout
				assert.False(t, logoutCalled)
			})
		}
	})

	// tests from groovy checking for keep old instances are already contained above. Search for '--delete-old-apps'

	t.Run("deploy cf native blue green keep old instance", func(t *testing.T) {

		config := cloudFoundryDeployOptions{
			DeployTool:      "cf_native",
			DeployType:      "blue-green",
			Org:             "myOrg",
			Space:           "mySpace",
			Username:        "me",
			Password:        "******",
			APIEndpoint:     "https://examples.sap.com/cf",
			Manifest:        "test-manifest.yml",
			AppName:         "myTestApp",
			KeepOldInstance: true,
		}

		defer cleanup()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {
				assert.Equal(t, loginOpts,
					cloudfoundry.LoginOptions{
						CfAPIEndpoint: "https://examples.sap.com/cf",
						CfOrg:         "myOrg",
						CfSpace:       "mySpace",
						Username:      "me",
						Password:      "******",
					})

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

				assert.True(t, logoutCalled)
			})
		}
	})

	t.Run("cf deploy blue green multiple applications", func(t *testing.T) {

		config := cloudFoundryDeployOptions{
			DeployTool:  "cf_native",
			DeployType:  "blue-green",
			Org:         "myOrg",
			Space:       "mySpace",
			Username:    "me",
			Password:    "******",
			APIEndpoint: "https://examples.sap.com/cf",
			Manifest:    "test-manifest.yml",
			AppName:     "myTestApp",
		}

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

		defer cleanup()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.EqualError(t, err, "Your manifest contains more than one application. For blue green deployments your manifest file may contain only one application") {
			t.Run("check shell calls", func(t *testing.T) {

				// no login in this case
				assert.Equal(t, cloudfoundry.LoginOptions{}, loginOpts)
				// no calls to the cf client in this case
				assert.Empty(t, s.Calls)
				// no logout
				assert.False(t, logoutCalled)
			})
		}
	})

	t.Run("cf native deploy with no route", func(t *testing.T) {

		config := cloudFoundryDeployOptions{
			DeployTool:  "cf_native",
			DeployType:  "blue-green",
			Org:         "myOrg",
			Space:       "mySpace",
			Username:    "me",
			Password:    "******",
			APIEndpoint: "https://examples.sap.com/cf",
			Manifest:    "test-manifest.yml",
			AppName:     "myTestApp",
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

		defer cleanup()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {
				assert.Equal(t, loginOpts,
					cloudfoundry.LoginOptions{
						CfAPIEndpoint: "https://examples.sap.com/cf",
						CfOrg:         "myOrg",
						CfSpace:       "mySpace",
						Username:      "me",
						Password:      "******",
					})

				assert.Equal(t, []mock.ExecCall{
					mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
					mock.ExecCall{Exec: "cf", Params: []string{
						"push",
						"myTestApp",
						"-f",
						"test-manifest.yml",
					}},
				}, s.Calls)

				assert.True(t, logoutCalled)
			})
		}
	})

	// TODO testCfNativeBlueGreenKeepOldInstanceShouldThrowErrorOnStopError

	t.Run("cf native deploy standard should not stop instance", func(t *testing.T) {
		config := cloudFoundryDeployOptions{
			DeployTool:      "cf_native",
			DeployType:      "standard",
			Org:             "myOrg",
			Space:           "mySpace",
			Username:        "me",
			Password:        "******",
			APIEndpoint:     "https://examples.sap.com/cf",
			Manifest:        "test-manifest.yml",
			AppName:         "myTestApp",
			KeepOldInstance: true,
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

		defer cleanup()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.NoError(t, err) {

			t.Run("check shell calls", func(t *testing.T) {
				assert.Equal(t, loginOpts,
					cloudfoundry.LoginOptions{
						CfAPIEndpoint: "https://examples.sap.com/cf",
						CfOrg:         "myOrg",
						CfSpace:       "mySpace",
						Username:      "me",
						Password:      "******",
					})

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

				assert.True(t, logoutCalled)
			})

		}
	})

	t.Run("testCfNativeWithoutAppNameBlueGreen", func(t *testing.T) {
		config := cloudFoundryDeployOptions{
			DeployTool:      "cf_native",
			DeployType:      "blue-green",
			Org:             "myOrg",
			Space:           "mySpace",
			Username:        "me",
			Password:        "******",
			APIEndpoint:     "https://examples.sap.com/cf",
			Manifest:        "test-manifest.yml",
		}

		_fileExists = func(name string) (bool, error) {
			return name == "test-manifest.yml", nil
		}

		_getManifest = func(name string) (cloudfoundry.Manifest, error) {
			return manifestMock{
					manifestFileName: "test-manifest.yml",
					apps: []map[string]interface{}{
						map[string]interface{}{
							"there-is":     "no-app-name",
						},
					},
				},
				nil
		}

		defer cleanup()

		s := mock.ExecMockRunner{}

		err := runCloudFoundryDeploy(&config, nil, &s)

		if assert.EqualError(t, err, "Blue-green plugin requires app name to be passed (see https://github.com/bluemixgaragelondon/cf-blue-green-deploy/issues/27)") {

			t.Run("check shell calls", func(t *testing.T) {

				// no login in this case
				assert.Equal(t, cloudfoundry.LoginOptions{}, loginOpts)
				// no calls to the cf client in this case
				assert.Empty(t, s.Calls)
				// no logout
				assert.False(t, logoutCalled)
			})
		}
	})

	t.Run("deploytool mtaDeployPlugin", func(t *testing.T) {

		config := cloudFoundryDeployOptions{
			DeployTool:  "mtaDeployPlugin",
			Org:         "myOrg",
			Space:       "mySpace",
			Username:    "me",
			Password:    "******",
			APIEndpoint: "https://examples.sap.com/cf",
		}

		t.Run("mta config file from project sources", func(t *testing.T) {

			defer cleanup()
			s := mock.ExecMockRunner{}
			err := runCloudFoundryDeploy(&config, nil, &s)

			if assert.NoError(t, err) {

				t.Run("check cf api calls", func(t *testing.T) {

					assert.Equal(t, s.Calls, []mock.ExecCall{
						mock.ExecCall{Exec: "cf", Params: []string{"api", "https://examples.sap.com/cf"}},
						mock.ExecCall{Exec: "cf", Params: []string{"plugins"}},
						mock.ExecCall{Exec: "cf", Params: []string{"deploy", "x.mtar"}},
					})

					t.Run("check cf login", func(t *testing.T) {
						assert.Equal(t, loginOpts,
							cloudfoundry.LoginOptions{
								CfAPIEndpoint: "https://examples.sap.com/cf",
								CfOrg:         "myOrg",
								CfSpace:       "mySpace",
								Username:      "me",
								Password:      "******",
							})
					})

					assert.True(t, logoutCalled)
				})

				t.Run("mtar retrieved", func(t *testing.T) {
					assert.True(t, mtarFileRetrieved)
				})
			}
		})

		t.Run("mta config file from project config does not exist", func(t *testing.T) {
			defer cleanup()
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

	_fileExists = func(name string) (bool, error) {
		return name == "a/varsA.txt" || name == "varsB.txt", nil
	}

	defer func() {
		_fileExists = piperutils.FileExists
	}()

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

		_glob = func(patterns []string) ([]*glob.FileAsset, []*glob.RegexpInfo, error) {
			return []*glob.FileAsset{&glob.FileAsset{Path: "x.mtar"}}, nil, nil
		}

		path, err := findMtar()

		if assert.NoError(t, err) {
			assert.Equal(t, "x.mtar", path)
		}
	})

	t.Run("No MTAR", func(t *testing.T) {

		_glob = func(patterns []string) ([]*glob.FileAsset, []*glob.RegexpInfo, error) {
			return []*glob.FileAsset{}, nil, nil
		}

		_, err := findMtar()

		assert.EqualError(t, err, "No mtar file matching pattern '**/*.mtar' found")
	})

	t.Run("Several MTARs", func(t *testing.T) {

		_glob = func(patterns []string) ([]*glob.FileAsset, []*glob.RegexpInfo, error) {
			return []*glob.FileAsset{&glob.FileAsset{Path: "x.mtar"}, &glob.FileAsset{Path: "y.mtar"}}, nil, nil
		}

		_, err := findMtar()
		assert.EqualError(t, err, "Found multiple mtar files matching pattern '**/*.mtar' (x.mtar,y.mtar), please specify file via mtaPath parameter 'mtarPath'")
	})
}
