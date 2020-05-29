package cloudfoundry

import (
	"fmt"
	"reflect"
	"github.com/SAP/jenkins-library/pkg/log"
)

//Substitute ...
func Substitute(document map[string]interface{}, replacements map[string]interface{}) error {
	log.Entry().Infof("Inside SUBSTITUTE")
	
	t, err := traverse(document)
	if err != nil {
		log.Entry().Warningf("Error: %v", err.Error())
	}

	log.Entry().Infof("transformed: %v", t)

	return err
}

func traverse(node interface{}) (interface{}, error) {

	log.Entry().Infof("Current node is: %v, type: %v", node, reflect.TypeOf(node))

	switch t := node.(type) {
	case string:
		return handleString(t)
	case bool:
		log.Entry().Infof("We have a boolean value: '%v'", t)
		return t, nil
	case int:
		log.Entry().Infof("We have an int value: '%v'", t)
		return t, nil
	case map[string]interface{}:
		return handleMap(t)
	case []interface{}:
		return handleSlice(t)
	default:
		return nil, fmt.Errorf("Unkown type received: '%v' (%v)", reflect.TypeOf(node), node)
	}
}

func handleString(t string) (interface{}, error) {
	log.Entry().Infof("We have a string value: '%v'", t)
	return t, nil
}

func handleSlice(t []interface{}) ([]interface{}, error) {
	log.Entry().Info("traversing slice ...")
	tNode := make([]interface{}, 0)
	for i, e := range t {
		log.Entry().Infof("traversing slice entry '%v' ...", i)
		if val, err := traverse(e); err == nil {
			tNode = append(tNode, val)

		} else {
			return nil, err
		}
		log.Entry().Infof("... slice entry '%v' traversed", i)
	}
	log.Entry().Infof("slice fully traversed.")
	return tNode, nil
}

func handleMap(t map[string]interface{}) (map[string]interface{}, error) {
	log.Entry().Info("Traversing map ...")
	tNode := make(map[string]interface{})
	for key, value := range t {
		
		log.Entry().Infof("traversing map entry '%v' ...", key)
		if val, err := traverse(value); err == nil {
			tNode[key] = val
		} else {
			return nil, err
		}
		log.Entry().Infof("... map entry '%v' traversed", key)
	}
	log.Entry().Infof("map fully traversed: %v", tNode)
	return tNode, nil
}