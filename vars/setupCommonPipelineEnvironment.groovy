import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import groovy.transform.Field

@Field String STEP_NAME = 'setupCommonPipelineEnvironment'
@Field Set GENERAL_CONFIG_KEYS = ['collectTelemetryData']

def call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters.script
        def cpe = script.commonPipelineEnvironment
        prepareDefaultValues script: script, customDefaults: parameters.customDefaults

        String configFile = parameters.get('configFile')

        loadConfigurationFromFile(cpe, configFile)

        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(cpe, GENERAL_CONFIG_KEYS)
            .use()

        new Utils().pushToSWA([step: STEP_NAME, stepParam4: parameters.customDefaults?'true':'false'], config)
    }
}

private boolean isYaml(String fileName) {
    return fileName.endsWith(".yml") || fileName.endsWith(".yaml")
}

private boolean isProperties(String fileName) {
    return fileName.endsWith(".properties")
}

private loadConfigurationFromFile(cpe, String configFile) {

    String defaultPropertiesConfigFile = '.pipeline/config.properties'
    String defaultYmlConfigFile = '.pipeline/config.yml'

    if (configFile?.trim()?.length() > 0 && isProperties(configFile)) {
        Map configMap = readProperties(file: configFile)
        cpe.setConfigProperties(configMap)
    } else if (fileExists(defaultPropertiesConfigFile)) {
        Map configMap = readProperties(file: defaultPropertiesConfigFile)
        cpe.setConfigProperties(configMap)
    }

    if (configFile?.trim()?.length() > 0 && isYaml(configFile)) {
        cpe.configuration = readYaml(file: configFile)
    } else if (fileExists(defaultYmlConfigFile)) {
        cpe.configuration = readYaml(file: defaultYmlConfigFile)
    }
}
