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
void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters)

        sh """#!/bin/bash
              whoami
              pwd
              ls -la
              /opt/sap/piper/bin/piper mtaBuild
        """


    }
}

