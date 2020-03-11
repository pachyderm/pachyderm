
seldon-core:
	git clone -b v1.0.1 --single-branch git@github.com:SeldonIO/seldon-core.git

seldon-core-operator/base: seldon-core clean-kustomize
	mkdir -p seldon-core-operator/base
	helm template -f values.yaml seldon-core/helm-charts/seldon-core-operator --namespace kubeflow > seldon-core-operator/base/resources.yaml
	cp kustomization.tpl seldon-core-operator/base/kustomization.yaml
	echo "Need to hand edit at present"
	echo "Add kubeflow specific namespace selectors"
	echo "Add sidecar.istio.io/inject: 'false' seldon operator deployment"

.PHONY: clean
clean:
	rm -rf seldon-core

.PHONY:clean-kustomize
clean-kustomize:
	rm -rf seldon-core-operator/base
