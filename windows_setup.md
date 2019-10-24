install:
vscode
git
docker toolbox https://github.com/docker/toolbox/releases
minikube
kubectl https://kubernetes.io/docs/tasks/tools/install-kubectl/ (raw exe, copy somewhere in PATH, like the minikube dir)

vscode extensions:
go
docker

vscode settings:
configure shell to use 'git bash'

git clone https://github.com/pachyderm/pachyderm

run docker quickstart terminal

after restart, docker commands hang:
 * run docker quickstart terminal after every restart

 compile containers don't have enough memory?  get signal: killed

debugging CreateContainerError:
 * use `minikube logs`
 * problem is that generated kube deploy file contains paths like '\var\pachyderm\pachd' which is an invalid path I guess

debugging CrashLoopBackOff in etcd pod
 * `kubectl logs -p <pod>` - gets logs from previous attempt
 * problem is that shared volumes with the host cannot be mmapped?
 * if we remove the volume, the pod launches (but data won't persist?)

personal options:
settings.json - all vscode settings
~/.inputrc - disable bash terminal bell

minikube delete
minikube start --memory=8192mb
eval $(minikube docker-env --shell bash)