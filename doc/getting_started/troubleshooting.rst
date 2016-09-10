Troubleshooting
===============

Below a list of common errors we run accross with users trying to run Pachyderm locally and following the :doc:`beginner_tutorial`. 


Error: 

Solution:



Client-Server version mismatch -unmarshalling...

Setting gopath correctly, which pachctl, which kubectl

Grep term not found. 

Stuck in "pulling" -- job-shim image

Not being able to unmount correctly

requiring "timeout" -- not a problem anymore?


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

