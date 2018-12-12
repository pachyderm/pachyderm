all: compile docker push

compile:
	CGO_ENABLED=0 go build -o sales

docker:
	sudo docker build --no-cache=true --force-rm=true -t dwhitena/bike-share-sales .

push:
	sudo docker push dwhitena/bike-share-sales
