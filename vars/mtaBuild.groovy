import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.MtaUtils
import com.sap.piper.Utils
import com.sap.piper.PiperGoUtils
import groovy.transform.Field

import static com.sap.piper.Utils.downloadSettingsFromUrl

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
    /** The name of the application which is being built. If the parameter has been provided and no `mta.yaml` exists, the `mta.yaml` will be automatically generated using this parameter and the information (`name` and `version`) from `package.json` before the actual build starts.*/
    'applicationName',
    /**
     * mtaBuildTool classic only: The target platform to which the mtar can be deployed.
     * @possibleValues 'CF', 'NEO', 'XSA'
     */
    'buildTarget',
    /**
     * Tool to use when building the MTA
     * @possibleValues 'classic', 'cloudMbt'
     */
    'mtaBuildTool',
    /** @see dockerExecute */
    'dockerImage',
    /** @see dockerExecute */
    'dockerEnvVars',
    /** @see dockerExecute */
    'dockerOptions',
    /** @see dockerExecute */
    'dockerWorkspace',
    /** The path to the extension descriptor file.*/
    'extension',
    /**
     * The location of the SAP Multitarget Application Archive Builder jar file, including file name and extension.
     * If you run on Docker, this must match the location of the jar file in the container as well.
     */
    'mtaJarLocation',
    /** Path or url to the mvn settings file that should be used as global settings file.*/
    'globalSettingsFile',
    /** The name of the generated mtar file including its extension. */
    'mtarName',
    /**
     * mtaBuildTool cloudMbt only: The target platform to which the mtar can be deployed.
     * @possibleValues 'CF', 'NEO', 'XSA'
     */
    'platform',
    /** Path or url to the mvn settings file that should be used as project settings file.*/
    'projectSettingsFile'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /** Url to the npm registry that should be used for installing npm dependencies.*/
    'defaultNpmRegistry'
])

/**
 * Executes the SAP Multitarget Application Archive Builder to create an mtar archive of the application.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        final script = checkScript(this, parameters) ?: this

        def utils = parameters.juStabUtils ?: new Utils()
        def piperGoUtils = parameters.piperGoUtils ?: new PiperGoUtils(utils)
        piperGoUtils.unstashPiperBin()

        // load default & individual configuration
        Map configuration = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .dependingOn('mtaBuildTool').mixin('dockerImage')
            .use()

        dockerExecute(
            script: script,
            dockerImage: configuration.dockerImage,
            dockerEnvVars: configuration.dockerEnvVars,
            dockerOptions: configuration.dockerOptions,
            dockerWorkspace: configuration.dockerWorkspace
        ) {

            sh """#!/bin/bash
                ./piper mtaBuild --mtaJarLocation=/opt/sap/mta/lib/mta.jar  --mtaBuildTool classic --buildTarget CF"""

            script.commonPipelineEnvironment.readFromDisk(script)

            echo "mtar file: ${script.commonPipelineEnvironment.mtarFilePath}"

        }
    }
}

def String getMtaId(String fileName){
    def mtaYaml = readYaml file: fileName
    if (!mtaYaml.ID) {
        error "Property 'ID' not found in ${fileName} file."
    }
    return mtaYaml.ID
}
