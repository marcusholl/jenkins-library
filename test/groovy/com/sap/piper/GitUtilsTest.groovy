package com.sap.piper

import hudson.AbortException

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsShellCallRule
import util.Rules

import static org.junit.Assert.assertEquals
import static org.hamcrest.Matchers.equalTo
import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertFalse
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.notNullValue
import static org.junit.Assert.assertNotNull
import static org.junit.Assert.assertNull
import static org.junit.Assert.assertThat

class GitUtilsTest extends BasePiperTest {

    JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this).around(jscr).around(thrown)

    @Before
    void init() throws Exception {
        jscr.setReturnValue('git rev-parse HEAD', 'testCommitId')
    }

    @Test
    void testIsInsideWorkTree() {
        jscr.setReturnValue('git rev-parse --is-inside-work-tree 2>/dev/null', 'true\n')
        assertTrue(gitUtils.insideWorkTree())
    }

    @Test
    void testNotInsideWorkTree() {
        jscr.setReturnValue('git rev-parse --is-inside-work-tree 2>/dev/null',
            {throw new AbortException('script returned exit code 128')})
        assertFalse(gitUtils.insideWorkTree())
    }

    @Test
    void testInsideGitTree() {
        jscr.setReturnValue('git rev-parse --is-inside-work-tree 2>/dev/null', 'false')
        assertFalse(gitUtils.insideWorkTree())
    }

    @Test
    void testGetGitCommitId() {
        jscr.setReturnValue('git rev-parse --is-inside-work-tree 2>/dev/null', 'true\n')
        assertEquals('testCommitId', gitUtils.getGitCommitIdOrNull())
    }

    @Test
    void testGetGitCommitIdNotAGitRepo() {
        jscr.setReturnValue('git rev-parse --is-inside-work-tree 2>/dev/null', 'false\n')
        assertNull(gitUtils.getGitCommitIdOrNull())
    }

    @Test
    void testExtractLogLinesWithDefaults() {
        gitUtils.extractLogLines()
        assertTrue(jscr.shell
                         .stream()
                           .anyMatch( { it ->
                             it.contains('git log --pretty=format:%b origin/master..HEAD')}))
    }

    @Test
    void testExtractLogLinesWithCustomValues() {
        gitUtils.extractLogLines('myFilter', 'HEAD~5', 'HEAD~1', '%B')
        assertTrue( jscr.shell
                          .stream()
                            .anyMatch( { it ->
                               it.contains('git log --pretty=format:%B HEAD~5..HEAD~1')}))
    }

    @Test
    void testExtractLogLinesFilter() {
        jscr.setReturnValue('#!/bin/bash git log --pretty=format:%b origin/master..HEAD', 'abc\n123')
        String[] log = gitUtils.extractLogLines('12.*')
        assertThat(log, is(notNullValue()))
        assertThat(log.size(),is(equalTo(1)))
        assertThat(log[0], is(equalTo('123')))
    }

    @Test
    void testExtractLogLinesFilterNoMatch() {
        jscr.setReturnValue('#!/bin/bash git log --pretty=format:%b origin/master..HEAD', 'abc\n123')
        String[] log = gitUtils.extractLogLines('xyz')
        assertNotNull(log)
        assertThat(log.size(),is(equalTo(0)))
	}
}
