.PHONY: docker-build tfjobs-clean tfjob

docker-build:
	docker build --tag ysimonson/pach_kflow_example:v3.1.4 .
	docker push ysimonson/pach_kflow_example:v3.1.4

tfjobs-clean:
	kubectl delete tfjob --all

tfjob:
	kubectl create -f mnist_tf_job_s3_gateway.yaml
