package cloudfoundry

import (
	"fmt"
	"reflect"
	"github.com/SAP/jenkins-library/pkg/log"
)

//Substitute ...
func Substitute(document map[string]interface{}, replacements map[string]interface{}) error {
	log.Entry().Infof("Inside SUBSTITUTE")
	//var transformed interface{}
	transformed := make(map[string]interface{})

	
	transformed["test22"] = "debug"
	

	log.Entry().Infof("TTTT: %v", transformed)
	err := traverse(document, transformed , "", false)
	if err != nil {
		log.Entry().Warningf("Error: %v", err.Error())
	}

	log.Entry().Infof("transformed: %v", transformed)

	return err
}

func traverse(node interface{}, transformedNode interface{}, key string, nestedCall bool) error {

	log.Entry().Infof("Current node is: %v, key: %s, type: %v", node, key, reflect.TypeOf(node))

	var replaced interface{}

	if s, ok := node.(string); ok {
		log.Entry().Infof("We have a string value: '%s'", s)
		replaced = s
		return nil
	}

	if m, ok := node.(map[string]interface{}); ok {

		t := make(map[string]interface{})

		if ! nestedCall {
			if outermodeNode, ok := transformedNode.(map[string]interface{}); ok {
				t = outermodeNode
			}
		}

		log.Entry().Info("Traversing map ...")
		for key, value := range m {
			
			log.Entry().Infof("traversing map entry '%v' ...", key)
			if err := traverse(value, t, key, true); err != nil {
				return err
			}
			log.Entry().Infof("... map entry '%v' traversed", key)
		}
		log.Entry().Info("map fully traversed")
		replaced = t
	}

	if v, ok := node.([]interface{}); ok {
		log.Entry().Info("traversing slice ...")
		t := make([]interface{}, 0)
		for i, e := range v {
			log.Entry().Infof("traversing slice entry '%v' ...", i)	
			if err := traverse(e, t, "", true); err != nil {
				log.Entry().Warningf("XERROR: %v", err.Error())
				return err
			}
			log.Entry().Infof("... slice entry '%v' traversed", i)
		}
		log.Entry().Infof("slice fully traversed.")
		replaced = t
	}

	if !nestedCall {
		if d, ok := transformedNode.(map[string]interface{}); ok {

			if len(key) == 0 {
				return fmt.Errorf("Empty key detected. Cannot insert %v into map %v", replaced, d)
			}
			log.Entry().Infof("inserted to map: %v, key %s", replaced, key)
			d[key] = replaced
		}

		if d, ok := transformedNode.([]interface{}); ok {
			log.Entry().Infof("appended to slice %v", replaced)
			d = append(d, replaced)
		}
	}

	return nil
	//log.Entry().Warningf("We received something which we can't handle: %v, %v", node, reflect.TypeOf(node))
	//return fmt.Errorf("We received something which we can't handle: %v, %v", node, reflect.TypeOf(node))
}