Troubleshooting
===============

common rookie mistakes
----------------------

Error: 

Solution:


Client-Server version mismatch -unmarshalling...

Setting gopath correctly

mount issues

Stuck in "pulling" -- job-shim image

port forwarding not set up correctly -- maybe be different for k8s in docker and minikube

requiring "timeout"

From chris:

"Right now if I make a new terminal window, it seems like it wouldn't be able to find pachyderm anymore."

"If the pipeline fails for some reason, I have no idea how to debug that, fix it, or reset it to a known state. "

Cluster gets into bad state such as killing the VM:
minikube stop
minikube delete
minikube start



pachctl delete all

pachctl archive all

reset k8s cluster: pachctl deploy --dry-run | kubectl delete -f -

