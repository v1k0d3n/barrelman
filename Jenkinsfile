pipeline {
    agent {
        label "ec2-flagship-ubuntu"
    }
    stages {
        stage('Go Test') {
            steps {
                sh "sudo docker build -f Dockertest ."
            }
        }
        stage('GO Build') {
            steps {
                sh "sudo docker build -f Dockerfile ."
            }
        }
        stage('Deploy') {
            steps {
                sh "I am deploying"
            }
        }
    }
}