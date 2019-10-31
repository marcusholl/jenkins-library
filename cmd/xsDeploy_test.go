package cmd

import (
	"strings"
	"testing"
	//"github.com/stretchr/testify/assert"
)

func TestXSLogin(t *testing.T) {

	s := shellMockRunner{}
	var myXsDeployOptions xsDeployOptions
	e := xsLogin(myXsDeployOptions, &s, func(f string) bool {
		return f == "xyz"
	})
	if e != nil {
		t.Errorf("XSDeploy command failed: %v", e)
	}

	t.Log("XXXX: " + strings.Join(s.calls, ", "))
}
