package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelloWorldCommand(t *testing.T) {

	testCmd := HelloWorldCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "helloWorld", testCmd.Use, "command name incorrect")

}
