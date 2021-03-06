def createDeployMessage(env) {
    env.DEPLOY_MESSAGE = "GIT_COMMIT : ${env.GIT_COMMIT}\n"
    if (env.CHANGE_ID)
        env.DEPLOY_MESSAGE +=  "CHANGE_ID : ${env.CHANGE_ID}\n"
    if (env.CHANGE_URL)
        env.DEPLOY_MESSAGE += "CHANGE_URL : ${env.CHANGE_URL}\n"
    if (env.CHANGE_TITLE)
        env.DEPLOY_MESSAGE += "CHANGE_TITLE : ${env.CHANGE_TITLE}\n"
    if (env.CHANGE_AUTHOR)
        env.DEPLOY_MESSAGE += "CHANGE_AUTHOR : ${env.CHANGE_AUTHOR}\n"
    if (env.CHANGE_AUTHOR_DISPLAY_NAME)
        env.DEPLOY_MESSAGE += "CHANGE_AUTHOR_DISPLAY_NAME : ${env.CHANGE_AUTHOR_DISPLAY_NAME}\n"
    if (env.CHANGE_AUTHOR_EMAIL)
        env.DEPLOY_MESSAGE += "CHANGE_AUTHOR_EMAIL : ${env.CHANGE_AUTHOR_EMAIL}\n"
    if (env.CHANGE_TARGET)
        env.DEPLOY_MESSAGE += "CHANGE_TARGET : ${env.CHANGE_TARGET}\n"
    if (env.GIT_BRANCH)
        env.DEPLOY_MESSAGE += "GIT_BRANCH : ${env.GIT_BRANCH}\n"
    else if (env.BRANCH_NAME)
        env.DEPLOY_MESSAGE += "BRANCH_NAME : ${env.BRANCH_NAME}\n"
    return env.DEPLOY_MESSAGE
}


pipeline {
    triggers { 
        pollSCM('H */4 * * 1-5') 
    }
    options {
        timestamps()
    //    skipDefaultCheckout(true)
    }

    agent { 
        dockerfile {
            args '--privileged --net=host'
            additionalBuildArgs '--network=host'
        }
    }

    stages {       
        stage('Checkout and prepare environment for testing') {
            steps {
                dir('src/eaas') {
                    checkout scm
                }
                sh("mongod --dbpath /tmp &")
                sh("/usr/local/bin/etcd -name etcd0  -advertise-client-urls https://127.0.0.1:2379,https://127.0.0.1:4001  -listen-client-urls https://0.0.0.0:2379,https://0.0.0.0:4001  -initial-advertise-peer-urls https://127.0.0.1:2380  -listen-peer-urls https://0.0.0.0:2380  -initial-cluster-token etcd-cluster-1  -initial-cluster etcd0=https://127.0.0.1:2380  -initial-cluster-state new --enable-v2=false --auto-tls --peer-auto-tls &")
            }
        }
        stage('Run tests') {
            steps {
                dir ('src/eaas') {
                    sh("make test")
                }
            }
        }
        stage('Race check') {
            steps {
                dir ('src/eaas') {
                    sh("make race") 
                }
            }
        }
        stage('Metalint check') {
            steps {
                dir ('src/eaas') {
                    sh("make metalint")
                }
            }
        }
        

        // PR On integration
        stage('Create and Deploy PR integration App') {

            when {
                allOf {
                    not { branch 'master' }
                    expression { return env.BRANCH_NAME.startsWith("PR-") }
                    expression { return env.CHANGE_TARGET.equals("integration") }
                }
            }
            steps {
                script {
                    tsuru.withAPI('integration') {
                        echo "Deploying application in ${tsuru.tsuruApi()}'s"
                        tsuru.connect()
                        appName = tsuru.createPRApp(env.JOB_NAME.tokenize('/')[1], env.BRANCH_NAME)
                        tsuru.deploy(appName, createDeployMessage(env))
                    }
                }
            }

        }

        // Promoting PR to Integration
        stage('Deploying Integration') {
            when { branch 'staging' }
            steps {
                script {
                    tsuru.withAPI('staging') {
                        appName = env.JOB_NAME.tokenize('/')[1]
                        echo "Deploying application in ${tsuru.tsuruApi()}'s to deploy application ${appName}"
                        tsuru.connect()
                        tsuru.deploy(appName, createDeployMessage(env))
                    }
                }
            }

        }

        // Promoting Integration to Production
        stage('Deploying Production') {
            when { branch 'release' }
            steps {
                timeout(time:5, unit:'DAYS') {
                    input message:'Approve deployment?', submitter: 'it-ops'
                }
                script {
                    tsuru.withAPI('production') {
                        appName = env.JOB_NAME.tokenize('/')[1]
                        echo "Deploying application in ${tsuru.tsuruApi()}'s to deploy application ${appName}"
                        tsuru.connect()
                        tsuru.deploy(appName, createDeployMessage(env))
                    }
                }
            }
        }
    }

    post {
        failure {
            mail to: 'none@whatever.com',
                subject: "Failed Pipeline: ${currentBuild.fullDisplayName}",
                body: "Something is wrong with ${env.BUILD_URL}"
        }
    }
}

