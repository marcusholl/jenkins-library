package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/transportrequest"
	"github.com/stretchr/testify/assert"
	"testing"
)

type CTSUploadActionMock struct {
	transportrequest.CTSUploadAction
}

func (action CTSUploadActionMock) Perform(utils transportRequestUploadMockUtils) error {
	return nil
}

type transportRequestUploadMockUtils struct {
	*mock.ShellMockRunner
	*ActionProvider
}

func newTransportRequestUploadTestsUtils() transportRequestUploadMockUtils {
	utils := transportRequestUploadMockUtils{
		ShellMockRunner: &mock.ShellMockRunner{},
		ActionProvider:  &ActionProvider{action: transportrequest.CTSUploadAction{}},
	}
	return utils
}

func TestRunTransportRequestUpload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		// init
		config := transportRequestUploadOptions{}
		config.Endpoint = "https://example.org:8000"
		config.Client = "001"
		config.Username = "me"
		config.Password = "********"
		config.ApplicationName = "myApp"
		config.AbapPackage = "myPackage"
		config.Description = "lorem ipsum"
		config.TransportRequestID = "XXXK123456"
		config.OsDeployUser = "node"                // default provided in config
		config.DeployConfigFile = "ui5-deploy.yaml" // default provided in config
		config.DeployToolDependencies = []string{"@ui5/cli", "@sap/ux-ui5-tooling"}
		config.NpmInstallOpts = []string{"--verbose", "--registry", "https://registry.example.org/"}
		utils := newTransportRequestUploadTestsUtils()

		// test
		err := runTransportRequestUpload(&config, nil, utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t, transportrequest.CTSUploadAction{
				Connection: transportrequest.CTSConnection{
					Endpoint: "https://example.org:8000",
					Client:   "001",
					User:     "me",
					Password: "********",
				},
				Application: transportrequest.CTSApplication{
					Name: "myApp",
					Pack: "myPackage",
					Desc: "lorem ipsum",
				},
				Node: transportrequest.CTSNode{
					DeployDependencies: []string{
						"@ui5/cli",
						"@sap/ux-ui5-tooling",
					},
					InstallOpts: []string{
						"--verbose",
						"--registry",
						"https://registry.example.org/",
					},
				},
				TransportRequestID: "XXXK123456",
				ConfigFile:         "ui5-deploy.yaml",
				DeployUser:         "node",
			}, *utils.GetAction())
		}
	})
}
