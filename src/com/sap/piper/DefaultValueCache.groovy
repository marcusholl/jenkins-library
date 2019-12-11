package com.sap.piper

import com.sap.piper.MapUtils

@API
class DefaultValueCache implements Serializable {
    private static DefaultValueCache instance

    private Map defaultValues
    private List additionalConfigs = []

    private DefaultValueCache(Map defaultValues, List additionalConfigs = []){
        this.defaultValues = defaultValues
        if(additionalConfigs) {
            this.additionalConfigs << additionalConfigs
        }
    }

    List getAdditionalConfigurations() {
        [] << this.additionalConfigs
    }

    static getInstance(){
        return instance
   }

    static createInstance(Map defaultValues, List additionalConfigs = []) {
        instance = new DefaultValueCache(defaultValues, additionalConfigs)
    }

    Map getDefaultValues(){
        return defaultValues
    }

    static reset(){
        instance = null
    }

    static void prepare(Script steps, Map parameters = [:]) {
        if(parameters == null) parameters = [:]
        if(!DefaultValueCache.getInstance() || parameters.customDefaults) {
            def defaultValues = [:]
            def configFileList = ['default_pipeline_environment.yml']
            def customDefaults = parameters.customDefaults

            if(customDefaults in String)
                customDefaults = [customDefaults]
            if(customDefaults in List)
                configFileList += customDefaults
            for (def configFileName : configFileList){
                if(configFileList.size() > 1) steps.echo "Loading configuration file '${configFileName}'"
                def configuration = steps.readYaml text: steps.libraryResource(configFileName)
                defaultValues = MapUtils.merge(
                        MapUtils.pruneNulls(defaultValues),
                        MapUtils.pruneNulls(configuration))
            }
            DefaultValueCache.createInstance(defaultValues, customDefaults)
        }
    }
}
