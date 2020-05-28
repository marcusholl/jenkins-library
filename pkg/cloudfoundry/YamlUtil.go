package cloudfoundry

import (
	"fmt"
	"reflect"
	"github.com/SAP/jenkins-library/pkg/log"
)

//Substitute ...
func Substitute(document map[string]interface{}, replacements map[string]interface{}) error {
	log.Entry().Infof("Inside SUBSTITUTE")
	var transformed interface{}
	transformed = make(map[string]interface{})

	if t, ok := transformed.(map[string]interface{}); ok {
		t["test"] = "debug"
	}

	log.Entry().Infof("TTTT: %v", transformed)
	err := traverse(document, &transformed , "")
	if err != nil {
		log.Entry().Warningf("Error: %v", err.Error())
	}

	log.Entry().Infof("transformed: %v", transformed)

	return err
}

func traverse(node interface{}, transformedNode interface{}, key string) error {

//	if tn, ok := transformedNode.(map[string]interface{}); ok { 
//		log.Entry().Info("DEBUG2 added")
//		tn["debug2"] = "tested"
//	}

	log.Entry().Infof("transformedNode: %v", transformedNode)

	log.Entry().Infof("Current node is: %v, type: %v", node, reflect.TypeOf(node))

	if s, ok := node.(string); ok {
		log.Entry().Infof("We have a string value: '%s'", s)
		return nil
	}

	if m, ok := node.(map[string]interface{}); ok {

		t := make(map[string]interface{})
		log.Entry().Info("Traversing map ...")
		for key, value := range m {
			
			log.Entry().Infof("traversing map entry '%v' ...", key)
			if err := traverse(value, t, ""); err != nil {
				log.Entry().Warningf("YERROR: %v", err.Error())
				return err
			}
			log.Entry().Infof("... map entry '%v' traversed", key)
		}
		log.Entry().Info("map fully traversed")

		if d, ok := transformedNode.(map[string]interface{}); ok {
			log.Entry().Infof("inserted to map: %v, key %s", t, key)
			d[key] = t
		}
		if d, ok := transformedNode.([]interface{}); ok {
			log.Entry().Infof("appended to slice %v", t)
			d = append(d, t)
		}


		return nil
	}

	if v, ok := node.([]interface{}); ok {
		log.Entry().Info("traversing slice ...")
		t := make([]interface{}, 0)
		for i, e := range v {
			log.Entry().Infof("traversing slice entry '%v' ...", i)	
			if err := traverse(e, t, ""); err != nil {
				log.Entry().Warningf("XERROR: %v", err.Error())
				return err
			}
			log.Entry().Infof("... slice entry '%v' traversed", i)
		}
		log.Entry().Infof("slice fully traversed.")

		if d, ok := transformedNode.(map[string]interface{}); ok {
			log.Entry().Infof("inserted to map: %v, key %s", t, key)
			d[key] = t
		}
		if d, ok := transformedNode.([]interface{}); ok {
			log.Entry().Infof("appended to slice %v", t)
			d = append(d, t)
		}
		return nil
	}

	log.Entry().Warningf("We received something which we can't handle: %v, %v", node, reflect.TypeOf(node))
	return fmt.Errorf("We received something which we can't handle: %v, %v", node, reflect.TypeOf(node))
}