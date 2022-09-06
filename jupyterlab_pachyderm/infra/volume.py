def profile_pvc(spawner):
    profile = spawner.user_options.get("profile", "")
    if profile in [
        "sidecar",
    ]:
        spawner.volumes.extend([
            {
                "name": "shared-pfs",
                "emptyDir": {},
            }
        ])
        spawner.volume_mounts.extend([
            {
                "name":  "shared-pfs",
                "mountPath":        "/pfs",
                "mountPropagation": "HostToContainer",
            }
        ])
c.KubeSpawner.pre_spawn_hook = profile_pvc
