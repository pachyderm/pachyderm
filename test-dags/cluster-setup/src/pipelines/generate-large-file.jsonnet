function(name, inputRepo, sizeFile, cpuLimit, memoryLimit) {
   pipeline: {
      name: name,
    },
    description: 'A pipeline for generating a large file in the load test.',
    transform: {
      cmd: ['sh'],
      stdin: [
        'echo "Creating 1 ' + sizeFile + 'B files..."',
        'openssl rand -base64 ' + sizeFile + ' > /pfs/out/0-0.txt',
      ],
      image: 'alpine/openssl:latest',
    },
    input: {
      pfs: {
        repo: inputRepo,
        glob: '/*',
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
