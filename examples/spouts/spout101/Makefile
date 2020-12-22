build-minikube:
	@eval $$(minikube docker-env) ;\
	docker image build -t spout101:0.1 .

deploy:
	pachctl create pipeline -f ./pipelines/spout.json
	pachctl create pipeline -f ./pipelines/processor.json