pipeline {
    agent none
    stages {
        stage('Run Tests') {
            parallel {
                stage('Go Test') {
                    agent {
                        label "ec2-flagship-ubuntu"
                    }
                    steps {
                        sh "make docker-test"
                    }
                    post {
                        always {
                            sh  "echo reached post"
                        }
                    }
                }
            }
        }
    }
}