function(name, inputRepo, cpuLimit, memoryLimit) {
   pipeline: {
      name: name,
    },
    description: 'A pipeline that echoes its name and then saves it to an output file.',
    transform: {
      cmd: ['sh'],
      stdin: [
        'echo "' + name + '"',
        'echo "' + name + '" > /pfs/out/name.txt',
      ],
      image: 'alpine:latest',
    },
    input: {
      pfs: {
        repo: inputRepo,
        glob: '/',
      }
    },
    resource_limits: {
      cpu: cpuLimit,
      memory: memoryLimit,
    },
    resource_requests: {
      cpu: cpuLimit,
      memory: memoryLimit,
    },
    sidecar_resource_requests: {
      cpu: cpuLimit,
      memory: memoryLimit,
    },
    autoscaling: true
}
