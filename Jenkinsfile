node ('docker') {
	stage('build application') {
	    withDockerContainer(args: "--user root", image: 'golang:1.8.0') {
			sh 'git clone  https://github.com/czerwe/ccutrans.git  /go/src/ccutrans'
	        sh 'cd /go/src/ccutrans'
	        sh 'go get -d'
	        sh 'CGO_ENABLED=0 GOOS=linux go build ccutrans.go'
	    }
	}

	stage('Build Docker Image ') {
	    docker.build("ccutrans:${BUILD_NUMBER}")
	}
}