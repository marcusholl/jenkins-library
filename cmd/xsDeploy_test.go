package cmd

import (
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestXSLogin(t *testing.T) {

	myXsDeployOptions := xsDeployOptions{
		APIURL: "https://example.org:12345",
		User: "me",
		Password: "secret",
		Org: "myOrg",
		Space: "mySpace",
		LoginOpts: "--skip-ssl-validation",
		XsSessionFile: ".xs_session",
	}
	t.Run("No xs session file", func(t *testing.T) {
		s := shellMockRunner{}

		e := xsLogin(myXsDeployOptions, &s, func(f string) bool {
			return false
		})

		if e == nil {
			t.Error("Missing xs session file not detected")
		} else if e.Error() != "file does not exist (.xs_session)" {
			t.Errorf("Failed with unexpected error: '%v'", e)
		}
	})

	t.Run("Success case", func(t *testing.T) {
		s := shellMockRunner{}

		e := xsLogin(myXsDeployOptions, &s, func(f string) bool {
			return f == ".xs_session"
		})

		if e != nil {
			t.Errorf("XSDeploy command failed: %v", e)
		}

		cmds := strings.Split(s.calls[0], "\n")
		assert.Equal(t, cmds[0], "#!/bin/bash")
		assert.Contains(t, cmds[1], "xs login -a https://example.org:12345 -u me -p 'secret' -o myOrg -s mySpace --skip-ssl-validation")
		assert.Contains(t, cmds[3], "cp \"${HOME}/.xs_session\" .")
	})

}
