package transportrequest

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUploadCTS(t *testing.T) {

	cts := CTS{endpoint: "https://example.org:8080/cts", client: "001", user: "me", password: "******"}
	cmd := mock.ExecMockRunner{}
	cts.Upload(&cmd)
	assert.Equal(t,
		[]mock.ExecCall{
			{Exec: "fiori", Params: []string{"deploy"}},
		},
		cmd.Calls)

	t.Log(cmd.Calls[0])
}