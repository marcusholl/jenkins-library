package com.sap.piper

import hudson.AbortException

boolean insideWorkTree() {

    boolean insideWorkTree = false
    try {
        // git command below returns:
        // - 'true'  on stdout and exit code 0 in case we are in a work-tree
        // - 'false' on stdout and exit code 0 in case we are in git-dir ('.git')
        // - nothing written to stdout and exit code != 0 in case we are not at all in a worktree.
        insideWorkTree = Boolean.valueOf(
                             sh(returnStdout: true, script: 'git rev-parse --is-inside-work-tree 2>/dev/null')
                                 .trim())
    } catch(AbortException e) {
      // script returned with exit code != 0, this is the normal
      // behavior when located outside a work-tree.
      //
      // <paranoia>Of course there are also other possible reasons for ending up here, e.g.
      // git is not in path. Anyway: when the command returns with != 0 we are on the save side
      // with returning 'false'.</paranoia>
    }
    return insideWorkTree
}

String getGitCommitIdOrNull() {
    if ( insideWorkTree() ) {
        return getGitCommitId()
    } else {
        return null
    }
}

String getGitCommitId() {
    return sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
}

String[] extractLogLines(String filter = '',
                         String from = 'origin/master',
                         String to = 'HEAD',
                         String format = '%b') {

   // Checks below: there was an value provided from outside, but the value was null.
   // Throwing an exception is more transparent than making a fallback to the defaults
   // used in case the paramter is omitted in the signature.
   if(filter == null) throw new IllegalArgumentException('Parameter \'filter\' not provided.')
   if(! from?.trim()) throw new IllegalArgumentException('Parameter \'from\' not provided.')
   if(! to?.trim()) throw new IllegalArgumentException('Parameter \'to\' not provided.')
   if(! format?.trim()) throw new IllegalArgumentException('Parameter \'format\' not provided.')

    sh ( returnStdout: true,
         script: """#!/bin/bash
                    git log --pretty=format:${format} ${from}..${to}
                 """
       )?.split('\n')
        ?.findAll { line -> line ==~ /${filter}/ }

}
