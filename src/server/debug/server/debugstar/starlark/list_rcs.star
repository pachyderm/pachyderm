def dump_rcs():
    rcs = k8s.replicationcontrollers.list(labels = {"suite": "pachyderm", "app": "pipeline"})
    for rc in rcs:
        name = rc["metadata"]["name"]
        dump("%s/rc.json" % name, json.encode(rc))
        pods = k8s.pods.list(labels = rc["spec"]["selector"])
        for pod in pods:
            podname = pod["metadata"]["name"]
            dump("%s/pod-%s.json" % (name, podname), json.encode(pod))

if "replicationcontrollers" in dir(k8s) and "pods" in dir(k8s):
    dump_rcs()
