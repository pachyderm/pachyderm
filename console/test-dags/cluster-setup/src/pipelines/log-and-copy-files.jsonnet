function(name, inputRepo, inputFile, sizeLogs, cpuLimit, memoryLimit) {
   pipeline: {
      name: name,
    },
    description: 'A pipeline for logging and copying over generated files.',
    transform: {
      cmd: ['sh'],
      stdin: [
        'head -c ' + sizeLogs + ' /pfs/' + inputRepo + '/' + inputFile,
        'cp ' + ' /pfs/' + inputRepo + '/' + inputFile + ' /pfs/out/' + inputFile,
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
