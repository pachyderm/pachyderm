from kubernetes import client
def modify_pod_hook(spawner, pod):
	pod.spec.containers[0].security_context = client.V1SecurityContext(
		allow_privilege_escalation=True,
		run_as_user=0,
		privileged=True,
		capabilities=client.V1Capabilities(
			add=['SYS_ADMIN']
		)
	)
	return pod
c.KubeSpawner.modify_pod_hook = modify_pod_hook
