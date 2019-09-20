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

        def piperGoUrl = parameters?.piperGoUrl

        if( ! piperGoUrl) throw new hudson.AbortException('No piper go version provided (parameter piperGoUrl)')

        def cmd = parameters?.cmd ?: 'echo "no parameters provided".'

        echo "Executing '${cmd}'"

        sh script: """#!/bin/bash
                curl --fail --insecure -o piper ${piperGoUrl} && chmod +x piper
        """, returnStdout: true

        def output = sh script: """#!/bin/bash
                ./piper dummy ${cmd}
        """, returnStdout: true

        echo "OUTPUT: ${output}"

    }
}

