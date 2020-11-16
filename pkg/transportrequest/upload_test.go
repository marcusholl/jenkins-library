package transportrequest

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUploadCTS(t *testing.T) {

	cts := CTS{endpoint: "https://example.org:8080/cts", client: "001", user: "me", password: "******"}
	cmd := mock.ExecMockRunner{}
	cts.Upload(&cmd, "12345678", "abapPackage", "appName")
	assert.Equal(t,
		[]mock.ExecCall{
			{Exec: "fiori", Params: []string{
				"deploy",
				"-f",
				"-y",
				"-u", "https://example.org:8080/cts",
				"-l", "001",
				"-t", "12345678",
				"-p", "abapPackage",
				"-n", "appName"}},
		},
		cmd.Calls)

	t.Log(cmd.Calls[0])
}