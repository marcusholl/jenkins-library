./piper --verbose \
    --noTelemetry \
  cloudFoundryDeploy \
    --apiEndpoint https://example.org/cf \
    --deployTool cf_native \
    --deployType standard \
    --org org \
    --username itsme \
    --password secret \
    --appName hugo \
    --manifestVariables a=x \
    --manifestVariables b=y \
    --manifestVariablesFiles 123.txt \
    --manifestVariablesFiles abc.txt \
    --manifest manifest.yml \

#  --space space \
# --smokeTestScript blueGreenCheckScript.sh \
# --keepOldInstance true \
# --mtaExtensionDescriptor ext \
