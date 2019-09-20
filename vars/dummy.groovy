import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import com.sap.piper.versioning.ArtifactVersioning

import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = getClass().getName()
@Field Map CONFIG_KEY_COMPATIBILITY = [gitSshKeyCredentialsId: 'gitCredentialsId']

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * A dummy comment for getting some experience with the interaction with go
 */
void call(Map parameters = [:], Closure body = null) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters)

        echo "Inside dummy step"

        def cmd = parameters?.cmd ?: 'echo "no parameters provided".'

        echo "Executing '${cmd}'"

        def output = sh script: """#!/bin/bash
                curl --fail --insecure -o piper https://nexussnap.wdf.sap.corp:8443/nexus/content/repositories/deploy.snapshots/com/sap/de/marcusholl/go/mygo/0.0.1-SNAPSHOT/mygo-0.0.1-20190920.115637-7-amd64.jar
                chmod +x piper
                ./piper dummy ${cmd}
        """, returnStdout: true

        echo "OUTPUT: ${output}"

    }
}

