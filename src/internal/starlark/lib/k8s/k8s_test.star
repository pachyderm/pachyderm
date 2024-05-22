load("cmp", "diff")

def test_get_pod(t):
    name = "some-pod-abc123"
    got = k8s.pods.get(name)
    if not (got):
        t.Errorf("pod object is unexpectedly false")
    want = k8s.resource({
        "apiVersion": "v1",
        "kind": "Pod",
        "metadata": {
            "name": "some-pod-abc123",
            "namespace": "default",
            "creationTimestamp": None,
        },
        "spec": {"containers": [
            {
                "name": "test",
                "resources": {},
            },
        ]},
        "status": {
            "phase": "Running",
        },
    })
    d = diff(want, got)
    if d != "":
        t.Errorf("pod %v (-want +got):\n%s", name, d)

def test_get_pod_field(t):
    name = "some-pod-abc123"
    got = k8s.pods.get(name)["spec"]["containers"][0]["name"]
    want = "test"
    if got != want:
        t.Errorf("get pod %v spec.containers[0].name:\n  got: %v\n want: %v", name, got, want)

def test_list_pods(t):
    list = k8s.pods.list(labels = {"suite": "test"})
    if not list:
        t.Errorf("list is unexpectedly false")
    got = [x["metadata"]["name"] for x in list]
    want = ["some-test-pod"]
    if got != want:
        t.Errorf("list pods:\n  got: %v, want: %v", got, want)

def test_errors_work_like_objects(t):
    # Ensure single objects can be used as multi-level dicts.
    for x in [k8s.pods.get("some-test-pod"), k8s.pods.get("does-not-exist")]:
        # Should be able to get keys.
        x["metadata"]["namespace"]

        # Should be able to JSON encode.
        json.encode(x)

    # Ensure lists can be used as objects.
    for x in [k8s.pods.list(), k8s.pods.list(labels = "label=does-not-exist")]:
        # Should be able to JSON encode.
        json.encode(x)

        # Should be able to iterate as list.
        for value in x:
            value["metadata"]["namespace"]

    # Should be able to index a pod.
    x = k8s.pods.list()
    x["some-pod-abc123"]["metadata"]["name"]

def test_errors_are_false(t):
    got = k8s.pods.get("does-not-exist")
    if got:
        t.Error("error is unexpectedly true")

def test_string(t):
    str(k8s.pods)
    str(k8s.pods.get("some-pod-abc123"))
    str(k8s.pods.get("does-not-exist"))
    str(k8s.pods.list())
