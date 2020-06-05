./piper --verbose \
    --noTelemetry \
  cloudFoundryDeploy \
    --apiEndpoint https://example.org/cf \
    --deployTool cf_native \
    --org org \
    --space space \
    --username itsme \
    --password secret \
    --mtaExtensionDescriptor philipsburg \
    --apiParameters 'my Api xParameters' \
    --deployType standard \
    --smokeTestScript blueGreenCheckScript.sh \
    --appName hugo \
    --keepOldInstance true \
    --manifestVariables a=x \
    --manifestVariables b=y \
    --manifestVariablesFiles 123.txt \
    --manifestVariablesFiles abc.txt \
    --manifest manifest.yml \
