package com.sap.piper


public class GoConfigHelper {

    
    /*
     * The returned string can be used directly in the command line for retrieving the configuration via go
     */
    String prepareConfigurations(List configs, String configCacheFolder) {
    
        for(def customDefault : configs) {
            writeFile(file: "${ADDITIONAL_CONFIGS_FOLDER}/${customDefault}", text: libraryResource(customDefault))
        }
        joinAndQuote(configs.reverse(), configCacheFolder)
    }
    
    /*
     * prefix is supposed to be provided without trailing slash
     */
    String joinAndQuote(List l, String prefix = '') {
        _l = []
    
        if(prefix == null) {
            prefix = ''
        }
        if(prefix.endsWith('/') || prefix.endsWith('\\'))
            throw new IllegalArgumentException("Provide prefix (${prefix}) without trailing slash")
    
        for(def e : l) {
            def _e = ''
            if(prefix.length() > 0) {
                _e += prefix
                _e += '/'
            }
            _e += e
            _l << '"' + _e + '"'
        }
        _l.join(' ')
    }
}
