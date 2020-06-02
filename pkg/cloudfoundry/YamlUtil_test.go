package cloudfoundry


import (
	"github.com/ghodss/yaml"
	"testing"
	"github.com/stretchr/testify/assert"
	"strings"

)
func TestXX(t *testing.T) {

	document := make(map[string]interface{})
	replacements := make(map[string]interface{})
 
yaml.Unmarshal([]byte(
`unique-prefix: uniquePrefix # A unique prefix. E.g. your D/I/C-User
xsuaa-instance-name: uniquePrefix-catalog-service-odatav2-xsuaa
hana-instance-name: uniquePrefix-catalog-service-odatav2-hana
integer-variable: 1
boolean-variable: Yes
float-variable: 0.25
json-variable: >
  [
    {"name":"token-destination",
     "url":"https://www.google.com",
     "forwardAuthToken": true}
  ]
object-variable:
  hello: "world"
  this:  "is an object with"
  one: 1
  float: 25.0
  bool: Yes`), &replacements)

	err := yaml.Unmarshal([]byte(
`applications:
- name: ((unique-prefix))-catalog-service-odatav2-0.0.1
  memory: 1024M
  disk_quota: 512M
  instances: ((integer-variable))
  buildpacks:
    - java_buildpack
  path: ./srv/target/srv-backend-0.0.1-SNAPSHOT.jar
  routes:
  - route: ((unique-prefix))-catalog-service-odatav2-001.cfapps.eu10.hana.ondemand.com

  services:
  - ((xsuaa-instance-name)) # requires an instance of xsuaa instantiated with xs-security.json of this project. See services-manifest.yml.
  - ((hana-instance-name))  # requires an instance of hana service with plan hdi-shared. See services-manifest.yml.

  env:
    spring.profiles.active: cloud # activate the spring profile named 'cloud'.
    xsuaa-instance-name: ((xsuaa-instance-name))
    db_service_instance_name: ((hana-instance-name))
    booleanVariable: ((boolean-variable))
    floatVariable: ((float-variable))
    json-variable: ((json-variable))
    object-variable: ((object-variable))
    string-variable: ((boolean-variable))-((float-variable))-((integer-variable))-((json-variable))
    single-var-with-string-constants: ((boolean-variable))-with-some-more-text
  `), &document)

	replaced, err := Substitute(document, replacements)

	assert.NoError(t, err)

		//
		// assertDataTypeAndSubstitutionCorrectness start

	if m, ok := replaced.(map[string]interface{}); ok {
	
		if apps, ok := m["applications"].([]interface{}); ok {
			app := apps[0]
			if appAsMap, ok := app.(map[string]interface{}); ok {

				instances := appAsMap["instances"]


				if one, ok := instances.(float64); ok {
					assert.Equal(t, 1, int(one))	
				}


				if services, ok := appAsMap["services"]; ok {
					if servicesAsSlice, ok := services.([]interface{}); ok {
						if _, ok := servicesAsSlice[0].(string); ok {
							assert.True(t, true)
						} else {
							assert.True(t, false)
						}
					} else {
						assert.True(t, false)
					}
				} else {
					assert.True(t, false)
				}


				if env, ok := appAsMap["env"]; ok {

					if envAsMap, ok := env.(map[string]interface{}); ok {

						if _, ok := envAsMap["floatVariable"].(float64); ok {

							assert.True(t, true)
						} else {
							assert.True(t, false)
						}

						if asBoolean, ok := envAsMap["booleanVariable"].(bool); ok {
							assert.True(t, asBoolean)
						} else {
							assert.True(t, false)
						}

						if _, ok := envAsMap["json-variable"].(string); ok {
							assert.True(t, true)
						} else {
							assert.True(t, false)
						}

						if _, ok := envAsMap["object-variable"].(map[string]interface{}); ok {
							assert.True(t, true)
						} else {
							assert.True(t, false)
						}

						if s, ok := envAsMap["string-variable"].(string); ok {
							assert.True(t, strings.HasPrefix(s, "true-0.25-1-"))
						} else {
							assert.True(t, false)
						}

						if s, ok := envAsMap["single-var-with-string-constants"].(string); ok {
							assert.Equal(t, s, "true-with-some-more-text")
						} else {
							assert.True(t, false)
						}


					} else {
						assert.True(t, false)
					}
				}
			}	
		}

		//assertDataTypeAndSubstitutionCorrectness END
		//


		//
		// assertCorrectVariableResolution START

		if m, ok := replaced.(map[string]interface{}); ok {
			if apps, ok := m["applications"].([]interface{}); ok {
				app := apps[0]
				if appAsMap, ok := app.(map[string]interface{}); ok {

					assert.Equal(t, "uniquePrefix-catalog-service-odatav2-0.0.1", appAsMap["name"])

					if env, ok := appAsMap["env"]; ok {

						if envAsMap, ok := env.(map[string]interface{}); ok {
							assert.Equal(t, "uniquePrefix-catalog-service-odatav2-xsuaa", envAsMap["xsuaa-instance-name"])
							assert.Equal(t, "uniquePrefix-catalog-service-odatav2-hana", envAsMap["db_service_instance_name"])
						} else {
							assert.True(t, false)
						}
						if servicesAsSlice, ok := appAsMap["services"].([]interface{}); ok {
							assert.Equal(t, "uniquePrefix-catalog-service-odatav2-xsuaa", servicesAsSlice[0])
							assert.Equal(t, "uniquePrefix-catalog-service-odatav2-hana", servicesAsSlice[1])
						} else {
							assert.True(t, false)
						}
					}
				}
			}
		}

		//
		// assertCorrectVariableResolution END


	}

	data, err := yaml.Marshal(&replaced)
	


	t.Logf("Data: %v", string(data))

	assert.NoError(t, err)

	assert.True(t, true, "Everything is fine")
}