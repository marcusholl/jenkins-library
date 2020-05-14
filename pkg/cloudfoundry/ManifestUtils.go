package cloudfoundry

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"io/ioutil"
)

var readFile = ioutil.ReadFile

// Application ...
type Application struct {
	Name string
}

// CFManifest ...
type CFManifest struct {
	Applications []Application
	NoRoute      bool `json:"no-route"` // TODO: revisit if no-route is configued here on that level
}

// ReadManifest ...
func ReadManifest(name string) (CFManifest, error) {

	var manifest CFManifest

	content, err := readFile(name)
	if err != nil {
		return manifest, errors.Wrapf(err, "cannot read file '%v'", name)
	}

	err = yaml.Unmarshal(content, &manifest)
	if err != nil {
		err = errors.Wrapf(err, "Cannot parse yaml file '%v'", name)
	}

	return manifest, err
}
