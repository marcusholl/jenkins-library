package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransportRequestUploadCTSCommand(t *testing.T) {

	testCmd := TransportRequestUploadCTSCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "transportRequestUploadCTS", testCmd.Use, "command name incorrect")

}
