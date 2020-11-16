package transportrequest

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUploadCTS(t *testing.T) {

	t.Run("all possible values provided", func(t *testing.T) {
		cts := CTS{endpoint: "https://example.org:8080/cts", client: "001", user: "me", password: "******"}
		cmd := mock.ExecMockRunner{}
		cts.Upload(&cmd, "12345678", CTSApp{pack: "abapPackage", name: "appName", desc: "the Desc"})
		assert.Equal(t,
			[]mock.ExecCall{
				{Exec: "fiori", Params: []string{
					"deploy",
					"-f",
					"-y",
					"-e", "the Desc",
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
		cts := CTS{endpoint: "", client: "", user: "me", password: "******"}
		cmd := mock.ExecMockRunner{}
		cts.Upload(&cmd, "12345678", CTSApp{pack: "", name: "", desc: ""})
		assert.Equal(t,
			[]mock.ExecCall{
				{Exec: "fiori", Params: []string{
					"deploy",
					"-f",
					"-y",
					"-e", "Deployed with Piper based on SAP Fiori tools",
					"-t", "12345678",
				}},
			},
			cmd.Calls)
	})

}
