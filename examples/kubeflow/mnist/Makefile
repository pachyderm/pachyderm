.PHONY: docker-build docker-push kubeflow-build delete-old-kubeflow-jobs mnist

docker-build:
	cd kubeflow_pipeline && make docker-build
	cd pachyderm_pipeline && make docker-build

docker-push:
	cd kubeflow_pipeline && make docker-push
	cd pachyderm_pipeline && make docker-push

kubeflow-build:
	cd kubeflow_pipeline && make kubeflow-build

clean:
	kubectl delete pod $$(kubectl get pod --output=json | jq '.items[].metadata.name' -r | grep mnist-kubeflow-pipeline) || exit 0
	echo "y\n" | pachctl delete all

mnist:
	pachctl create repo input-repo
	pachctl put file input-repo@master:/mnist.npz -f mnist.npz
	cd pachyderm_pipeline && pachctl create pipeline -f pipeline.yaml
