package config

import (
	"io"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

// PipelineDefaults defines the structure of the pipeline defaults
type PipelineDefaults struct {
	Defaults []Config `json:"defaults"`
}

// ReadPipelineDefaults loads defaults and returns its content
func (d *PipelineDefaults) ReadPipelineDefaults(defaultSources []io.ReadCloser) error {

	for _, def := range defaultSources {

		def.Close()

		var c Config
		var err error

		content, err := ioutil.ReadAll(def)
		if err != nil {
			return errors.Wrapf(err, "error reading %v", def)
		}

		err = yaml.Unmarshal(content, &c)
		if err != nil {
			return errors.Wrapf(err, "error unmarshalling: %v", string(content))
		}

		d.Defaults = append(d.Defaults, c)
	}
	return nil
}
