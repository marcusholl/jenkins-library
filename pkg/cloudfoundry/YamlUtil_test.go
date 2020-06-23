package cloudfoundry

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"os"
	"regexp"
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

type fMock struct {
	files map[string][]byte
}

func (f *fMock) ReadFile(name string) ([]byte, error) {

	result := f.files[name]

	if result != nil {
		return result, nil
	}
	return []byte{}, fmt.Errorf("open %s: no such file or directory", name)
}

func TestFilesRelated(t *testing.T) {

	writeFileCalled := false
	traverseCalled := false

	var replacements map[string]interface{}

	oldStat := _stat
	oldWriteFile := _writeFile
	oldTraverse := _traverse

	_fileUtils = &fMock{}

	defer func() {
		_stat = oldStat
		_writeFile = oldWriteFile
		_traverse = oldTraverse
		_fileUtils = &myutil{}
	}()

	reset := func() {

		writeFileCalled = false

		traverseCalled = false

		replacements = make(map[string]interface{})

		_stat = func(name string) (os.FileInfo, error) {
			return fileInfoMock{}, nil
		}

		_fileUtils = &fMock {
			files: map[string][]byte {
				"manifest.yml": []byte("a: dummy"),
				"replacements.yml": []byte{}},
		}

		_writeFile = func(name string, data []byte, mode os.FileMode) error {
			writeFileCalled = true
			return nil
		}

		_traverse = func(_ interface{}, _replacements map[string]interface{}) (interface{}, bool, error) {
			replacements = _replacements
			traverseCalled = true
			return nil, true, nil
		}
	}

	reset()

	t.Run("WriteFileOnUpdate", func(t *testing.T) {

		defer reset()

		updated, err := Substitute("manifest.yml", map[string]interface{}{}, []string{"replacements.yml"})

		if assert.NoError(t, err) {
			assert.True(t, updated)
			assert.True(t, writeFileCalled)
			assert.True(t, traverseCalled)
		}
	})

	t.Run("DontWriteOnNoUpdate", func(t *testing.T) {

		_traverse = func(interface{}, map[string]interface{}) (interface{}, bool, error) {
			traverseCalled = true
			return nil, false, nil
		}

		defer reset()

		updated, err := Substitute("manifest.yml", map[string]interface{}{}, []string{"replacements.yml"})

		if assert.NoError(t, err) {
			assert.False(t, updated)
			assert.True(t, traverseCalled)
			assert.False(t, writeFileCalled)
		}
	})

	t.Run("Read multiple replacement yamls in one file", func(t *testing.T) {

		// expected behaviour in case of multiple yaml documents in one "file":
		// we merge the content. The latest wins

		_fileUtils = &fMock {
			files: map[string][]byte {
				"manifest.yml": []byte("a: dummy"),
				"replacements.yml": []byte("a: b # A comment.\nc: d\n---\nzz: 1234\n"),
			},
		}

		defer reset()

		_, err := Substitute("manifest.yml", map[string]interface{}{}, []string{"replacements.yml"})

		if assert.NoError(t, err) {
			assert.Equal(t, map[string]interface{}{"a": "b", "c": "d", "zz": 1234}, replacements)
		}
	})

	t.Run("Handle multi manifest", func(t *testing.T) {

		var written string

		_writeFile = func(name string, data []byte, mode os.FileMode) error {
			written = string(data)
			return nil
		}

		_traverse = func(_ interface{}, _replacements map[string]interface{}) (interface{}, bool, error) {
			return map[string]interface{}{"called": true}, true, nil
		}

		_fileUtils = &fMock {
			files: map[string][]byte {
				"manifest.yml": []byte("a: dummy\n---\n b: otherDummy\n"),
				// here we have two yaml documents in one "file" ...
				"replacements.yml": []byte("a: b # A comment.\nc: d\n---\nzz: 1234\n"),
			},
		}


		defer reset()

		_, err := Substitute("manifest.yml", map[string]interface{}{}, []string{"replacements.yml"})

		if assert.NoError(t, err) {
			// ... the two yaml files results in two yaml documents, separated by '---'
			assert.Equal(t, "called: true\n---\ncalled: true\n", written)
		}
	})

	t.Run("Handle single manifest", func(t *testing.T) {


		_fileUtils = &fMock {
			files: map[string][]byte {
				"manifest.yml": []byte("a: dummy\n"),
				// here we have two yaml documents in one "file" ...
				"replacements.yml": []byte("a: b # A comment.\nc: d\n---\nzz: 1234\n"),
			},
		}

		var written string

		_writeFile = func(name string, data []byte, mode os.FileMode) error {
			written = string(data)
			return nil
		}

		_traverse = func(_ interface{}, _replacements map[string]interface{}) (interface{}, bool, error) {
			return map[string]interface{}{"called": true}, true, nil
		}

		defer reset()

		_, err := Substitute("manifest.yml", map[string]interface{}{}, []string{"replacements.yml"})

		if assert.NoError(t, err) {
			// we have a single yaml document (no '---' inbetween)
			assert.Equal(t, "called: true\n", written)
		}
	})

	t.Run("Manifest does not exist", func(t *testing.T) {

		defer reset()

		_, err := Substitute("manifestDoesNotExist.yml", map[string]interface{}{}, []string{"replacements.yml"})

		if assert.EqualError(t, err, "open manifestDoesNotExist.yml: no such file or directory") {
			assert.False(t, writeFileCalled)
			assert.False(t, traverseCalled)
		}
	})

	t.Run("Replacements does not exist", func(t *testing.T) {

		defer reset()

		_, err := Substitute("manifest.yml", map[string]interface{}{}, []string{"replacementsDoesNotExist.yml"})

		if assert.EqualError(t, err, "open replacementsDoesNotExist.yml: no such file or directory") {
			assert.False(t, writeFileCalled)
			assert.False(t, traverseCalled)
		}
	})

	t.Run("Replacements from map has precedence over replacments from file", func(t *testing.T) {

		_fileUtils = &fMock {
			files: map[string][]byte {
				"manifest.yml": []byte("a: ((a))\nb: ((b))"),
				// here we have two yaml documents in one "file" ...
				"replacements.yml": []byte("a: aa # A comment.\nb: bb\n"),
			},
		}

		var written string

		_writeFile = func(name string, data []byte, mode os.FileMode) error {
			written = string(data)
			return nil
		}

		_traverse = traverse

		defer reset()

		_, err := Substitute("manifest.yml", map[string]interface{}{"b": "xx"}, []string{"replacements.yml"})

		if assert.NoError(t, err) {
			assert.Equal(t, "a: aa\nb: xx\n", written)
		}
	})
}

func TestSubstitution(t *testing.T) {

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

	t.Run("The basics", func(t *testing.T) {
		assert.NoError(t, err)
		assert.True(t, updated)
	})

	t.Run("Check no variables left", func(t *testing.T) {

		data, err := yaml.Marshal(&replaced)

		if assert.NoError(t, err) {
			assert.Nil(t, regexp.MustCompile("\\(\\(.*\\)\\)").Find(data))
		}
	})

	app := getApplication(t, replaced, 0)

	t.Run("Assert data type and substitution correctness", func(t *testing.T) {

		t.Run("Check instances", func(t *testing.T) {
			if one, ok := app["instances"].(float64); ok {
				assert.Equal(t, 1, int(one))
			} else {
				assert.Fail(t, "Value for 'instances' is not a float64")
			}
		})

		t.Run("Check services", func(t *testing.T) {
			if services, ok := app["services"].([]interface{}); ok {
				assert.IsType(t, "string", services[0])
			} else {
				assert.FailNow(t, "first service is not of type string.")
			}
		})

		t.Run("Assert correct variable substitution - name", func(t *testing.T) {
			assert.Equal(t, "uniquePrefix-catalog-service-odatav2-0.0.1", app["name"])
		})

		t.Run("Assert correct variable substitution - services", func(t *testing.T) {
			if services, ok := app["services"].([]interface{}); ok {
				assert.Equal(t, "uniquePrefix-catalog-service-odatav2-xsuaa", services[0])
				assert.Equal(t, "uniquePrefix-catalog-service-odatav2-hana", services[1])
			} else {
				assert.Fail(t, "'services' node is not a list")
			}
		})

		t.Run("Check env", func(t *testing.T) {

			if env, ok := app["env"].(map[string]interface{}); ok {

				t.Run("Check float", func(t *testing.T) {
					assert.IsType(t, float64(42), env["floatVariable"])
				})

				t.Run("Check boolean", func(t *testing.T) {
					assert.IsType(t, true, env["booleanVariable"])
				})

				t.Run("Check json variable is string", func(t *testing.T) {
					assert.IsType(t, "string", env["json-variable"])
				})

				t.Run("Check object variable is map", func(t *testing.T) {
					assert.IsType(t, make(map[string]interface{}), env["object-variable"])
				})

				t.Run("Check string variable (composed 1)", func(t *testing.T) {
					assert.IsType(t, "string", env["string-variable"])
					assert.Regexp(t, "^true-0.25-1-.*", env["string-variable"])
				})

				t.Run("Check string variable (composed 2)", func(t *testing.T) {
					assert.Equal(t, "true-with-some-more-text", env["single-var-with-string-constants"])
				})

				t.Run("Assert correct variable substitution - xsuaa-instance-name", func(t *testing.T) {
					assert.Equal(t, "uniquePrefix-catalog-service-odatav2-xsuaa", env["xsuaa-instance-name"])
				})

				t.Run("Assert correct variable substitution - db_service_instance_name", func(t *testing.T) {
					assert.Equal(t, "uniquePrefix-catalog-service-odatav2-hana", env["db_service_instance_name"])
				})

			} else {
				assert.FailNow(t, "env is not a map")
			}
		})
	})
}

func getApplication(t *testing.T, tree interface{}, index int) map[string]interface{} {

	const nodeNameApplications = "applications"

	if m, ok := tree.(map[string]interface{}); !ok {
		assert.FailNow(t, "Cannot prepare tests", "Outermost node inside replaced structure is not a map.")
	} else {
		if apps, ok := m[nodeNameApplications].([]interface{}); !ok {
			assert.FailNowf(t, "Cannot prepare tests", "Node '%s' is not an interface slice.", nodeNameApplications)
		} else {
			if app, ok := apps[0].(map[string]interface{}); ok {
				return app
			} else {
				assert.FailNowf(t, "Cannot prepare tests", "The first node inside '%s' is not a map.", nodeNameApplications)
			}
		}
	}
	return nil // we should end up earlier in an assert or return above ...
}
