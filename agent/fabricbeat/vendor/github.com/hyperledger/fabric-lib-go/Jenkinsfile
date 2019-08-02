//
// Copyright IBM Corp All Rights Reserved
//
// SPDX-License-Identifier: Apache-2.0
//

@Library("fabric-ci-lib") _
timeout(40) {
  node ('hyp-x') { // trigger build on x86_64 node
    timestamps {
      try {
        def ARCH = sh(script: "uname -m", returnStdout: true).trim()
        if (ARCH == "x86_64") {
          ARCH = "amd64"
        } else {
          error "Unable to detect system architecture using 'uname -m'"
        }

        stage("Cleanup Environment") {
          wrap([$class: 'AnsiColorBuildWrapper', 'colorMapName': 'xterm']) {
            fabBuildLibrary.cleanupEnv() // Cleanup the leftover build artifacts
            fabBuildLibrary.envOutput() //  Output Jenkins environment details on the console
          }
        }

        stage("Checkout SCM") {
          // Delete working directory
          deleteDir()
          // Clone the repository
          fabBuildLibrary.cloneRefSpec('fabric-lib-go')
        }

        // load the properties file
        props = fabBuildLibrary.loadProperties()

        // Set GOPATH
        env.GOROOT = "/opt/go/go${props["GO_VER"]}.linux.${ARCH}"
        env.GOPATH = "${WORKSPACE}/gopath"
        env.PATH = "$GOROOT/bin:$GOPATH/bin:$PATH"

        // Run Checks
        stage("Checks") {
          wrap([$class: 'AnsiColorBuildWrapper', 'colorMapName': 'xterm']) {
            try {
              dir("${WORKSPACE}/${BASE_DIR}") {
                sh 'make checks'
              }
            }
            catch (err) {
              failure_stage = "checks"
              currentBuild.result = 'FAILURE'
              throw err
            }
          }
        }

        // Run Unit-Tests
        stage("Unit Tests") {
          wrap([$class: 'AnsiColorBuildWrapper', 'colorMapName': 'xterm']) {
            try {
              dir("${WORKSPACE}/${BASE_DIR}") {
                sh 'make unit-tests'
              }
            }
            catch (err) {
              failure_stage = "Unit Tests"
              currentBuild.result = 'FAILURE'
              throw err
            }
          }
        }
      }
      finally {
        if (env.JOB_TYPE == "merge") {
          if (currentBuild.result == 'FAILURE') {
            fabBuildLibrary.sendRocketChatNotification() // Send the merge build failure notifications to "jenkins-robot" rocketChat channel
            fabBuildLibrary.sendEmailNotification() // Send the merge build failure email notifications to the person who initiated the build
          }
        }
        cleanWs()
      } // finally block
    } // timestamps block
  } // node block
} // timeout block