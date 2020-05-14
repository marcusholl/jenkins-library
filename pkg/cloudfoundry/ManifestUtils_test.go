package cloudfoundry

import (
	"testing"

	"fmt"
	"github.com/stretchr/testify/assert"
)

func TestReadManifest(t *testing.T) {

	readFile = func(filename string) ([]byte, error) {
		if filename == "myManifest.yaml" {
			return []byte("applications: [{name: 'manifestAppName'}]"), nil
		}
		return []byte{}, fmt.Errorf("File '%s' not found", filename)
	}

	cfManifest, err := ReadManifest("myManifest.yaml")

	assert.Equal(t, "manifestAppName", cfManifest.Applications[0].Name)
	assert.NoError(t, err)
}

func TestNoRoute(t *testing.T) {

	readFile = func(filename string) ([]byte, error) {
		if filename == "myManifest.yaml" {
			return []byte("no-route: true\napplications: [{name: 'manifestAppName'}]"), nil
		}
		return []byte{}, fmt.Errorf("File '%s' not foound", filename)
	}

	cfManifest, err := ReadManifest("myManifest.yaml")

	t.Logf("Manifest: %v", cfManifest.NoRoute)
	if assert.NoError(t, err) {
		assert.True(t, cfManifest.NoRoute)
	}
}
