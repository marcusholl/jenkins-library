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

func TestTransformGoodCase(t *testing.T) {

	readFile = func(filename string) ([]byte, error) {
		if filename == "myManifest.yaml" {
			return []byte("no-route: true\napplications: [{name: 'manifestAppName', buildPacks: [sap_java_buildpack]}]"), nil
		}
		return []byte{}, fmt.Errorf("File '%s' not foound", filename)
	}

	cfManifest, err := ReadManifest("myManifest.yaml")
	assert.NoError(t, err)

	changed, err := Transform(&cfManifest)

	assert.NoError(t, err)
	assert.Equal(t, "sap_java_buildpack", cfManifest.Applications[0].Buildpack)
	assert.Equal(t, []string{}, cfManifest.Applications[0].Buildpacks)
	assert.True(t, changed)

}

func TestTransformMultipleBuildPacks(t *testing.T) {
	readFile = func(filename string) ([]byte, error) {
		if filename == "myManifest.yaml" {
			return []byte("no-route: true\napplications: [{name: 'manifestAppName', buildPacks: [sap_java_buildpack, 'another_buildpack']}]"), nil
		}
		return []byte{}, fmt.Errorf("File '%s' not foound", filename)
	}

	cfManifest, err := ReadManifest("myManifest.yaml")
	assert.NoError(t, err)

	_, err = Transform(&cfManifest)

	assert.EqualError(t, err, "More than one Cloud Foundry Buildpack is not supported. Please check your manifest file")
}

func TestTransformUnchanged(t *testing.T) {
	readFile = func(filename string) ([]byte, error) {
		if filename == "myManifest.yaml" {
			return []byte("no-route: true\napplications: [{name: 'manifestAppName', buildPack: sap_java_buildpack}]"), nil
		}
		return []byte{}, fmt.Errorf("File '%s' not foound", filename)
	}

	cfManifest, err := ReadManifest("myManifest.yaml")
	assert.NoError(t, err)

	changed, err := Transform(&cfManifest)

	assert.NoError(t, err)
	assert.Equal(t, "sap_java_buildpack", cfManifest.Applications[0].Buildpack)
	assert.Nil(t, cfManifest.Applications[0].Buildpacks)
	assert.False(t, changed)
}