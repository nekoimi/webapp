#!/usr/bin/groovy

node {
    stage('Checkout') {
        checkout scm
    }

    stage('Build') {
        docker.withRegistry('', 'dockerhub_access') {
            docker.build("nekoimi/webapp:latest").push()
        }
    }

    stage('Clean') {
        cleanWs()
    }
}
