import static org.junit.Assert.assertThat

import java.util.regex.Matcher

import org.hamcrest.Matchers

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.contains
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.equalTo
import static org.hamcrest.Matchers.equals
import static org.hamcrest.Matchers.hasSize
import static org.hamcrest.Matchers.is

import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.CommandLineMatcher
import util.JenkinsCredentialsRule
import util.JenkinsDockerExecuteRule
import util.JenkinsFileExistsRule
import util.JenkinsLockRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsReadJsonRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

import com.sap.piper.JenkinsUtils
import com.sap.piper.PiperGoUtils

import hudson.AbortException

class XsDeployTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()

    private List existingFiles =  [
        '.xsconfig',
        'myApp.mta'
    ]

    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsLockRule lockRule = new JenkinsLockRule(this)
    private JenkinsLoggingRule logRule = new JenkinsLoggingRule(this)
    private JenkinsDockerExecuteRule dockerRule = new JenkinsDockerExecuteRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
                                        .around(new JenkinsReadYamlRule(this))
                                        .around(new JenkinsReadJsonRule(this))
                                        .around(stepRule)
                                        .around(dockerRule)
                                        .around(writeFileRule)
                                        .around(new JenkinsCredentialsRule(this)
                                            .withCredentials('myCreds', 'cred_xs', 'topSecret')
                                            .withCredentials('XS2', 'user', 'pass'))
                                        .around(new JenkinsFileExistsRule(this, existingFiles))
                                        .around(lockRule)
                                        .around(shellRule)
                                        .around(logRule)
                                        .around(thrown)

    @Test
    public void testSanityChecks() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage(
            allOf(
                containsString('ERROR - NO VALUE AVAILABLE FOR:'),
                containsString('apiUrl'),
                containsString('org'),
                containsString('space'),
                containsString('mtaPath')))

        stepRule.step.xsDeploy(script: nullScript)
    }

    @Test
    public void testLoginFailed() {

        thrown.expect(AbortException)
        thrown.expectMessage('xs login failed')

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '#!/bin/bash xs login .*', 1)

        try {
            stepRule.step.xsDeploy(
                script: nullScript,
                apiUrl: 'https://example.org/xs',
                org: 'myOrg',
                space: 'mySpace',
                credentialsId: 'myCreds',
                mtaPath: 'myApp.mta'
            )
        } catch(AbortException e ) {

            assertThat(shellRule.shell,
                allOf(
                    // first item: the login attempt
                    // second item: we try to provide the logs
                    hasSize(2),
                    new CommandLineMatcher()
                        .hasProlog("#!/bin/bash")
                        .hasSnippet('xs login'),
                    new CommandLineMatcher()
                        .hasProlog('LOG_FOLDER')
                        .hasSnippet('cat \\$\\{LOG_FOLDER\\}/\\*')
                )
            )
            throw e
        }
    }

    @Test
    public void testDeployFailed() {

        thrown.expect(AbortException)
        thrown.expectMessage('Failed command(s): [xs deploy]. Check earlier log for details.')

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '#!/bin/bash.*xs deploy .*', {throw new AbortException()})

        try {
            stepRule.step.xsDeploy(
                script: nullScript,
                apiUrl: 'https://example.org/xs',
                org: 'myOrg',
                space: 'mySpace',
                credentialsId: 'myCreds',
                mtaPath: 'myApp.mta'
            )
        } catch(AbortException e ) {

            assertThat(shellRule.shell,
                allOf(
                    hasSize(4),
                    new CommandLineMatcher()
                        .hasProlog("#!/bin/bash")
                        .hasSnippet('xs login'),
                    new CommandLineMatcher()
                        .hasProlog("#!/bin/bash")
                        .hasSnippet('xs deploy'),
                    new CommandLineMatcher()
                        .hasProlog('#!/bin/bash')
                        .hasSnippet('xs logout'), // logout must be present in case deployment failed.
                    new CommandLineMatcher()
                        .hasProlog('')
                        .hasSnippet('rm \\$\\{XSCONFIG\\}') // remove the session file after logout
                )
            )
            throw e
        }
    }

    @Test
    public void testNothingHappensWhenModeIsNone() {

        stepRule.step.xsDeploy(
            script: nullScript,
            mode: 'NONE'
        )

        assertThat(logRule.log, containsString('Deployment skipped intentionally.'))
        assertThat(shellRule.shell, hasSize(0))
    }

    @Test
    public void testDeploymentFailsWhenDeployableIsNotPresent() {

        thrown.expect(AbortException)
        thrown.expectMessage('Deployable \'myApp.mta\' does not exist.')

        existingFiles.remove('myApp.mta')

        try {
            stepRule.step.xsDeploy(
                script: nullScript,
                apiUrl: 'https://example.org/xs',
                org: 'myOrg',
                space: 'mySpace',
                credentialsId: 'myCreds',
                mtaPath: 'myApp.mta'
            )
        } catch(AbortException e) {

            // no shell operation happened in this case.
            assertThat(shellRule.shell.size(), is(0))

            throw e
        }
    }

    @Test
    public void testDeployStraighForward() {

        stepRule.step.xsDeploy(
            script: nullScript,
            apiUrl: 'https://example.org/xs',
            org: 'myOrg',
            space: 'mySpace',
            credentialsId: 'myCreds',
            deployOpts: '-t 60',
            mtaPath: 'myApp.mta'
        )

        assertThat(shellRule.shell,
            allOf(
                new CommandLineMatcher()
                    .hasProlog("#!/bin/bash xs login")
                    .hasSnippet('xs login')
                    .hasOption('a', 'https://example.org/xs')
                    .hasOption('u', 'cred_xs')
                    .hasSingleQuotedOption('p', 'topSecret')
                    .hasOption('o', 'myOrg')
                    .hasOption('s', 'mySpace'),
                new CommandLineMatcher()
                    .hasProlog("#!/bin/bash")
                    .hasSnippet('xs deploy')
                    .hasOption('t', '60')
                    .hasArgument('\'myApp.mta\''),
                new CommandLineMatcher()
                    .hasProlog("#!/bin/bash")
                    .hasSnippet('xs logout')
            )
        )

        assertThat(lockRule.getLockResources(), contains('xsDeploy:https://example.org/xs:myOrg:mySpace'))

    }

    @Test
    public void testInvalidDeploymentModeProviced() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage('No enum constant')

        stepRule.step.xsDeploy(
            script: nullScript,
            apiUrl: 'https://example.org/xs',
            org: 'myOrg',
            space: 'mySpace',
            credentialsId: 'myCreds',
            deployOpts: '-t 60',
            mtaPath: 'myApp.mta',
            mode: 'DOES_NOT_EXIST'
        )
    }

    @Test
    public void testActionProvidedForStandardDeployment() {

        thrown.expect(AbortException)
        thrown.expectMessage(
            'Cannot perform action \'resume\' in mode \'deploy\'. Only action \'none\' is allowed.')

        stepRule.step.xsDeploy(
            script: nullScript,
            apiUrl: 'https://example.org/xs',
            org: 'myOrg',
            space: 'mySpace',
            credentialsId: 'myCreds',
            deployOpts: '-t 60',
            mtaPath: 'myApp.mta',
            mode: 'DEPLOY', // this is the default anyway
            action: 'RESUME'
        )
    }

    @Test
    public void testBlueGreenDeployFailes() {

        thrown.expect(AbortException)
        thrown.expectMessage('Failed command(s): [xs bg-deploy]')

        logRule.expect('Something went wrong')

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '#!/bin/bash.*xs bg-deploy .*',
            { throw new AbortException('Something went wrong.') })

        try {
            stepRule.step.xsDeploy(
                script: nullScript,
                apiUrl: 'https://example.org/xs',
                org: 'myOrg',
                space: 'mySpace',
                credentialsId: 'myCreds',
                mtaPath: 'myApp.mta',
                mode: 'BG_DEPLOY'
            )
        } catch(AbortException e) {

            // in case there is a deployment failure we have to logout also for bg-deployments
            assertThat(shellRule.shell,
                new CommandLineMatcher()
                    .hasProlog('#!/bin/bash')
                    .hasSnippet('xs logout')
            )

            throw e
        }
    }

    @Test
    public void testParametersViaSignature() {

        String paramsAsJson

        helper.registerAllowedMethod('libraryResource', [String], { configFile -> "{name: ${configFile}}"})
        helper.registerAllowedMethod('withEnv', [List, Closure], {l, c -> println "PARAMS: ${l}" ; paramsAsJson = l[0];  c()})
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*xsDeploy .*', '{"operationId": "1234"}')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*getConfig --contextConfig --stepMetadata.*', '{"dockerImage": "xs"}')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*getConfig --stepMetadata.*', '{"mode": "BG_DEPLOY", "action": "NONE", "apiUrl": "https://example.org/xs", "org": "myOrg", "space": "mySpace"}')

        PiperGoUtils goUtils = new PiperGoUtils(null) {
            void unstashPiperBin() {
            }
        }
        stepRule.step.xsDeploy(
            script: nullScript,
            apiUrl: 'https://example.org/xs',
            org: 'myOrg',
            space: 'mySpace',
            credentialsId: 'myCreds',
            deployOpts: '-t 60',
            mtaPath: 'myApp.mta',
            mode: 'BG_DEPLOY',
            piperGoUtils: goUtils
        )

        assertThat(paramsAsJson, equalTo('PIPER_parametersJSON={"apiUrl":"https://example.org/xs","org":"myOrg","space":"mySpace","credentialsId":"myCreds","deployOpts":"-t 60","mtaPath":"myApp.mta","mode":"BG_DEPLOY"}'))
    }

    @Test
    public void testBlueGreenDeployInitStraighForward() {

        boolean unstashCalled

        helper.registerAllowedMethod('libraryResource', [String], { configFile -> "{name: ${configFile}}"})
        helper.registerAllowedMethod('withEnv', [List, Closure], {l, c -> c()})
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*xsDeploy .*', '{"operationId": "1234"}')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*getConfig --contextConfig --stepMetadata.*', '{"dockerImage": "xs"}')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*getConfig --stepMetadata.*', '{"mode": "BG_DEPLOY", "action": "NONE", "apiUrl": "https://example.org/xs", "org": "myOrg", "space": "mySpace"}')

        PiperGoUtils goUtils = new PiperGoUtils(null) {
            void unstashPiperBin() {
                unstashCalled = true
            }
        }
        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils
        )

        assertThat(unstashCalled, equalTo(true))

        assertThat(nullScript.commonPipelineEnvironment.xsDeploymentId, is('1234'))

        assertThat(writeFileRule.files.keySet(), contains('metadata/xsDeploy.yaml'))
        
        assertThat(dockerRule.dockerParams.dockerImage, equalTo('xs'))
        assertThat(dockerRule.dockerParams.dockerPullImage, equalTo(false))
        
        assertThat(shellRule.shell,
            allOf(
                new CommandLineMatcher()
                    .hasProlog('./piper version'),
                new CommandLineMatcher()
                    .hasProlog('./piper getConfig --contextConfig --stepMetadata \'metadata/xsDeploy.yaml\''),
                new CommandLineMatcher()
                    .hasProlog('./piper getConfig --stepMetadata \'metadata/xsDeploy.yaml\''),
                new CommandLineMatcher()
                    .hasProlog('#!/bin/bash ./piper xsDeploy --user \\$\\{USERNAME\\} --password \\$\\{PASSWORD\\}') //  
            )
        )

        assertThat(lockRule.getLockResources(), contains('xsDeploy:https://example.org/xs:myOrg:mySpace'))
    }

    @Test
    public void testBlueGreenDeployResumeWithoutDeploymentId() {

        // this happens in case we would like to complete a deployment without having a (successful) deployments before.

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage(
            allOf(
                containsString('No operation id provided'),
                containsString('Was there a deployment before?')))

        helper.registerAllowedMethod('libraryResource', [String], { configFile -> "{name: ${configFile}}"})
        helper.registerAllowedMethod('withEnv', [List, Closure], {l, c -> c()})

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*getConfig --contextConfig --stepMetadata.*', '{"dockerImage": "xs"}')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*getConfig --stepMetadata.*', '{"mode": "BG_DEPLOY", "action": "NONE", "apiUrl": "https://example.org/xs", "org": "myOrg", "space": "mySpace"}')

        PiperGoUtils goUtils = new PiperGoUtils(null) {
            void unstashPiperBin() {
            }
        }

        nullScript.commonPipelineEnvironment.xsDeploymentId = null // is null anyway, just for clarification

        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils,
        )
    }

    @Test
    public void testBlueGreenDeployWithoutExistingSession() {

        thrown.expect(AbortException)
        thrown.expectMessage(
            'For the current configuration an already existing session is required.' +
            ' But there is no already existing session')

        existingFiles.remove('.xsconfig')

        stepRule.step.xsDeploy(
            script: nullScript,
            apiUrl: 'https://example.org/xs',
            org: 'myOrg',
            space: 'mySpace',
            credentialsId: 'myCreds',
            mode: 'BG_DEPLOY',
            action: 'RESUME'
        )

    }

    @Test
    public void testBlueGreenDeployResumeFails() {

        // e.g. we try to resume a deployment which did not succeed or which was already resumed or aborted.

        thrown.expect(AbortException)
        thrown.expectMessage('Failed command(s): [xs bg-deploy -a resume].')

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'xs bg-deploy -i .*', 1)

        nullScript.commonPipelineEnvironment.xsDeploymentId = '1234'

        try {
            stepRule.step.xsDeploy(
                script: nullScript,
                apiUrl: 'https://example.org/xs',
                org: 'myOrg',
                space: 'mySpace',
                credentialsId: 'myCreds',
                mode: 'BG_DEPLOY',
                action: 'RESUME'
            )
        } catch(AbortException e) {

            // logout must happen also in case of a failed deployment
            assertThat(shellRule.shell,
                new CommandLineMatcher()
                    .hasProlog('')
                    .hasSnippet('xs logout'))
            throw e
        }
    }

    @Test
    public void testBlueGreenDeployResume() {

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'xs bg-deploy -i .*', 0)

        nullScript.commonPipelineEnvironment.xsDeploymentId = '1234'

        stepRule.step.xsDeploy(
            script: nullScript,
            apiUrl: 'https://example.org/xs',
            org: 'myOrg',
            space: 'mySpace',
            credentialsId: 'myCreds',
            mode: 'BG_DEPLOY',
            action: 'RESUME'
        )

        // there is no login in case of a resume since we have to use the old session which triggered the deployment.
        assertThat(shellRule.shell,
            allOf(
                hasSize(3),
                new CommandLineMatcher()
                    .hasProlog('#!/bin/bash')
                    .hasSnippet('xs bg-deploy')
                    .hasOption('i', '1234')
                    .hasOption('a', 'resume'),
                new CommandLineMatcher()
                    .hasProlog("#!/bin/bash")
                    .hasSnippet('xs logout'),
                new CommandLineMatcher()
                    .hasProlog('')
                    .hasSnippet('rm \\$\\{XSCONFIG\\}') // delete the session file after logout
            )
        )

        assertThat(lockRule.getLockResources(), contains('xsDeploy:https://example.org/xs:myOrg:mySpace'))

    }

}