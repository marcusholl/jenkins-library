package cloudfoundry

import (
	"github.com/SAP/jenkins-library/pkg/log"
)

//Substitute ...
func Substitute(document map[string]interface{}, replacements map[string]interface{}) error {
	log.Entry().Infof("Inside SUBSTITUTE")
	traverse(document)
	return nil
}

func traverse(document map[string]interface{}) error {

	log.Entry().Infof("The document is: %v", document)
	for key, value := range document {
		log.Entry().Infof("traversing '%v': '%v' ...", key, value)

		if v, ok := value.(string); ok {
			log.Entry().Infof("We have a string value: %s", v)
			continue
		}
		if v, ok := value.([]interface{}); ok {
			log.Entry().Infof("We have an interface slice: %v", v)
			continue
		}

		if v, ok := value.(map[interface{}]interface{}); ok {
			log.Entry().Infof("We have a map: %v", v)
			continue
		}

		log.Entry().Infof("We have something else: %v:%v", key, value)
	}

	return nil
}