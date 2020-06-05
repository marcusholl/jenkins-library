package cmd

import (
	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/godo.v2/glob"
	"testing"
)

func TestCfDeployment(t *testing.T) {

	var loginOpts cloudfoundry.LoginOptions
	var logoutCalled, mtarFileRetrieved bool

	var cleanup = func() {
		loginOpts = cloudfoundry.LoginOptions{}
		logoutCalled = false
		mtarFileRetrieved = false
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
