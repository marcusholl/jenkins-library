import com.sap.piper.JenkinsUtils
import com.sap.piper.PiperGoUtils


import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils

import groovy.transform.Field

@Field String METADATA_FILE = 'metadata/xsDeploy.yaml'
@Field String STEP_NAME = getClass().getName()

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

        //
        // The parameters map in provided from outside. That map might be used elsewhere in the pipeline
        // hence we should not modify it here. So we create a new map based on the parameters map.
	parameters = [:] << parameters

        // hard to predict how these two parameters looks like in its serialized form. Anyhow it is better
        // not to have these parameters forwarded somehow to the go layer.
        parameters.remove('juStabUtils')
        parameters.remove('script')

        //
        // For now - since the xsDeploy step is not merged and covered by a release - we stash
        // a locally built version of the piper-go binary in the pipeline script (Jenkinsfile) with
        // stash name "piper-bin". That stash is used inside method "unstashPiperBin".
        new PiperGoUtils(this, utils).unstashPiperBin()

        //
        // Printing the piper-go version. Should not be done here, but somewhere during materializing
        // the piper binary.
	def piperGoVersion = sh(returnStdout: true, script: "./piper version")
	echo "PiperGoVersion: ${piperGoVersion}"

        //
        // since there is no valid config provided (... null) telemetry is disabled.
        utils.pushToSWA([
            step: STEP_NAME,
        ], null)

        writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {

            sh "echo \"Parameters: \${PIPER_parametersJSON}\""

            //
            // context config gives us e.g. the docker image name. --> How does this work for customer maintained images?
            // There is a name provided in the metadata file. But we do not provide a docker image for that.
            // The user has to build that for her/his own. How do we expect to configure this?
            Map contextConfig = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}'"))

            //
            // The project config is used since we have project related parameters in the groovy layer. This is not handed over to the "payload"
            // go step at the moment. That step does the evaluation again on its own. Maybe it would be better to handover the project config via
            // --parametersJSON. But not sure how this works since we have also the corresponding environment variable (I think the command line parameters
            // has precedence.
            // The idea in this case would be to fully calculate the config ouside and just give it to the second call of the piper lib
            // --> we should discuss in order to find a suitable pattern here.
            Map projectConfig = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --stepMetadata '${METADATA_FILE}'"))

            echo "Context-Config: ${contextConfig}"
            echo "Project-Config: ${projectConfig}"

            // That config map here is only used in the groovy layer. Nothing is handed over to go.
            Map config = contextConfig <<
                [
                    apiUrl: projectConfig.apiUrl, // required on groovy level for acquire the lock
                    org: projectConfig.org,       // required on groovy level for acquire the lock
                    space: projectConfig.space,   // required on groovy level for acquire the lock
                    credentialsId: 'XS2',         // I saw the 'secrets' part in the metadata. But is does not work for me. --> hard coded for now.
                    docker: [
                        dockerImage: contextConfig.dockerImage,
                        dockerPullImage: false    // dockerPullImage apparently not provided by context config.
                    ]
                ]

            lock(getLockIdentifier(config)) {

                withCredentials([usernamePassword(
                        credentialsId: config.credentialsId,
                        passwordVariable: 'PASSWORD',
                        usernameVariable: 'USERNAME')]) {

                    dockerExecute([script: this].plus(config.docker)) {
                        sh """#!/bin/bash
                        ./piper --verbose xsDeploy --user \${USERNAME} --password \${PASSWORD}
                        """
                    }
                }
            }
        }
    }
}

String getLockIdentifier(Map config) {
    "$STEP_NAME:${config.apiUrl}:${config.org}:${config.space}"
}

