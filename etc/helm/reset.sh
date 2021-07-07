#/bin/sh

helm delete pachyderm
echo "Wait for pod teardown"
kubectl wait pod --for delete -l suite=pachyderm
echo "Remove Persistent Volume Claims"
kubectl delete pvc -l suite=pachyderm
helm install pachyderm pachyderm --set pachd.storage.backend=LOCAL