all: compile docker push clean

compile:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o preprocess

docker:
	docker build --force-rm=true -t pachyderm/kubeflow-worker .

push:
	docker push pachyderm/kubeflow-worker

clean:
	rm preprocess
