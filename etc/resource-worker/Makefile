all: compile docker push clean

compile:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o preprocess

docker:
	sudo docker build --force-rm=true -t pachyderm/kubeflow-worker .

push:
	sudo docker push pachyderm/kubeflow-worker

clean:
	rm preprocess
