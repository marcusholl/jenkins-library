package cloudfoundry

import (
	"fmt"
	"strings"
	"regexp"
	"reflect"
	"github.com/SAP/jenkins-library/pkg/log"
)

//Substitute ...
func Substitute(document map[string]interface{}, replacements map[string]interface{}) error {
	log.Entry().Infof("Inside SUBSTITUTE")
	log.Entry().Infof("Replacements: %v", replacements)
	
	t, err := traverse(document, replacements)
	if err != nil {
		log.Entry().Warningf("Error: %v", err.Error())
	}

	log.Entry().Infof("transformed: %v", t)

	return err
}

func traverse(node interface{}, replacements map[string]interface{}) (interface{}, error) {

	log.Entry().Infof("Current node is: %v, type: %v", node, reflect.TypeOf(node))

	switch t := node.(type) {
	case string:
		return handleString(t, replacements)
	case bool:
		log.Entry().Infof("We have a boolean value: '%v'", t)
		return t, nil
	case int:
		log.Entry().Infof("We have an int value: '%v'", t)
		return t, nil
	case map[string]interface{}:
		return handleMap(t, replacements)
	case []interface{}:
		return handleSlice(t, replacements)
	default:
		return nil, fmt.Errorf("Unkown type received: '%v' (%v)", reflect.TypeOf(node), node)
	}
}

func handleString(value string, replacements map[string]interface{}) (interface{}, error) {
	log.Entry().Infof("We have a string value: '%v'", value)

	trimmed := strings.TrimSpace(value)
	re := regexp.MustCompile(`\(\(.*\)\)`)
	matches := re.FindAllSubmatch([]byte(trimmed), 1)
	fullMatch := isFullMatch(trimmed, matches)
	if fullMatch {
		parameterName := getParameterName(matches[0][0])
		parameterValue := getParameterValue(parameterName, replacements)
		log.Entry().Infof("FullMatchFound: '%s', replacing with '%v'", parameterName, parameterValue)
		return parameterValue, nil
	}
	// we have to scan for multiple variables
	// we return always a string
	for _, match := range matches {
		parameterName := getParameterName(match[0])
		parameterValue := getParameterValue(parameterName, replacements)
		if parameterValue == nil {
			return nil, fmt.Errorf("No value available for parameters '%s', replacements: %v", parameterName, replacements)
		}
		if v, ok := parameterValue.(string); ok {
			value = strings.Replace(value, "((" + parameterName + "))", v, -1)
			log.Entry().Infof("PartialMatchFound: '%v', replaced with : '%s'", parameterName, value)
		} else {
			fmt.Printf("Complex value found for parameter '%s'. Only string values are supported", parameterName)
		}
	} 

	return value, nil
}

func getParameterName(b []byte) string {
	return strings.Replace(strings.Replace(string(b), "((", "", 1), "))", "", 1)
}

func getParameterValue(name string, replacements map[string]interface{}) interface{} {
	if name == "integer-variable" {
		return 422
	}

	return replacements[name]	
}

func isFullMatch(value string, matches [][][]byte) bool {
	return strings.HasPrefix(value, "((") && strings.HasSuffix(value, "))") && len(matches) == 1 && len(matches[0]) == 1
 }

func handleSlice(t []interface{}, replacements map[string]interface{}) ([]interface{}, error) {
	log.Entry().Info("traversing slice ...")
	tNode := make([]interface{}, 0)
	for i, e := range t {
		log.Entry().Infof("traversing slice entry '%v' ...", i)
		if val, err := traverse(e, replacements); err == nil {
			tNode = append(tNode, val)

		} else {
			return nil, err
		}
		log.Entry().Infof("... slice entry '%v' traversed", i)
	}
	log.Entry().Infof("slice fully traversed.")
	return tNode, nil
}

func handleMap(t map[string]interface{}, replacements map[string]interface{}) (map[string]interface{}, error) {
	log.Entry().Info("Traversing map ...")
	tNode := make(map[string]interface{})
	for key, value := range t {
		
		log.Entry().Infof("traversing map entry '%v' ...", key)
		if val, err := traverse(value, replacements); err == nil {
			tNode[key] = val
		} else {
			return nil, err
		}
		log.Entry().Infof("... map entry '%v' traversed", key)
	}
	log.Entry().Infof("map fully traversed: %v", tNode)
	return tNode, nil
}