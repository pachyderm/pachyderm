# To update, bump this SHA to the one you want to use, then run 'make update'
SHA:=031aa18ef763c0dba1ad54fedeeb0831b6790832

update:
	rm *.yaml
	curl https://raw.githubusercontent.com/kubernetes/heapster/$(SHA)/deploy/kube-config/influxdb/grafana.yaml -o grafana.yaml
	curl https://raw.githubusercontent.com/kubernetes/heapster/$(SHA)/deploy/kube-config/influxdb/heapster.yaml -o heapster.yaml
	curl https://raw.githubusercontent.com/kubernetes/heapster/$(SHA)/deploy/kube-config/influxdb/influxdb.yaml -o influxdb.yaml
	curl https://raw.githubusercontent.com/kubernetes/heapster/$(SHA)/deploy/kube-config/rbac/heapster-rbac.yaml -o heapster-rbac.yaml

