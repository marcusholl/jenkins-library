package transportrequest

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUploadCTS(t *testing.T) {

	filesMock := mock.FilesMock{}
	files = &filesMock
	defer func() { files = piperutils.Files{} }()


	t.Run("all possible values provided", func(t *testing.T) {
		cmd := mock.ExecMockRunner{}
		cts := CTS{endpoint: "https://example.org:8080/cts", client: "001", user: "me", password: "******"}
		cts.Upload(&cmd, "12345678", "ui5-deploy.yaml", CTSApp{pack: "abapPackage", name: "appName", desc: "the Desc"})
		assert.Equal(t,
			[]mock.ExecCall{
				{Exec: "fiori", Params: []string{
					"deploy",
					"-f",
					"-y",
					"-e", "the Desc",
					"--noConfig",
					"-u", "https://example.org:8080/cts",
					"-l", "001",
					"-t", "12345678",
					"-p", "abapPackage",
					"-n", "appName",
				}},
			},
			cmd.Calls)
	})

	t.Run("all possible values omitted", func(t *testing.T) {
		// In this case the values are expected inside the fiori deploy config file
		cmd := mock.ExecMockRunner{}
		cts := CTS{endpoint: "", client: "", user: "me", password: "******"}
		cts.Upload(&cmd, "12345678", "ui5-deploy.yaml", CTSApp{pack: "", name: "", desc: ""})
		assert.Equal(t,
			[]mock.ExecCall{
				{Exec: "fiori", Params: []string{
					"deploy",
					"-f",
					"-y",
					"-e", "Deployed with Piper based on SAP Fiori tools",
					"--noConfig",
					"-t", "12345678",
				}},
			},
			cmd.Calls)
	})
}
