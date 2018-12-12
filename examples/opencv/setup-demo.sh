#minikube delete
minikube start
sleep 10 #minikube doesn't seem to fully startup without a sleep
kubectl get all
which pachctl
pachctl deploy
until timeout 1s ../../../etc/kube/check_pachd_ready.sh; do sleep 1; done
pachctl port-forward &
pachctl version
# Setup Mount
umount ~/pfs
rm -rf ~/pfs
mkdir ~/pfs
pachctl mount ~/pfs &
open -a /Applications/Google\ Chrome.app "file:///$HOME/pfs"
# Build the opencv image (this took me ~17m w a beefy VM)
docker build -t opencv .
docker tag opencv `whoami`/opencv
docker push `whoami`/opencv
# Step 0
pachctl create-repo images 
# Step 1:
# pachctl put-file images master -c -i images.txt

