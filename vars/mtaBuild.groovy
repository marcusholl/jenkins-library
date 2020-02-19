import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GoConfigHelper
import com.sap.piper.Utils
import com.sap.piper.PiperGoUtils
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()
@Field String PIPER_DEFAULTS = 'default_pipeline_environment.yml'
@Field String METADATA_FOLDER = '.pipeline' // metadata file contains already the "metadata" folder level, hence we end up in a folder ".pipeline/metadata"
@Field String ADDITIONAL_CONFIGS_FOLDER='.pipeline/additionalConfigs'

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
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
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

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

        def contextConfig = sh(returnStdout: true, script: "./piper getConfig --stepMetadata '${METADATA_FOLDER}/${METADATA_FILE}' --defaultConfig ${configFiles} --contextConfig")

        echo "CONTEXT_CONFIG: ${contextConfig}"

//Map contextConfig = readJSON (text: sh(returnStdout: true, script: contextConfigScript))


        String configFiles = GoConfigHelper.prepareConfigurations([PIPER_DEFAULTS].plus(script.commonPipelineEnvironment.getCustomDefaults()), ADDITIONAL_CONFIGS_FOLDER)

        writeFile(file: "${METADATA_FOLDER}/${METADATA_FILE}", text: libraryResource(METADATA_FILE))

        def contextConfig = "./piper getConfig --stepMetadata '${METADATA_FOLDER}/${METADATA_FILE}' --defaultConfig ${configFiles} --contextConfig"

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {

            dockerExecute(
                script: script,
                dockerImage: configuration.dockerImage,
                dockerEnvVars: configuration.dockerEnvVars,
                dockerOptions: configuration.dockerOptions,
                dockerWorkspace: configuration.dockerWorkspace
            ) {

                sh """#!/bin/bash
                    ./piper mtaBuild"""

                script.commonPipelineEnvironment.readFromDisk(script)

            }
            echo "mtar file created by the build: '${script.commonPipelineEnvironment.mtarFilePath}'"
        }
    }
}

