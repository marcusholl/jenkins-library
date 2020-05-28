package cloudfoundry

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
)

//Substitute ...
func Substitute(document map[string]interface{}, replacements map[string]interface{}) error {
	log.Entry().Infof("Inside SUBSTITUTE")
	return traverse(document)
}

func traverse(node interface{}) error {

	log.Entry().Infof("Current node is: %v", node)

	if s, ok := node.(string); ok {
		log.Entry().Infof("We have a string value: '%s'", s)
		return nil
	}

	if m, ok := node.(map[string]interface{}); ok {

		for key, value := range m {
			log.Entry().Infof("traversing '%v' ...", key)	
			if err := traverse(value); err != nil {
				return err
			}
		}
	}

	if v, ok := node.([]interface{}); ok {
		for _, e := range v {
			if err := traverse(e); err != nil {
				return err
			}
		}
	}

	return fmt.Errorf("We received something which we can't handle: %v", node)
}