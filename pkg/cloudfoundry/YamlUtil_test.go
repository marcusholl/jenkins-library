package cloudfoundry

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
	"time"
)

type fileInfoMock struct {
	name string
	data []byte
}

func (fInfo fileInfoMock) Name() string       { return "" }
func (fInfo fileInfoMock) Size() int64        { return int64(0) }
func (fInfo fileInfoMock) Mode() os.FileMode  { return 0444 }
func (fInfo fileInfoMock) ModTime() time.Time { return time.Time{} }
func (fInfo fileInfoMock) IsDir() bool        { return false }
func (fInfo fileInfoMock) Sys() interface{}   { return nil }

func TestWriteFileOnUpate(t *testing.T) {

	writeFileCalled := false

	oldStat := _stat
	oldReadFile := _readFile
	oldWriteFile := _writeFile
	oldTraverse := _traverse

	defer func() {

		_stat = oldStat
		_readFile = oldReadFile
		_writeFile = oldWriteFile
		_traverse = oldTraverse
	}()

	_stat = func(name string) (os.FileInfo, error) {
		return fileInfoMock{}, nil
	}

	_readFile = func(name string) ([]byte, error) {
		if name == "manifest.yml" || name == "replacements.yml" {
			return []byte{}, nil
		}
		return []byte{}, fmt.Errorf("open %s: no such file or directory", name)
	}

	_writeFile = func(name string, data []byte, mode os.FileMode) error {
		writeFileCalled = true
		return nil
	}

	_traverse = func(interface{}, map[string]interface{}) (interface{}, bool, error) {
		return nil, true, nil
	}

	updated, err := Substitute("manifest.yml", "replacements.yml")

	if assert.NoError(t, err) {
		assert.True(t, updated)
		assert.True(t, writeFileCalled)
	}
}

func TestDontWriteFileOnNoUpate(t *testing.T) {

	writeFileCalled := false

	oldStat := _stat
	oldReadFile := _readFile
	oldWriteFile := _writeFile
	oldTraverse := _traverse

	defer func() {

		_stat = oldStat
		_readFile = oldReadFile
		_writeFile = oldWriteFile
		_traverse = oldTraverse
	}()

	_stat = func(name string) (os.FileInfo, error) {
		return fileInfoMock{}, nil
	}

	_readFile = func(name string) ([]byte, error) {
		if name == "manifest.yml" || name == "replacements.yml" {
			return []byte{}, nil
		}
		return []byte{}, fmt.Errorf("open %s: no such file or directory", name)
	}

	_writeFile = func(name string, data []byte, mode os.FileMode) error {
		writeFileCalled = true
		return nil
	}

	_traverse = func(interface{}, map[string]interface{}) (interface{}, bool, error) {
		return nil, false, nil
	}

	updated, err := Substitute("manifest.yml", "replacements.yml")

	if assert.NoError(t, err) {
		assert.False(t, updated)
		assert.False(t, writeFileCalled)
	}
}

func TestManifestFileDoesNotExist(t *testing.T) {

	writeFileCalled := false
	traverseCalled := false

	oldStat := _stat
	oldReadFile := _readFile
	oldWriteFile := _writeFile
	oldTraverse := _traverse

	defer func() {

		_stat = oldStat
		_readFile = oldReadFile
		_writeFile = oldWriteFile
		_traverse = oldTraverse
	}()

	_stat = func(name string) (os.FileInfo, error) {
		return fileInfoMock{}, nil
	}

	_readFile = func(name string) ([]byte, error) {
		if name == "manifest.yml" || name == "replacements.yml" {
			return []byte{}, nil
		}
		return []byte{}, fmt.Errorf("open %s: no such file or directory", name)
	}

	_writeFile = func(name string, data []byte, mode os.FileMode) error {
		writeFileCalled = true
		return nil
	}

	_traverse = func(interface{}, map[string]interface{}) (interface{}, bool, error) {
		traverseCalled = true
		return nil, false, nil
	}

	_, err := Substitute("manifestDoesNotExist.yml", "replacements.yml")

	if assert.EqualError(t, err, "open manifestDoesNotExist.yml: no such file or directory") {
		assert.False(t, writeFileCalled)
		assert.False(t, traverseCalled)
	}
}

func TestReplacementsFileDoesNotExist(t *testing.T) {

	writeFileCalled := false
	traverseCalled := false

	oldStat := _stat
	oldReadFile := _readFile
	oldWriteFile := _writeFile
	oldTraverse := _traverse

	defer func() {

		_stat = oldStat
		_readFile = oldReadFile
		_writeFile = oldWriteFile
		_traverse = oldTraverse
	}()

	_stat = func(name string) (os.FileInfo, error) {
		return fileInfoMock{}, nil
	}

	_readFile = func(name string) ([]byte, error) {
		if name == "manifest.yml" || name == "replacements.yml" {
			return []byte{}, nil
		}
		return []byte{}, fmt.Errorf("open %s: no such file or directory", name)
	}

	_writeFile = func(name string, data []byte, mode os.FileMode) error {
		writeFileCalled = true
		return nil
	}

	_traverse = func(interface{}, map[string]interface{}) (interface{}, bool, error) {
		traverseCalled = true
		return nil, false, nil
	}

	_, err := Substitute("manifest.yml", "replacementsDoesNotExist.yml")

	if assert.EqualError(t, err, "open replacementsDoesNotExist.yml: no such file or directory") {
		assert.False(t, writeFileCalled)
		assert.False(t, traverseCalled)
	}
}

func TestXX(t *testing.T) {

	document := make(map[string]interface{})
	replacements := make(map[string]interface{})

	yaml.Unmarshal([]byte(
		`unique-prefix: uniquePrefix # A unique prefix. E.g. your D/I/C-User
xsuaa-instance-name: uniquePrefix-catalog-service-odatav2-xsuaa
hana-instance-name: uniquePrefix-catalog-service-odatav2-hana
integer-variable: 1
boolean-variable: Yes
float-variable: 0.25
json-variable: >
  [
    {"name":"token-destination",
     "url":"https://www.google.com",
     "forwardAuthToken": true}
  ]
object-variable:
  hello: "world"
  this:  "is an object with"
  one: 1
  float: 25.0
  bool: Yes`), &replacements)

	err := yaml.Unmarshal([]byte(
		`applications:
- name: ((unique-prefix))-catalog-service-odatav2-0.0.1
  memory: 1024M
  disk_quota: 512M
  instances: ((integer-variable))
  buildpacks:
    - java_buildpack
  path: ./srv/target/srv-backend-0.0.1-SNAPSHOT.jar
  routes:
  - route: ((unique-prefix))-catalog-service-odatav2-001.cfapps.eu10.hana.ondemand.com

  services:
  - ((xsuaa-instance-name)) # requires an instance of xsuaa instantiated with xs-security.json of this project. See services-manifest.yml.
  - ((hana-instance-name))  # requires an instance of hana service with plan hdi-shared. See services-manifest.yml.

  env:
    spring.profiles.active: cloud # activate the spring profile named 'cloud'.
    xsuaa-instance-name: ((xsuaa-instance-name))
    db_service_instance_name: ((hana-instance-name))
    booleanVariable: ((boolean-variable))
    floatVariable: ((float-variable))
    json-variable: ((json-variable))
    object-variable: ((object-variable))
    string-variable: ((boolean-variable))-((float-variable))-((integer-variable))-((json-variable))
    single-var-with-string-constants: ((boolean-variable))-with-some-more-text
  `), &document)

	replaced, updated, err := traverse(document, replacements)

	assert.NoError(t, err)
	assert.True(t, updated)
	//
	// assertDataTypeAndSubstitutionCorrectness start

	if m, ok := replaced.(map[string]interface{}); ok {

		if apps, ok := m["applications"].([]interface{}); ok {
			app := apps[0]
			if appAsMap, ok := app.(map[string]interface{}); ok {

				instances := appAsMap["instances"]

				if one, ok := instances.(float64); ok {
					assert.Equal(t, 1, int(one))
				}

				if services, ok := appAsMap["services"]; ok {
					if servicesAsSlice, ok := services.([]interface{}); ok {
						if _, ok := servicesAsSlice[0].(string); ok {
							assert.True(t, true)
						} else {
							assert.True(t, false)
						}
					} else {
						assert.True(t, false)
					}
				} else {
					assert.True(t, false)
				}

				if env, ok := appAsMap["env"]; ok {

					if envAsMap, ok := env.(map[string]interface{}); ok {

						if _, ok := envAsMap["floatVariable"].(float64); ok {

							assert.True(t, true)
						} else {
							assert.True(t, false)
						}

						if asBoolean, ok := envAsMap["booleanVariable"].(bool); ok {
							assert.True(t, asBoolean)
						} else {
							assert.True(t, false)
						}

						if _, ok := envAsMap["json-variable"].(string); ok {
							assert.True(t, true)
						} else {
							assert.True(t, false)
						}

						if _, ok := envAsMap["object-variable"].(map[string]interface{}); ok {
							assert.True(t, true)
						} else {
							assert.True(t, false)
						}

						if s, ok := envAsMap["string-variable"].(string); ok {
							assert.True(t, strings.HasPrefix(s, "true-0.25-1-"))
						} else {
							assert.True(t, false)
						}

						if s, ok := envAsMap["single-var-with-string-constants"].(string); ok {
							assert.Equal(t, s, "true-with-some-more-text")
						} else {
							assert.True(t, false)
						}

					} else {
						assert.True(t, false)
					}
				}
			}
		}

		//assertDataTypeAndSubstitutionCorrectness END
		//

		//
		// assertCorrectVariableResolution START

		if m, ok := replaced.(map[string]interface{}); ok {
			if apps, ok := m["applications"].([]interface{}); ok {
				app := apps[0]
				if appAsMap, ok := app.(map[string]interface{}); ok {

					assert.Equal(t, "uniquePrefix-catalog-service-odatav2-0.0.1", appAsMap["name"])

					if env, ok := appAsMap["env"]; ok {

						if envAsMap, ok := env.(map[string]interface{}); ok {
							assert.Equal(t, "uniquePrefix-catalog-service-odatav2-xsuaa", envAsMap["xsuaa-instance-name"])
							assert.Equal(t, "uniquePrefix-catalog-service-odatav2-hana", envAsMap["db_service_instance_name"])
						} else {
							assert.True(t, false)
						}
						if servicesAsSlice, ok := appAsMap["services"].([]interface{}); ok {
							assert.Equal(t, "uniquePrefix-catalog-service-odatav2-xsuaa", servicesAsSlice[0])
							assert.Equal(t, "uniquePrefix-catalog-service-odatav2-hana", servicesAsSlice[1])
						} else {
							assert.True(t, false)
						}
					}
				}
			}
		}

		//
		// assertCorrectVariableResolution END

	}

	data, err := yaml.Marshal(&replaced)

	t.Logf("Data: %v", string(data))

	assert.NoError(t, err)

	assert.True(t, true, "Everything is fine")
}
