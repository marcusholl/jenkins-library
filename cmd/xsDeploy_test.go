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

	t.Run("No xs session file", func(t *testing.T) {

		s := shellMockRunner{}

		e := xsLogin(myXsDeployOptions, &s, func(f string) bool {
			return false
		}, func(src, dest string) (int64, error) {
			return 0, nil
		})

		if e == nil {
			t.Error("Missing xs session file not detected")
		} else if e.Error() != "xs session file does not exist (.xs_session)" {
			t.Errorf("Failed with unexpected error: '%v'", e)
		}
	})

	t.Run("Login failure", func(t *testing.T) {

		s := shellMockRunner{shouldFailWith: errors.New("xs login failed")}

		e := xsLogin(myXsDeployOptions, &s, func(f string) bool {
			return f == ".xs_session"
		}, func(src, dest string) (int64, error) {
			return 0, nil
		})

		if e != nil && e.Error() != "xs login failed" {
			t.Errorf("Exception exception not seen. Instead we got: '%s'", e.Error())
		}

		if e == nil {
			t.Error("Login failure expected, but not seen." + e.Error())
		}
	})

	t.Run("Success case", func(t *testing.T) {

		s := shellMockRunner{}

		e := xsLogin(myXsDeployOptions, &s, func(f string) bool {
			return f == ".xs_session"
		}, func(src, dest string) (int64, error) {
			return 0, nil
		})

		if e != nil {
			t.Errorf("XSDeploy command failed: %v", e)
		}

		cmds := strings.Split(s.calls[0], "\n")
		assert.Equal(t, cmds[0], "#!/bin/bash")
		assert.Contains(t, cmds[1], "xs login -a https://example.org:12345 -u me -p 'secret' -o myOrg -s mySpace --skip-ssl-validation")
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

		e := xsLogout(myXsDeployOptions, &s,

			func(f string) bool {
				return f == ".xs_session"
			},

			func(src, dest string) (int64, error) {
				return 0, nil
			},

			func(f string) error {
				return nil
			},
		)

		if e != nil {
			t.Errorf("XSDeploy command failed: %v", e)
		}

		cmds := strings.Split(s.calls[0], "\n")
		assert.Equal(t, cmds[0], "#!/bin/bash")
		assert.Contains(t, cmds[2], "xs logout")
	})

	t.Run("Logout fails", func(t *testing.T) {

		s := shellMockRunner{shouldFailWith: errors.New("xs logout failed")}

		e := xsLogout(myXsDeployOptions, &s,

			func(f string) bool {
				return f == ".xs_session"
			},

			func(src, dest string) (int64, error) {
				return 0, nil
			},

			func(f string) error {
				return nil
			},
		)

		if e == nil {
			t.Error("XSDeploy: Expected logout error not received.")
		}
	})

	t.Run("xssession file delete fails", func(t *testing.T) {

		s := shellMockRunner{}

		e := xsLogout(myXsDeployOptions, &s,

			func(f string) bool {
				return f == ".xs_session"
			},

			func(src, dest string) (int64, error) {
				return 0, nil
			},

			func(f string) error {
				return errors.New("Cannot delete xs session file")
			},
		)

		if e == nil {
			t.Error("XSDeploy: Expected logout error not received.")
		} else if e.Error() != "Cannot delete xs session file" {
			t.Errorf("Failed with unexpected error: '%v'", e)
		}
	})

	t.Run("xssession file not present prior to logout", func(t *testing.T) {

		s := shellMockRunner{}

		e := xsLogout(myXsDeployOptions, &s,

			func(f string) bool {
				return false
			},

			func(src, dest string) (int64, error) {
				return 0, nil
			},

			func(f string) error {
				return nil
			},
		)

		if e == nil {
			t.Error("XSDeploy: Expected logout error not received.")
		} else if e.Error() != "xs session file does not exist (.xs_session)" {
			t.Errorf("Failed with unexpected error: '%v'", e)
		}
	})

}
