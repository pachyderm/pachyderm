"""
The k8s module allows access to the Kubernetes API.

In personalities where Kubernetes is available, a k8s object will be available to code.  The k8s
object contains an attribute for each resource discovered on the server, like "pods", "services",
etc.  Each of these attributes contains a client for accessing that resource.  No access outside the
configured namespace is available, except for namespace-less resources (like "nodes").

The k8s module is built by talking to the server to see what resources it has.  You must test for
existence of a resource before using it:

    if "pods" in dir(k8s):
        k8s.pods.list()

"""

def Unstructured():
    """Unstructured is a Kubernetes object without compile-time Go type information.

    Unstructured is not a function; this is a placeholder for documentation.

    An Unstructured object is logically a dict; use like:
    k8s.pods.get("some-pod")["metadata"]["name"] to get the name of a pod.

    k8s.pods.get("some-pod") is an Unstructured, but k8s.pods.get("some-pod")["metadata"] is
    an ordinary dict.

    type(Unstructured) is the runtime type of the object, like /v1/Pod.
    str(Unstructured) is the Go representation of the object, not the JSON representation of the object.
    json.encode(Unstructured) formats the JSON as kubectl would.
    A non-error Unstructured object is non-False (if pod.get("foo") is true, you have the pod.)

    If writing were implemented, this would work:
        pachd = k8s.deployments.get("pachd")
        pachd["replias"] = 42
        k8s.deployments.put(pachd)
    """

def resource(object: str | dict) -> Unstructured:
    """Resource returns a runtime object based on the provided JSON or dictionary.

    This is mostly intended for testing, to create literals to cmp.diff() against.

    Args:
        object: A Kubernetes object, as JSON or a dict.
    Returns:
        An Unstructured object.
    """

def logs(podName: str, container: str, previous: bool = False) -> bytes:
    """Logs returns up to 10MB of logs for the named container.

    Logs may return an error message instead of the actual logs, like get and list below.

    Args:
        container: The container to return logs from.  We don't attempt to guess the container
            because customer installs may have sidecars, so we can't assume that there is only
            one container and that randomly selecting one of the Pod's containers will be useful.
            Get the pod first and iterate over its containers.
        previous: If true, return logs from the previous instance of the container.

    Returns:
        A bytestring containing the logs.
    """

def get(name: str, resourceVersion: None | str = None) -> Unstructured:
    """Get returns the named resource at the provided version.

    Get can return an error instead of an object; the error has the same API as the Unstructured
    object, but returns itself for every operation.  The error value is treated specially during
    JSON marshaling or debug dumping.  For example, k8s.pods.get("foo")["metadata"]["name"] will
    work whether "foo" is found or not, but if you json.encode(that thing), it will encode as
    `{"error": "pod foo not found"}` instead of an actual pod.

    Args:
        name: The name of the resource.
        resourceVersion: The resource version to fetch.

    Returns:
        An Unstructured object.
    """

def list(labels: None | dict | str = None, fields: None | dict | str = None) -> Iterable[Unstructured]:
    """List returns an Iterable of all resrouces of the client type.

    The returned collection may be an error, as with get() above.  An error has the same API as a
    non-error, but again with special treatment during marshaling or debug dumping.

    Args:
        labels: A label selector specifying which objects to match.  Strings are parsed as `kubectl -l`,
            while dicts are parsed as the Go client would treat a labels.Selector{} object.
            `{"suite": "pachyderm"}` (dict) equals `-l suite=pachyderm`.
        fields: A field selector specifying which objects to match.  The syntax is the same as a label
            selector.

    Returns:
        A collection of Unstructured objects.  The collection is indexable by name (result["some-pod"])
        and is iterable (`for pod in k8s.pods.list()`).
    """
