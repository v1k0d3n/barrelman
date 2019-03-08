pipeline {
    environment {
     IPADDRESS = ""
    }
    agent none
    stages {
        stage('Init Flagship') {
            parallel {
                stage('Test On Ubuntu') {
                    agent {
                        label "ec2-flagship-ubuntu"
                    }
                    steps {
                            script {
                                sh "make test"
                            }
                        }
                    }
                    post {
                        always {
                            sh "echo cool"
                        }
                    }
                }
            }
        }
    }
}
