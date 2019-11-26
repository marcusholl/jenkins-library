import com.sap.piper.JenkinsUtils

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.BashUtils
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils

import groovy.transform.Field

import hudson.AbortException

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    'action',
    'apiUrl',
    'credentialsId',
    'deploymentId',
    'deployIdLogPattern',
    'deployOpts',
    /** A map containing properties forwarded to dockerExecute. For more details see [here][dockerExecute] */
    'docker',
    'loginOpts',
    'mode',
    'mtaPath',
    'org',
    'space',
    'xsSessionFile',
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS


/**
  * Performs an XS deployment
  *
  * In case of blue-green deployments the step is called for the deployment in the narrower sense
  * and later again for resuming or aborting. In this case both calls needs to be performed from the
  * same directory.
  */
@GenerateDocumentation
void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def utils = parameters.juStabUtils ?: new Utils()

        final script = checkScript(this, parameters) ?: this

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .addIfEmpty('mtaPath', script.commonPipelineEnvironment.getMtarFilePath())
            .addIfEmpty('deploymentId', script.commonPipelineEnvironment.xsDeploymentId)
            .mixin(parameters, PARAMETER_KEYS)

        Map config = configHelper.use()


        configHelper
            .collectValidationFailures()
            /** The credentialsId */
            .withMandatoryProperty('credentialsId')
            .use()

        utils.pushToSWA([
            step: STEP_NAME,
        ], config)

        echo "DOCKER-CONFIG: ${config.docker}"


        // for now we copy the piper bin into the workspace (in order to be able to use it from xs docker image)
        sh "cp \${JENKINS_HOME}/piper ."

        lock(getLockIdentifier(config)) {

            withCredentials([usernamePassword(
                    credentialsId: config.credentialsId,
                    passwordVariable: 'PASSWORD',
                    usernameVariable: 'USERNAME')]) {
                dockerExecute([script: script].plus(config.docker)) {
		    sh """#!/bin/bash
                        ./piper --verbose --customConfig .pipeline/config.yml xsDeploy --user \${USERNAME} --password \${PASSWORD}
                    """
                }
            }
        }
    }
}

String getLockIdentifier(Map config) {
    "$STEP_NAME:${config.apiUrl}:${config.org}:${config.space}"
}

