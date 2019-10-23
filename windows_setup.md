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

minikube start

kubectl is missing apps/v1beta1 version...
 * apps/v1 should probably work?

personal options:
settings.json - all vscode settings
~/.inputrc - disable bash terminal bell