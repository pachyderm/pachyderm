build:
	go build -buildmode=plugin ./

docker-build: build
	docker build . -t pachyderm_filter
