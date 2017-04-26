node ('docker') {
	stage('Checkout') {
	    git credentialsId: '1db449c1-a2af-4972-b32b-f7cdd65473e8', url: 'git@github.com:czerwe/ccutrans.git'
	}

	stage('build application') {
	    withDockerContainer(args: "--user root", image: 'golang:1.8.0') {
			sh 'git clone  https://github.com/czerwe/ccutrans.git  /go/src/ccutrans'
	        sh 'cd /go/src/ccutrans'
	        sh 'go get -d'
	        // sh 'go get github.com/influxdata/influxdb/client/v2'
	        // sh 'go get github.com/jessevdk/go-flags'
	        // sh 'go get github.com/gorilla/mux'
	        // sh 'go get github.com/Sirupsen/logrus'
	        sh 'CGO_ENABLED=0 GOOS=linux go build ccutrans.go'
	    }
	}

	stage('Build Docker Image ') {
	    docker.build("ccutrans:${BUILD_NUMBER}")
	}
}