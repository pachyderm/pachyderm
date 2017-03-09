# Notes

These manifests were pulled from [here](https://github.com/kubernetes/heapster/tree/b15cdc04dceb49fc1396d41a2c3c672c3b87d73a/deploy/kube-config/influxdb) on roughly that commit (may have been a few days earlier).

To update them, I'd recommend:

1) pulling the latest versions from that place
2) putting them here, and running `make launch-monitoring`, and navigating to `localhost:3000` to confirm it works
3) if so, update the commit hash / link in this readme and submit as PR
