import {
  chunkSpecFromObject,
  cronInputFromObject,
  egressFromObject,
  gpuSpecFromObject,
  inputFromObject,
  parallelismSpecFromObject,
  pfsInputFromObject,
  pipelineFromObject,
  pipelineInfoFromObject,
  pipelineInfosFromObject,
  resourceSpecFromObject,
  secretMountFromObject,
  serviceFromObject,
  spoutFromObject,
  tfJobFromObject,
  transformFromObject,
  jobInfoFromObject,
  jobFromObject,
  getLogsRequestFromObject,
} from '../pps';

describe('grpc/builders/pps', () => {
  it('should create Pipeline from an object', () => {
    const pipeline = pipelineFromObject({
      name: 'testPipeline',
    });

    expect(pipeline.getName()).toBe('testPipeline');
  });

  it('should create SecretMount from an object', () => {
    const secretMount = secretMountFromObject({
      name: 'testSecret',
      key: 'test',
      mountPath: '/test',
      envVar: 'testVar',
    });

    expect(secretMount.getName()).toBe('testSecret');
    expect(secretMount.getKey()).toBe('test');
    expect(secretMount.getMountPath()).toBe('/test');
    expect(secretMount.getEnvVar()).toBe('testVar');
  });

  it('should create Transform from an object', () => {
    const transform = transformFromObject({
      image: 'pachyderm/opencv',
      cmdList: ['python3', '/edges.py'],
      errCmdList: ['python3', '/error.py'],
      secretsList: [
        {
          name: 'testSecret',
          key: 'test',
          mountPath: '/test',
          envVar: 'testVar',
        },
      ],
      imagePullSecretsList: ['asdkldsfsdf'],
      stdinList: ['python2', '/test.py'],
      errStdinList: ['python2', '/error.py'],
      acceptReturnCodeList: [2, 4],
      debug: true,
      user: 'peter',
      workingDir: '/test',
      dockerfile: 'docker',
    });

    expect(transform.getImage()).toBe('pachyderm/opencv');
    expect(transform.getCmdList()).toStrictEqual(['python3', '/edges.py']);
    expect(transform.getErrCmdList()).toStrictEqual(['python3', '/error.py']);
    expect(transform.getSecretsList()[0]?.getName()).toEqual('testSecret');
    expect(transform.getSecretsList()[0]?.getKey()).toEqual('test');
    expect(transform.getSecretsList()[0]?.getMountPath()).toEqual('/test');
    expect(transform.getSecretsList()[0]?.getEnvVar()).toEqual('testVar');
    expect(transform.getImagePullSecretsList()).toStrictEqual(['asdkldsfsdf']);
    expect(transform.getStdinList()).toStrictEqual(['python2', '/test.py']);
    expect(transform.getErrStdinList()).toStrictEqual(['python2', '/error.py']);
    expect(transform.getAcceptReturnCodeList()).toStrictEqual([2, 4]);
    expect(transform.getDebug()).toBe(true);
    expect(transform.getUser()).toBe('peter');
    expect(transform.getWorkingDir()).toBe('/test');
  });

  it('should create Transform from an object with defaults', () => {
    const transform = transformFromObject({
      image: 'pachyderm/opencv',
      cmdList: ['python3', '/edges.py'],
    });

    expect(transform.getImage()).toBe('pachyderm/opencv');
    expect(transform.getCmdList()).toStrictEqual(['python3', '/edges.py']);
    expect(transform.getErrCmdList()).toStrictEqual([]);
    expect(transform.getSecretsList()).toStrictEqual([]);
    expect(transform.getImagePullSecretsList()).toStrictEqual([]);
    expect(transform.getStdinList()).toStrictEqual([]);
    expect(transform.getErrStdinList()).toStrictEqual([]);
    expect(transform.getAcceptReturnCodeList()).toStrictEqual([]);
    expect(transform.getDebug()).toBe(false);
    expect(transform.getUser()).toBe('');
    expect(transform.getWorkingDir()).toBe('');
  });

  it('should create TFJob from an object', () => {
    const tfJob = tfJobFromObject({
      tfJob: 'example-job',
    });

    expect(tfJob.getTfJob()).toBe('example-job');
  });

  it('should create ParallelismSpec from an object', () => {
    const parallelismSpec = parallelismSpecFromObject({
      constant: 1,
    });

    expect(parallelismSpec.getConstant()).toBe(1);
  });

  it('should create Egress from an object', () => {
    const egress = egressFromObject({
      url: 's3://bucket/dir',
    });

    expect(egress.getUrl()).toBe('s3://bucket/dir');
  });

  it('should create GPUSpec from an object', () => {
    const gpuSpec = gpuSpecFromObject({
      type: 'good',
      number: 3,
    });

    expect(gpuSpec.getType()).toBe('good');
    expect(gpuSpec.getNumber()).toBe(3);
  });

  it('should create ResourceSpec from an object', () => {
    const resourceSpec = resourceSpecFromObject({
      cpu: 8,
      memory: '12mb',
      gpu: {type: 'good', number: 3},
      disk: '2t',
    });

    expect(resourceSpec.getCpu()).toBe(8);
    expect(resourceSpec.getMemory()).toBe('12mb');
    expect(resourceSpec.getGpu()?.getType()).toBe('good');
    expect(resourceSpec.getGpu()?.getNumber()).toBe(3);
    expect(resourceSpec.getDisk()).toBe('2t');
  });

  it('should create PFSInput from an object with defaults', () => {
    const pfsInput = pfsInputFromObject({
      name: 'images',
      repo: 'imagesRepo',
      branch: 'master',
      commit: 'uweioruwejrij098w0e9r809we',
      glob: '/test/*',
      joinOn: 'test',
      outerJoin: true,
      groupBy: 'name',
      lazy: true,
      emptyFiles: true,
      s3: true,
      trigger: {
        branch: 'master',
        all: true,
        cronSpec: '@every 10s',
        size: 'big',
        commits: 12,
      },
    });

    expect(pfsInput.getName()).toBe('images');
    expect(pfsInput.getRepo()).toBe('imagesRepo');
    expect(pfsInput.getBranch()).toBe('master');
    expect(pfsInput.getCommit()).toBe('uweioruwejrij098w0e9r809we');
    expect(pfsInput.getGlob()).toBe('/test/*');
    expect(pfsInput.getJoinOn()).toBe('test');
    expect(pfsInput.getOuterJoin()).toBe(true);
    expect(pfsInput.getGroupBy()).toBe('name');
    expect(pfsInput.getLazy()).toBe(true);
    expect(pfsInput.getEmptyFiles()).toBe(true);
    expect(pfsInput.getS3()).toBe(true);
    expect(pfsInput.getTrigger()?.getBranch()).toBe('master');
    expect(pfsInput.getTrigger()?.getAll()).toBe(true);
    expect(pfsInput.getTrigger()?.getCronSpec()).toBe('@every 10s');
    expect(pfsInput.getTrigger()?.getSize()).toBe('big');
    expect(pfsInput.getTrigger()?.getCommits()).toBe(12);
  });

  it('should create PFSInput from an object', () => {
    const pfsInput = pfsInputFromObject({
      name: 'images',
      repo: 'imagesRepo',
      branch: 'master',
      commit: 'uweioruwejrij098w0e9r809we',
      glob: '/test/*',
      joinOn: 'test',
      outerJoin: true,
      groupBy: 'name',
      lazy: true,
      emptyFiles: true,
      s3: true,
      trigger: {
        branch: 'master',
        all: true,
        cronSpec: '@every 10s',
        size: 'big',
        commits: 12,
      },
    });

    expect(pfsInput.getName()).toBe('images');
    expect(pfsInput.getRepo()).toBe('imagesRepo');
    expect(pfsInput.getBranch()).toBe('master');
    expect(pfsInput.getCommit()).toBe('uweioruwejrij098w0e9r809we');
    expect(pfsInput.getGlob()).toBe('/test/*');
    expect(pfsInput.getJoinOn()).toBe('test');
    expect(pfsInput.getOuterJoin()).toBe(true);
    expect(pfsInput.getGroupBy()).toBe('name');
    expect(pfsInput.getLazy()).toBe(true);
    expect(pfsInput.getEmptyFiles()).toBe(true);
    expect(pfsInput.getS3()).toBe(true);
    expect(pfsInput.getTrigger()?.getBranch()).toBe('master');
    expect(pfsInput.getTrigger()?.getAll()).toBe(true);
    expect(pfsInput.getTrigger()?.getCronSpec()).toBe('@every 10s');
    expect(pfsInput.getTrigger()?.getSize()).toBe('big');
    expect(pfsInput.getTrigger()?.getCommits()).toBe(12);
  });

  it('should create CronInput from an object', () => {
    const cronInput = cronInputFromObject({
      name: 'images',
      repo: 'imagesRepo',
      commit: 'uweioruwejrij098w0e9r809we',
      spec: '*/10 * * * *',
      overwrite: true,
      start: {
        seconds: 1614736724,
        nanos: 344218476,
      },
    });

    expect(cronInput.getName()).toBe('images');
    expect(cronInput.getRepo()).toBe('imagesRepo');
    expect(cronInput.getCommit()).toBe('uweioruwejrij098w0e9r809we');
    expect(cronInput.getSpec()).toBe('*/10 * * * *');
    expect(cronInput.getOverwrite()).toBe(true);
    expect(cronInput.getStart()?.getSeconds()).toBe(1614736724);
    expect(cronInput.getStart()?.getNanos()).toBe(344218476);
  });

  it('should create Input from an object', () => {
    const input = inputFromObject({
      pfs: {
        name: 'imagesPfs',
        repo: 'imagesRepo',
        branch: 'master',
      },
      joinList: [
        {
          pfs: {
            name: 'joinList',
            repo: 'imagesRepo',
            branch: 'master',
          },
        },
      ],
      groupList: [
        {
          pfs: {
            name: 'groupList',
            repo: 'imagesRepo',
            branch: 'master',
          },
        },
      ],
      crossList: [
        {
          pfs: {
            name: 'crossList',
            repo: 'imagesRepo',
            branch: 'master',
          },
        },
      ],
      unionList: [
        {
          pfs: {
            name: 'unionList',
            repo: 'imagesRepo',
            branch: 'master',
          },
        },
      ],

      cron: {
        name: 'imagesCron',
        repo: 'imagesRepo',
        commit: 'uweioruwejrij098w0e9r809we',
        spec: '*/10 * * * *',
        overwrite: true,
      },
    });

    expect(input.getPfs()?.getName()).toBe('imagesPfs');
    expect(input.getCron()?.getName()).toBe('imagesCron');

    expect(input.getJoinList()[0]?.getPfs()?.getName()).toBe('joinList');
    expect(input.getGroupList()[0]?.getPfs()?.getName()).toBe('groupList');
    expect(input.getCrossList()[0]?.getPfs()?.getName()).toBe('crossList');
    expect(input.getUnionList()[0]?.getPfs()?.getName()).toBe('unionList');
  });

  it('should create Service from an object', () => {
    const service = serviceFromObject({
      internalPort: 8888,
      externalPort: 30888,
      ip: '172.16.254.1',
      type: 'good',
    });

    expect(service.getInternalPort()).toBe(8888);
    expect(service.getExternalPort()).toBe(30888);
    expect(service.getIp()).toBe('172.16.254.1');
    expect(service.getType()).toBe('good');
  });

  it('should create Spout from an object', () => {
    const spout = spoutFromObject({
      service: {
        internalPort: 8888,
        externalPort: 30888,
        ip: '172.16.254.1',
        type: 'good',
      },
    });

    expect(spout.getService()?.getIp()).toBe('172.16.254.1');
  });

  it('should create ChunkSpec from an object', () => {
    const chunkSpec = chunkSpecFromObject({
      number: 123,
      sizeBytes: 23498769,
    });

    expect(chunkSpec.getNumber()).toBe(123);
    expect(chunkSpec.getSizeBytes()).toBe(23498769);
  });

  it('should create PipelineInfo from an object with defaults', () => {
    const pipelineInfo = pipelineInfoFromObject({
      pipeline: {
        name: 'testPipeline',
      },
    });

    expect(pipelineInfo.getPipeline()?.getName()).toBe('testPipeline');
    expect(pipelineInfo.getVersion()).toBe(1);
    expect(pipelineInfo.getState()).toBe(0);
    expect(pipelineInfo.getStopped()).toBe(false);
    expect(pipelineInfo.getLastJobState()).toBe(0);
    expect(pipelineInfo.getDetails()?.getTransform()).toBe(undefined);
    expect(pipelineInfo.getDetails()?.getTfJob()).toBe(undefined);
    expect(pipelineInfo.getDetails()?.getParallelismSpec()).toBe(undefined);
    expect(pipelineInfo.getDetails()?.getEgress()).toBe(undefined);
    expect(pipelineInfo.getDetails()?.getCreatedAt()).toBe(undefined);
    expect(pipelineInfo.getDetails()?.getRecentError()).toBe('');
    expect(pipelineInfo.getDetails()?.getWorkersRequested()).toBe(0);
    expect(pipelineInfo.getDetails()?.getWorkersAvailable()).toBe(0);
    expect(pipelineInfo.getDetails()?.getOutputBranch()).toBe('master');
    expect(pipelineInfo.getDetails()?.getResourceRequests()).toBe(undefined);
    expect(pipelineInfo.getDetails()?.getResourceLimits()).toBe(undefined);
    expect(pipelineInfo.getDetails()?.getSidecarResourceLimits()).toBe(
      undefined,
    );
    expect(pipelineInfo.getDetails()?.getInput()).toBe(undefined);
    expect(pipelineInfo.getDetails()?.getDescription()).toBe('');
    expect(pipelineInfo.getDetails()?.getSalt()).toBe('');
    expect(pipelineInfo.getDetails()?.getReason()).toBe('');
    expect(pipelineInfo.getDetails()?.getService()).toBe(undefined);
    expect(pipelineInfo.getDetails()?.getSpout()).toBe(undefined);
    expect(pipelineInfo.getDetails()?.getDatumSetSpec()).toBe(undefined);
    expect(pipelineInfo.getDetails()?.getDatumTimeout()).toBe(undefined);
    expect(pipelineInfo.getDetails()?.getJobTimeout()).toBe(undefined);
    expect(pipelineInfo.getDetails()?.getDatumTries()).toBe(0);
    expect(pipelineInfo.getDetails()?.getPodSpec()).toBe('');
    expect(pipelineInfo.getDetails()?.getPodPatch()).toBe('');
    expect(pipelineInfo.getDetails()?.getS3Out()).toBe(false);
  });

  it('should create PipelineInfo from an object with defaults', () => {
    const pipelineInfo = pipelineInfoFromObject({
      pipeline: {
        name: 'testPipeline',
      },
      version: 4,
      transform: {
        image: 'pachyderm/opencv',
        cmdList: ['python3', '/edges.py'],
      },
      tfJob: {tfJob: 'example-job'},
      parallelismSpec: {
        constant: 1,
      },
      egress: {
        url: 's3://bucket/dir',
      },
      createdAt: {
        seconds: 1614736724,
        nanos: 344218476,
      },
      state: 3,
      stopped: true,
      recentError: 'err',
      workersRequested: 23,
      workersAvailable: 2,
      lastJobState: 3,
      outputBranch: 'testBranch',
      resourceRequests: {
        cpu: 8,
        memory: '12mb',
        gpu: {type: 'good', number: 3},
        disk: '2t',
      },
      resourceLimits: {
        cpu: 5,
        memory: '12mb',
        gpu: {type: 'good', number: 3},
        disk: '2t',
      },
      sidecarResourceLimits: {
        cpu: 12,
        memory: '12mb',
        gpu: {type: 'good', number: 3},
        disk: '2t',
      },
      input: {
        pfs: {
          name: 'imagesPfs',
          repo: 'imagesRepo',
          branch: 'master',
        },
      },
      description: 'yo yo yo!',
      salt: 'd5631d7df40d4b1195bc46f1f146d6a5',
      reason: 'because',
      service: {
        internalPort: 8888,
        externalPort: 30888,
        ip: '172.16.254.1',
        type: 'good',
      },
      spout: {
        service: {
          internalPort: 8888,
          externalPort: 30888,
          ip: '172.16.254.1',
          type: 'good',
        },
      },
      chunkSpec: {
        number: 123,
        sizeBytes: 23498769,
      },
      datumTimeout: {
        seconds: 23424,
        nanos: 254345,
      },
      jobTimeout: {
        seconds: 564645,
        nanos: 867867,
      },
      specCommit: {
        branch: {name: '', repo: {name: '__spec__'}},
        id: '4af40d34a0384f23a5b98d3bd7eaece1',
      },
      datumTries: 12,
      podSpec: 'podSpec',
      podPatch: 'podPatch',
      s3Out: true,
    });

    expect(pipelineInfo.getPipeline()?.getName()).toBe('testPipeline');
    expect(pipelineInfo.getVersion()).toBe(4);
    expect(pipelineInfo.getState()).toBe(3);
    expect(pipelineInfo.getStopped()).toBe(true);
    expect(pipelineInfo.getLastJobState()).toBe(3);
    expect(pipelineInfo.getDetails()?.getTransform()?.getImage()).toBe(
      'pachyderm/opencv',
    );
    expect(pipelineInfo.getDetails()?.getTfJob()?.getTfJob()).toBe(
      'example-job',
    );
    expect(pipelineInfo.getDetails()?.getEgress()?.getUrl()).toBe(
      's3://bucket/dir',
    );
    expect(pipelineInfo.getDetails()?.getCreatedAt()?.getSeconds()).toBe(
      1614736724,
    );
    expect(pipelineInfo.getDetails()?.getRecentError()).toBe('err');
    expect(pipelineInfo.getDetails()?.getWorkersRequested()).toBe(23);
    expect(pipelineInfo.getDetails()?.getWorkersAvailable()).toBe(2);
    expect(pipelineInfo.getDetails()?.getOutputBranch()).toBe('testBranch');
    expect(pipelineInfo.getDetails()?.getResourceRequests()?.getCpu()).toBe(8);
    expect(pipelineInfo.getDetails()?.getResourceLimits()?.getCpu()).toBe(5);
    expect(
      pipelineInfo.getDetails()?.getSidecarResourceLimits()?.getCpu(),
    ).toBe(12);
    expect(pipelineInfo.getDetails()?.getInput()?.getPfs()?.getName()).toBe(
      'imagesPfs',
    );
    expect(pipelineInfo.getDetails()?.getDescription()).toBe('yo yo yo!');
    expect(pipelineInfo.getDetails()?.getSalt()).toBe(
      'd5631d7df40d4b1195bc46f1f146d6a5',
    );
    expect(pipelineInfo.getReason()).toBe('because');
    expect(pipelineInfo.getDetails()?.getService()?.getIp()).toBe(
      '172.16.254.1',
    );
    expect(pipelineInfo.getDetails()?.getDatumSetSpec()?.getNumber()).toBe(123);
    expect(pipelineInfo.getDetails()?.getDatumTimeout()?.getSeconds()).toBe(
      23424,
    );
    expect(pipelineInfo.getDetails()?.getJobTimeout()?.getSeconds()).toBe(
      564645,
    );
    expect(pipelineInfo.getDetails()?.getDatumTries()).toBe(12);
    expect(pipelineInfo.getDetails()?.getPodSpec()).toBe('podSpec');
    expect(pipelineInfo.getDetails()?.getPodPatch()).toBe('podPatch');
    expect(pipelineInfo.getDetails()?.getS3Out()).toBe(true);
  });

  it('should create PipelineInfos from an object', () => {
    const pipelineInfos = pipelineInfosFromObject({
      pipelineInfoList: [
        {
          pipeline: {
            name: 'pipeline_one',
          },
        },
        {
          pipeline: {
            name: 'pipeline_two',
          },
        },
      ],
    });

    expect(
      pipelineInfos.getPipelineInfoList()[0]?.getPipeline()?.getName(),
    ).toBe('pipeline_one');
    expect(
      pipelineInfos.getPipelineInfoList()[1]?.getPipeline()?.getName(),
    ).toBe('pipeline_two');
  });
});

it('should create PipelineJob from an object', () => {
  const pipelineJob = jobFromObject({id: '23efw4ef098few0'});

  expect(pipelineJob.getId()).toBe('23efw4ef098few0');
});

it('should create JobInfo from an object', () => {
  const pipelineJob = jobInfoFromObject({
    state: 1,
    job: {id: '1', pipeline: {name: 'montage'}},
    createdAt: {
      seconds: 564645,
      nanos: 0,
    },
  });

  expect(pipelineJob.getState()).toBe(1);
  expect(pipelineJob.getStarted()?.getSeconds()).toBe(564645);
  expect(pipelineJob.getJob()?.getId()).toBe('1');
});

it('should create GetLogsRequestObject from a pipeline request', () => {
  const getLogsRequest = getLogsRequestFromObject({
    pipelineName: 'PipelineName',
    since: 564645,
    follow: true,
  });

  expect(getLogsRequest.getPipeline()?.getName()).toBe('PipelineName');
  expect(getLogsRequest.getJob()).toBe(undefined);
  expect(getLogsRequest.getSince()?.getSeconds()).toBe(564645);
  expect(getLogsRequest.getFollow()).toBe(true);
});

it('should create GetLogsRequestObject from a job request', () => {
  const getLogsRequest = getLogsRequestFromObject({
    pipelineName: 'PipelineName',
    jobId: '2222222',
    since: 564645,
    follow: true,
  });

  expect(getLogsRequest.getPipeline()).toBe(undefined);
  expect(getLogsRequest.getJob()?.getId()).toBe('2222222');
  expect(getLogsRequest.getJob()?.getPipeline()?.getName()).toBe(
    'PipelineName',
  );
  expect(getLogsRequest.getSince()?.getSeconds()).toBe(564645);
  expect(getLogsRequest.getFollow()).toBe(true);
});
