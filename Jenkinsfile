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


//options {
//  timestamps()
//}


stage("Prepare environment") {
    def environment  = docker.image 'tsuru/go:latest'
    environment.inside {
        stage('Install stuffs')
            steps
                sh("sudo apt-get update || true; sudo apt-get install build-essential -y") 
        stage('Run tests') 
            steps 
                sh("make test")
            
        stage('Run Race check') 
            steps 
                sh("make race")
            
        
        stage('Run lint check') 
            steps 
                sh("make metalint")
    
    }    
}

// PR On integration
stage('Create and Deploy PR integration App') {

    when {
    allOf {
        not {
            branch 'master'
        }
        expression {
            return env.BRANCH_NAME.startsWith("PR-")
        }
        expression {
            return env.CHANGE_TARGET.equals("integration")
        }
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
    when {
    branch 'staging'
    }
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
    when {
    branch 'release'
    }
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

post {
    failure {
        mail to: 'george.fleury@trustyou.com',
            subject: "Failed Pipeline: ${currentBuild.fullDisplayName}",
            body: "Something is wrong with ${env.BUILD_URL}"
    }
}

