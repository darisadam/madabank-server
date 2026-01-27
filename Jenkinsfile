pipeline {
    agent any
    
    environment {
        // Docker Image Config
        IMAGE_NAME = 'darisadam/madabank-server'
        REGISTRY = 'ghcr.io'
        
        // Environment Files
        ENV_DEV = credentials('madabank-env-dev')
        ENV_STAGING = credentials('madabank-env-staging')
        ENV_PROD = credentials('madabank-env-prod')
    }
    
    options {
        buildDiscarder(logRotator(numToKeepStr: '5'))
        disableConcurrentBuilds()
    }
    
    stages {
        stage('Build & Test') {
            steps {
                script {
                    docker.build("${REGISTRY}/${IMAGE_NAME}:${env.BUILD_NUMBER}").inside {
                        sh 'go version'
                        sh 'go test ./...'
                    }
                }
            }
        }
        
        stage('Push Image') {
            when {
                anyOf {
                    branch 'develop'
                    branch 'staging'
                    branch 'main'
                }
            }
            steps {
                script {
                    docker.withRegistry('https://ghcr.io', 'github-registry-credentials') {
                         def customImage = docker.build("${REGISTRY}/${IMAGE_NAME}:${env.BUILD_NUMBER}")
                         customImage.push()
                         customImage.push('latest')
                    }
                }
            }
        }

        stage('Deploy Staging') {
            when {
                branch 'staging'
            }
            steps {
                sh 'echo "Deploying to Staging Env on VPS..."'
                // Command to update docker compose for staging service
                sh "COMMITHASH=${env.GIT_COMMIT} docker compose --env-file ${ENV_STAGING} -f docker/docker-compose-staging.yml up -d --pull always"
            }
        }

        stage('Deploy Production') {
            when {
                branch 'main'
            }
            steps {
                sh 'echo "Deploying to Production Env on VPS..."'
                // Command to update docker compose for prod service
                sh "COMMITHASH=${env.GIT_COMMIT} docker compose --env-file ${ENV_PROD} -f docker/docker-compose.yml up -d --pull always"
            }
        }
        
        stage('Release & Tag') {
            when {
                branch 'main'
            }
            steps {
                withCredentials([usernamePassword(credentialsId: 'github-git-creds', passwordVariable: 'GIT_PASSWORD', usernameVariable: 'GIT_USERNAME')]) {
                    sh """
                        git config user.email "jenkins@madabank.art"
                        git config user.name "Jenkins Bot"
                        git tag -a v1.0.${env.BUILD_NUMBER} -m "Release v1.0.${env.BUILD_NUMBER}"
                        git push https://${GIT_USERNAME}:${GIT_PASSWORD}@github.com/darisadam/madabank-server.git v1.0.${env.BUILD_NUMBER}
                    """
                }
            }
        }
    }
    
    post {
        always {
            cleanWs()
        }
    }
}
