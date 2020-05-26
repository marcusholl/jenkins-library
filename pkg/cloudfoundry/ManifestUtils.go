package cloudfoundry

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"io/ioutil"
	"fmt"
	"reflect"

	"github.com/SAP/jenkins-library/pkg/log"
)

// Manifest ...
type Manifest struct {
	self []interface{}
}

var m Manifest

var readFile = ioutil.ReadFile

const defaultManifestFileName = "manifest.yml"
const defaultManifestVariablesFileName = "manifest-variables.yml"

// ReadManifest ...
func ReadManifest(name string) (Manifest, error) {

	log.Entry().Infof("Reading manifest file  '%s'", name)

	var m Manifest
	//manifest := make(map[interface{}]interface{})
	var manifest []interface{}

	content, err := readFile(name)
	if err != nil {
		return m, errors.Wrapf(err, "cannot read file '%v'", name)
	}

	log.Entry().Infof("Manifest file content: %v", string(content))

	err = yaml.Unmarshal(content, &manifest)

	if err != nil {
		return m, errors.Wrapf(err, "Cannot parse yaml file '%v'", name)
	}

	log.Entry().Debugf("Manifest file '%s' has been unmarshalled", name)

	//d, err := yaml.Marshal(&manifest)
	if err != nil {
			fmt.Printf("Error: %v", err)
			return m, err
	}
	//ioutil.WriteFile("manifest.new.yml", d, 0644)

	log.Entry().Infof("Manifest file parsed: %v", manifest)

	m.self = manifest
	
	log.Entry().Infof("Manifest file '%s' has been read: %v", name, m.self)
	return m, err
}

// GetApplications ...
func (m Manifest)GetApplications() ([]interface{}, error) {
	return ToSlice(m.self)
}

// GetApplicationProperty ...
func (m Manifest)GetApplicationProperty(index int, name string) (string, error) {

	log.Entry().Debugf("Entering ManifestUtils.GetApplicationProperty\n")

	s, err := ToSlice(m.self)
	if err != nil {
		return "", err
	}
	_m, err := ToMap(s[index])
	if err != nil {
		return "", err
	}

	value := _m[name]

	if value != nil {
		if val, ok := value.(string); ok {
			return val, nil
		}
	}

	return "", fmt.Errorf("No such property: '%s' available in application at position %d", name, index)
}

// GetAppName ...
func (m Manifest)GetAppName() (string, error) {

	apps, err := ToSlice(m)

	if err != nil {
		return "", err
	}
	_m, err := ToMap(apps[0])

	if err != nil {
		return "", err
	}

	if name, ok := _m["Name"].(string); ok {
		return name, err		
	}

	return "", fmt.Errorf("Cannot retrieve app name. Cannot cast %v to string", _m["Name"])
}

// Transform ...
func (m Manifest)Transform() (bool, error) {
	apps, err := ToSlice(&m.self)

	if err != nil {
		return false, err
	}

	if len(apps) == 0 {
		return false, fmt.Errorf("No applications found in manifest")
	}

	for _, app := range apps {
		m, err := ToMap(app)
		if err != nil {
			return false, err
		}
		buildPacks, err := ToSlice(m["Buildpacks"])

		if err != nil {
			return false, err
		}

		if len(buildPacks) > 1 {
			return false, fmt.Errorf("More than one Cloud Foundry Buildpack is not supported. Please check your manifest file")
		}
		if len(buildPacks) == 1 {
			m["Buildpack"] = buildPacks[0]
			delete(m, "Buildpacks")
			fmt.Printf("Buildpacks: %v\n", buildPacks)
			return true, nil
		}
		fmt.Printf("No build packs found\n")
	}

	return false, nil
}

func ToMap(i interface{}) (map[interface{}]interface{}, error) {

	if m, ok := i.(map[interface{}]interface{}); ok {
			return m, nil
	}
	return nil, fmt.Errorf("Failed to convert %v to map. Was %v", i, reflect.TypeOf(i))
}

func ToSlice(i interface{}) ([]interface{}, error) {

	if s, ok := i.([]interface{}); ok {
			return s, nil
	}
	return nil, fmt.Errorf("Failed to convert %v to slice. Was %v", i, reflect.TypeOf(i))
}
