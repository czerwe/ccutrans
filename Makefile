
build:
	docker build --no-cache -t 10.0.0.4:5000/ccutrans:0.3.0 .
	docker push 10.0.0.4:5000/ccutrans:0.3.0

buildarm:
	docker build --no-cache --build-arg "GOARCH=arm" -t 10.0.0.4:5000/ccutrans:0.3.0-arm .
	docker push 10.0.0.4:5000/ccutrans:0.3.0-arm