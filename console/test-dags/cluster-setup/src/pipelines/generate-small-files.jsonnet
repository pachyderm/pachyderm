function(name, inputRepo, numFiles, sizeFile, cpuLimit, memoryLimit, workers) {
   pipeline: {
      name: name,
    },
    description: 'A pipeline for generating small files in the load test.',
    transform: {
      cmd: ['sh'],
      stdin: [
        'echo "Creating ' + numFiles + ' + 1 ' + sizeFile + 'B files..."',
        'for i in `seq 0 ' + numFiles + '`;do openssl rand -base64 ' + sizeFile + ' > /pfs/out/$(cat /pfs/' + inputRepo + '/input-*.txt)-$i.txt; done',
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
    parallelism_spec: {
      constant: workers
    },
    autoscaling: true
}
