// =============================================================================
// MadaBank Server - Jenkins CI/CD Pipeline
// =============================================================================
// Flow:
//   feature → develop: CI + auto rebase
//   develop → staging: CI + auto merge commit
//   staging → main:    CI (cached) + CD to VPS + auto merge + tag + release
// =============================================================================

pipeline {
    agent any
    
    environment {
        // Docker Image Config
        IMAGE_NAME = 'darisadam/madabank-server'
        REGISTRY = 'ghcr.io'
        FULL_IMAGE = "${REGISTRY}/${IMAGE_NAME}"
        
        // Go Version
        GO_VERSION = '1.24.0'
        
        // VPS Production Directory (existing infrastructure)
        DEPLOY_DIR = '/opt/madabankapp'
        
        // Environment Files (Removed: unused)
        // ENV_PROD = credentials('madabank-env-prod')
        
        // Docker Registry Credentials (Removed: reusing github-git-creds)
        // DOCKER_USERNAME = credentials('github-registry-username')
        // DOCKER_PASSWORD = credentials('github-registry-password')
    }
    
    options {
        buildDiscarder(logRotator(numToKeepStr: '10'))
        disableConcurrentBuilds()
        timestamps()
        timeout(time: 30, unit: 'MINUTES')
    }

    stages {
        // =====================================================================
        // STAGE 1: Checkout & Environment Info
        // =====================================================================
        stage('Checkout') {
            steps {
                checkout([
                    $class: 'GitSCM',
                    branches: scm.branches,
                    doGenerateSubmoduleConfigurations: false,
                    extensions: [
                        [$class: 'CloneOption', depth: 1, noTags: false, reference: '', shallow: true, timeout: 60],
                        [$class: 'WipeWorkspace']
                    ],
                    submoduleCfg: [],
                    userRemoteConfigs: scm.userRemoteConfigs
                ])
                script {
                    env.GIT_COMMIT_SHORT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    env.GIT_BRANCH_NAME = sh(script: 'git rev-parse --abbrev-ref HEAD', returnStdout: true).trim()
                    env.GIT_COMMIT_MSG = sh(script: 'git log -1 --pretty=%s', returnStdout: true).trim()
                }
                echo "Branch: ${env.GIT_BRANCH_NAME} (Env: ${env.BRANCH_NAME})"
                echo "Commit: ${env.GIT_COMMIT_SHORT} - ${env.GIT_COMMIT_MSG}"
                echo "PR Target: ${env.CHANGE_TARGET} (Change ID: ${env.CHANGE_ID})"
            }
        }

        // =====================================================================
        // STAGE 2: Setup Go Environment
        // =====================================================================
        stage('Setup Go') {
            steps {
                sh '''
                    # Add common Go paths to PATH
                    export PATH=$PATH:$HOME/go/bin:$HOME/sdk/go${GO_VERSION}/bin
                    
                    if ! command -v go &> /dev/null; then
                        echo "Installing Go ${GO_VERSION} locally..."
                        curl -sLO https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
                        
                        # Install to $HOME/go_dist instead of /usr/local/go
                        mkdir -p $HOME/go_dist
                        tar -C $HOME/go_dist -xzf go${GO_VERSION}.linux-amd64.tar.gz
                        rm go${GO_VERSION}.linux-amd64.tar.gz
                        
                        # Update PATH for this script execution
                        export PATH=$PATH:$HOME/go_dist/go/bin
                    fi
                    
                    go version
                '''
            }
        }

        // =====================================================================
        // STAGE 3: Install Dependencies
        // =====================================================================
        stage('Dependencies') {
            steps {
                sh '''
                    export PATH=$PATH:$HOME/go_dist/go/bin:$HOME/go/bin
                    go mod download
                    go mod verify
                '''
            }
        }

        // =====================================================================
        // STAGE 4: Code Quality (Lint & Format)
        // =====================================================================
        stage('Lint') {
            steps {
                sh '''
                    export PATH=$PATH:$HOME/go_dist/go/bin:$HOME/go/bin
                    # Reduce memory usage (aggressive GC)
                    export GOGC=20
                    
                    # Check code formatting
                    fmt_output=$(gofmt -l .)
                    if [ -n "$fmt_output" ]; then
                        echo "❌ Code not formatted:"
                        echo "$fmt_output"
                        exit 1
                    fi
                    
                    # Run go vet
                    go vet ./...
                    
                    # Install & run golangci-lint
                    if ! command -v golangci-lint &> /dev/null; then
                        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $HOME/go/bin v1.64.5
                    fi
                    golangci-lint run --timeout=5m --concurrency=2
                '''
            }
        }

        // =====================================================================
        // STAGE 5: Unit Tests
        // =====================================================================
        stage('Test') {
            steps {
                sh '''
                    export PATH=$PATH:$HOME/go_dist/go/bin:$HOME/go/bin
                    # Note: -race removed to avoid CGO/GCC requirement on Jenkins agent
                    go test -v -coverprofile=coverage.out -covermode=atomic ./...
                '''
            }
            post {
                always {
                    sh 'export PATH=$PATH:$HOME/go_dist/go/bin && go tool cover -html=coverage.out -o coverage.html 2>/dev/null || true'
                    archiveArtifacts artifacts: 'coverage.html,coverage.out', allowEmptyArchive: true
                }
            }
        }

        // =====================================================================
        // STAGE 6: Security Scan
        // =====================================================================
        stage('Security') {
            steps {
                sh '''
                    export PATH=$PATH:$HOME/go_dist/go/bin:$HOME/go/bin
                    
                    # Install gosec if not present
                    if ! command -v gosec &> /dev/null; then
                        go install github.com/securego/gosec/v2/cmd/gosec@latest
                    fi
                    
                    # Run security scan
                    gosec -fmt json -out gosec-report.json ./... || true
                    
                    # Install govulncheck if not present
                    if ! command -v govulncheck &> /dev/null; then
                        go install golang.org/x/vuln/cmd/govulncheck@latest
                    fi
                    
                    # Check for vulnerabilities
                    govulncheck ./... || true
                '''
            }
            post {
                always {
                    archiveArtifacts artifacts: 'gosec-report.json', allowEmptyArchive: true
                }
            }
        }

        // =====================================================================
        // STAGE 7: Build Binary
        // =====================================================================
        stage('Build Binary') {
            steps {
                sh '''
                    export PATH=$PATH:$HOME/go_dist/go/bin:$HOME/go/bin
                    export CGO_ENABLED=0
                    export GOOS=linux
                    export GOARCH=amd64
                    
                    mkdir -p bin
                    go build -ldflags "-s -w -X main.version=${BUILD_NUMBER} -X main.commit=${GIT_COMMIT_SHORT}" \
                        -o bin/api-linux-amd64 cmd/api/main.go
                    go build -ldflags "-s -w" -o bin/migrate-linux-amd64 cmd/migrate/main.go
                    
                    echo "✅ Binaries built:"
                    ls -lh bin/
                '''
            }
        }

        // =====================================================================
        // STAGE 8: Build & Push Docker Image
        // =====================================================================
        stage('Docker Build & Push') {
            when {
                branch 'staging'
            }
            steps {
                script {
                    // Use 'github-git-creds' for GHCR as well (same user/PAT)
                    docker.withRegistry("https://${REGISTRY}", 'github-git-creds') {
                        // For PRs, use the PR number or commit hash as tag
                        def tag = env.CHANGE_ID ? "pr-${env.CHANGE_ID}" : "staging-${BUILD_NUMBER}"
                        
                        // Pass build args for OS/Arch
                        def customImage = docker.build("${FULL_IMAGE}:${tag}", "--build-arg TARGETOS=linux --build-arg TARGETARCH=amd64 -f docker/Dockerfile.fast .")
                        customImage.push()
                        
                        // Also push as 'latest-staging' or 'latest-pr' for easy reference if needed
                        if (env.BRANCH_NAME == 'staging') {
                            customImage.push('staging-latest')
                        }
                    }
                }
                echo "✅ Docker image pushed."
            }
        }

        // =====================================================================
        // STAGE 9: Deploy to Production (PR Staging -> Main)
        // =====================================================================
        stage('Deploy Production') {
            when {
                branch 'staging'
            }
            steps {
                echo "🚀 Deploying to Production VPS (PR Preview/Release Candidate)..."
                
                script {
                    // Determine tag based on build type (must match Docker Build stage logic)
                    def imageTag = env.CHANGE_ID ? "pr-${env.CHANGE_ID}" : "staging-${env.BUILD_NUMBER}"
                    
                    withCredentials([usernamePassword(credentialsId: 'github-git-creds', 
                                                       passwordVariable: 'DOCKER_PASSWORD', 
                                                       usernameVariable: 'DOCKER_USERNAME')]) {
                        // Deploy API service using docker compose
                        env.IMAGE_TAG = imageTag
                        sh '''
                            cd $DEPLOY_DIR
                            
                            # Login to GHCR (Password via env var is safe)
                            echo $DOCKER_PASSWORD | docker login ghcr.io -u $DOCKER_USERNAME --password-stdin
                            
                            # Pull the new image
                            docker pull $FULL_IMAGE:$IMAGE_TAG
                            
                            # Retag as 'latest' locally on VPS
                            docker tag $FULL_IMAGE:$IMAGE_TAG $FULL_IMAGE:latest
                            
                            # Restart API service using a transient container with docker compose support
                            # This avoids needing to mount CLI plugins from the host
                            docker run --rm \\
                                -v /var/run/docker.sock:/var/run/docker.sock \\
                                -v "$DEPLOY_DIR:$DEPLOY_DIR" \\
                                -v /var/jenkins_home/.docker/config.json:/root/.docker/config.json \\
                                -w "$DEPLOY_DIR" \\
                                docker:latest \\
                                compose up -d --force-recreate api
                            
                            # Wait for health check
                            sleep 20
                            
                            # Verify deployment
                            curl -sf http://localhost:8080/health || exit 1
                            
                            echo "✅ Production deployment successful for PR-${CHANGE_ID}!"
                        '''
                    }
                }
            }
        }


        // =====================================================================
        // STAGE 10: Create Release Tag (only after merge to main)
        // =====================================================================
        stage('Release & Tag') {
            when {
                branch 'main'
            }
            steps {
                withCredentials([usernamePassword(credentialsId: 'github-git-creds', 
                                                   passwordVariable: 'GIT_PASSWORD', 
                                                   usernameVariable: 'GIT_USERNAME')]) {
                    sh """
                        git config user.email "jenkins@madabank.art"
                        git config user.name "Jenkins Bot"
                        
                        # Create annotated tag
                        git tag -a v1.0.${BUILD_NUMBER} -m "Release v1.0.${BUILD_NUMBER}\\nCommit: ${GIT_COMMIT_SHORT}"
                        
                        # Push tag
                        git push https://${GIT_USERNAME}:${GIT_PASSWORD}@github.com/${IMAGE_NAME}.git v1.0.${BUILD_NUMBER}
                        
                        echo "✅ Release tag created: v1.0.${BUILD_NUMBER}"
                    """
                }
            }
        }

        // =====================================================================
        // STAGE 11: Cleanup
        // =====================================================================
        stage('Cleanup') {
            steps {
                sh '''
                    # Remove old Docker images (keep last 5)
                    docker images ${FULL_IMAGE} --format "{{.ID}} {{.Tag}}" | \
                        grep -v latest | sort -t. -k3 -n | head -n -5 | \
                        awk '{print $1}' | xargs -r docker rmi 2>/dev/null || true
                    
                    # Prune dangling images
                    docker image prune -f 2>/dev/null || true
                '''
            }
        }
    }
    
    post {
        success {
            echo """
            ╔══════════════════════════════════════════════════════════════╗
            ║  ✅ PIPELINE SUCCESS                                         ║
            ║  Branch: ${env.GIT_BRANCH_NAME}                              ║
            ║  Commit: ${env.GIT_COMMIT_SHORT}                             ║
            ║  Build:  #${BUILD_NUMBER}                                    ║
            ╚══════════════════════════════════════════════════════════════╝
            """
        }
        failure {
            echo """
            ╔══════════════════════════════════════════════════════════════╗
            ║  ❌ PIPELINE FAILED                                          ║
            ║  Branch: ${env.GIT_BRANCH_NAME}                              ║
            ║  Commit: ${env.GIT_COMMIT_SHORT}                             ║
            ║  Check logs for details.                                     ║
            ╚══════════════════════════════════════════════════════════════╝
            """
        }
        always {
            script {
                try {
                    cleanWs()
                } catch (e) {
                    echo "Warning: Failed to clean workspace (likely early pipeline failure): ${e.getMessage()}"
                    // e.printStackTrace() // Suppress stack trace to reduce noise
                }
            }
        }
    }
}