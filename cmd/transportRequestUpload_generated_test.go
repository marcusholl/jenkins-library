package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransportRequestUploadCommand(t *testing.T) {

	testCmd := TransportRequestUploadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "transportRequestUpload", testCmd.Use, "command name incorrect")

}
