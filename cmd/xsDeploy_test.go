package cmd

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestXSLogin(t *testing.T) {

	myXsDeployOptions := xsDeployOptions{
		APIURL:        "https://example.org:12345",
		User:          "me",
		Password:      "secret",
		Org:           "myOrg",
		Space:         "mySpace",
		LoginOpts:     "--skip-ssl-validation",
		XsSessionFile: ".xs_session",
	}

	t.Run("Login succeeds", func(t *testing.T) {

		s := shellMockRunner{}

		e := xsLogin(myXsDeployOptions, &s)

		if e != nil {
			t.Errorf("XSDeploy command failed: %v", e)
		}

		cmds := strings.Split(s.calls[0], "\n")
		assert.Equal(t, cmds[0], "#!/bin/bash")
		assert.Contains(t, cmds[1], "xs login -a https://example.org:12345 -u me -p 'secret' -o myOrg -s mySpace --skip-ssl-validation")
	})

	t.Run("Login fails", func(t *testing.T) {

		s := shellMockRunner{shouldFailWith: errors.New("cmd failed")}

		e := xsLogin(myXsDeployOptions, &s)

		if e == nil {
			t.Errorf("Expected error not seen")
		}
	})

}

func TestXSLogout(t *testing.T) {

	myXsDeployOptions := xsDeployOptions{
		APIURL:        "https://example.org:12345",
		User:          "me",
		Password:      "secret",
		Org:           "myOrg",
		Space:         "mySpace",
		LoginOpts:     "--skip-ssl-validation",
		XsSessionFile: ".xs_session",
	}

	t.Run("Success case", func(t *testing.T) {

		s := shellMockRunner{}

		e := xsLogout(myXsDeployOptions, &s)

		if e != nil {
			t.Errorf("XSDeploy command failed: %v", e)
		}

		cmds := strings.Split(s.calls[0], "\n")
		assert.Equal(t, cmds[0], "#!/bin/bash")
		assert.Contains(t, cmds[1], "xs logout")
	})

	t.Run("Logout fails", func(t *testing.T) {

		s := shellMockRunner{shouldFailWith: errors.New("xs logout failed")}

		e := xsLogout(myXsDeployOptions, &s)

		if e == nil {
			t.Error("XSDeploy: Expected logout error not received.")
		}
	})
}
