package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransportRequestUploadCTSXXXCommand(t *testing.T) {
	t.Parallel()

	testCmd := TransportRequestUploadCTSXXXCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "transportRequestUploadCTSXXX", testCmd.Use, "command name incorrect")

}
