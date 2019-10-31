package cmd

import (
	"strings"
	"testing"
	//"github.com/stretchr/testify/assert"
)

func TestXSLogin(t *testing.T) {

	s := shellMockRunner{}
	var myXsDeployOptions xsDeployOptions {apiUrl: "https://example.org:12345",
		user: "me",
		password: "secret",
		org: "myOrg",
		space: "mySpace",
		loginOpts: "--skip-ssl-validation",
		xsSessionFile: ".xs_session"
	}

	e := xsLogin(myXsDeployOptions, &s, func(f string) bool {
		return f == "xyz"
	})
	if e != nil {
		t.Errorf("XSDeploy command failed: %v", e)
	}

	t.Log("XXXX: " + strings.Join(s.calls, ", "))
}
