package cloudfoundry

import (
	"fmt"
	"reflect"
	"github.com/SAP/jenkins-library/pkg/log"
)

//Substitute ...
func Substitute(document map[string]interface{}, replacements map[string]interface{}) error {
	log.Entry().Infof("Inside SUBSTITUTE")
	err := traverse(document)
	if err != nil {
		log.Entry().Warningf("Error: %v", err.Error())
	}
	return err
}

func traverse(node interface{}) error {

	log.Entry().Infof("Current node is: %v, type: %v", node, reflect.TypeOf(node))

	if s, ok := node.(string); ok {
		log.Entry().Infof("We have a string value: '%s'", s)
		return nil
	}

	if m, ok := node.(map[string]interface{}); ok {

		log.Entry().Info("Traversing map ...")
		for key, value := range m {
			log.Entry().Infof("traversing map entry '%v' ...", key)
			if err := traverse(value); err != nil {
				log.Entry().Warningf("YERROR: %v", err.Error())
				return err
			}
			log.Entry().Infof("... map entry '%v' traversed", key)
		}
		log.Entry().Info("map fully traversed")
		return nil
	}

	if v, ok := node.([]interface{}); ok {
		for i, e := range v {
			log.Entry().Infof("traversing slice entry '%v' ...", i)	
			if err := traverse(e); err != nil {
				log.Entry().Warningf("XERROR: %v", err.Error())
				return err
			}
			log.Entry().Infof("... slice entry '%v' traversed", i)
		}
		log.Entry().Infof("slice fully traversed.")
		return nil
	}

	log.Entry().Warningf("We received something which we can't handle: %v, %v", node, reflect.TypeOf(node))
	return fmt.Errorf("We received something which we can't handle: %v, %v", node, reflect.TypeOf(node))
}