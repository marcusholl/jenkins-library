package cloudfoundry

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"io/ioutil"
	"fmt"
)

var readFile = ioutil.ReadFile

const defaultManifestFileName = "manifest.yml"
const defaultManifestVariablesFileName = "manifest-variables.yml"

// Application ...
type Application struct {
	Name string
	Buildpacks []string
	Buildpack string
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

// Transform ...
func Transform(manifest *CFManifest) (bool, error) {
	if len(manifest.Applications) == 0 {
		return false, fmt.Errorf("No applications found in manifest")
	}

	for i, app := range manifest.Applications {
		buildPacks := app.Buildpacks

		if len(buildPacks) > 1 {
			return false, fmt.Errorf("More than one Cloud Foundry Buildpack is not supported. Please check your manifest file")
		}
		if len(buildPacks) == 1 {
			app.Buildpack = buildPacks[0]
			app.Buildpacks = []string{}
			fmt.Printf("Buildpacks: %v\n", app.Buildpacks)
			manifest.Applications[i] = app
			return true, nil
		}
		fmt.Printf("No build packs found\n")
	}

	return false, nil
}
