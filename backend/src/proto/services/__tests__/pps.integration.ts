import crypto from 'crypto';

import {status} from '@grpc/grpc-js';

import apiClientRequestWrapper from '../../client';
import {
  DatumState,
  Input,
  PFSInput,
  PipelineState,
  Transform,
} from '../../proto/pps/pps_pb';

jest.setTimeout(30_000);

describe('services/pps', () => {
  afterAll(async () => {
    const pachClient = apiClientRequestWrapper();
    const pps = pachClient.pps;
    const pfs = pachClient.pfs;
    await pps.deleteAll();
    await pfs.deleteAll();
  });

  const createSandBox = async (name: string) => {
    const pachClient = apiClientRequestWrapper();
    const pps = pachClient.pps;
    const pfs = pachClient.pfs;
    await pps.deleteAll();
    await pfs.deleteAll();
    const inputRepoName = crypto.randomBytes(5).toString('hex');
    await pfs.createRepo({projectId: 'default', repo: {name: inputRepoName}});
    const transform = new Transform()
      .setCmdList(['sh'])
      .setImage('alpine')
      .setStdinList([`cp /pfs/${inputRepoName}/* /pfs/out/`]);
    const input = new Input();
    const pfsInput = new PFSInput().setGlob('/*').setRepo(inputRepoName);
    input.setPfs(pfsInput);

    await pps.createPipeline({
      pipeline: {name},
      transform: transform.toObject(),
      input: input.toObject(),
      dryRun: false,
    });

    const input2 = new Input();
    const pfsInput2 = new PFSInput().setGlob('/*').setRepo(name);
    input2.setPfs(pfsInput2);

    await pps.createPipeline({
      pipeline: {name: `${name}-2`},
      transform: transform.toObject(),
      input: input2.toObject(),
      dryRun: false,
    });

    return {pachClient, inputRepoName};
  };

  // These are long running tests if there is data in pachyderm. So run these first!
  describe('listDatums + inspectDatum', () => {
    it('should inspect a datum for a pipeline job /', async () => {
      const {pachClient, inputRepoName} = await createSandBox('listDatums');
      const commit = await pachClient.pfs.startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: inputRepoName}},
      });

      const fileClient = await pachClient.pfs.modifyFile();

      await fileClient
        .setCommit(commit)
        .putFileFromBytes('dummyData.csv', Buffer.from('a,b,c'))
        .end();

      await pachClient.pfs.finishCommit({projectId: 'default', commit});

      // should inspect a datum for a pipeline job
      const jobs = await pachClient.pps.listJobs({projectId: 'default'});
      const jobId = jobs[0]?.job?.id;
      expect(jobId).toBeDefined();

      await pachClient.pps.inspectJob({
        id: jobId || '',
        pipelineName: 'listDatums',
        wait: true,
        projectId: '',
      });

      const datums = await pachClient.pps.listDatums({
        projectId: 'default',
        jobId: jobId || '',
        pipelineName: 'listDatums',
      });

      expect(datums).toHaveLength(1);

      const datum = await pachClient.pps.inspectDatum({
        projectId: 'default',
        id: datums[0]?.datum?.id || '',
        jobId: jobId || '',
        pipelineName: 'listDatums',
      });

      const datumObject = datum.toObject();

      expect(datumObject.state).toEqual(DatumState.SUCCESS);
      expect(datumObject.dataList[0]?.file?.path).toBe('/dummyData.csv');
      expect(datumObject.dataList[0]?.sizeBytes).toBe(5);

      // should list datums for a pipeline job
      expect(jobId).toBeDefined();

      await pachClient.pps.inspectJob({
        id: jobId || '',
        pipelineName: 'listDatums',
        wait: true,
        projectId: '',
      });

      expect(datums).toHaveLength(1);
      expect(datums[0].state).toEqual(DatumState.SUCCESS);
      expect(datums[0].dataList[0]?.file?.path).toBe('/dummyData.csv');
      expect(datums[0].dataList[0]?.sizeBytes).toBe(5);

      // should filter datum list
      expect(jobId).toBeDefined();

      await pachClient.pps.inspectJob({
        id: jobId || '',
        pipelineName: 'listDatums',
        wait: true,
        projectId: '',
      });

      const filteredDatums1 = await pachClient.pps.listDatums({
        projectId: 'default',
        jobId: jobId || '',
        pipelineName: 'listDatums',
        filter: [DatumState.FAILED],
      });

      expect(filteredDatums1).toHaveLength(0);

      const filteredDatums2 = await pachClient.pps.listDatums({
        projectId: 'default',
        jobId: jobId || '',
        pipelineName: 'listDatums',
        filter: [DatumState.FAILED, DatumState.SUCCESS],
      });

      expect(filteredDatums2).toHaveLength(1);
    }, 60_000);
  });

  describe('listPipeline', () => {
    it('should return a list of pipelines in the pachyderm cluster', async () => {
      const {pachClient} = await createSandBox('listPipeline');

      const pipelines = await pachClient.pps.listPipeline({projectIds: []});

      expect(pipelines).toHaveLength(2);
      expect(pipelines[1].pipeline?.name).toBe('listPipeline');
    });
  });

  describe('listJobs', () => {
    it('should return a list of jobs', async () => {
      const {pachClient} = await createSandBox('listJobs');

      const jobs = await pachClient.pps.listJobs({projectId: 'default'});
      expect(jobs).toHaveLength(2);
    });

    it('should return a list of jobs given a pipeline id', async () => {
      const {pachClient} = await createSandBox('listJobs');

      const jobs = await pachClient.pps.listJobs({
        projectId: 'default',
        pipelineId: 'listJobs',
      });
      expect(jobs).toHaveLength(1);
    });

    it('should return a list of jobs given a jq filter', async () => {
      const {pachClient} = await createSandBox('listJobs');

      const jobs = await pachClient.pps.listJobs({
        projectId: 'default',
        jqFilter: `select(.job.pipeline.name == "listJobs")`,
      });
      expect(jobs).toHaveLength(1);
    });
  });

  describe('inspectPipeline', () => {
    it('should return details about the pipeline', async () => {
      const {pachClient, inputRepoName} = await createSandBox(
        'inspectPipeline',
      );

      const pipeline = await pachClient.pps.inspectPipeline({
        pipelineId: 'inspectPipeline',
        projectId: '',
      });

      expect(pipeline.pipeline?.name).toBe('inspectPipeline');
      expect(pipeline.state).toEqual(PipelineState.PIPELINE_STARTING);
      expect(pipeline.details?.input?.pfs?.repo).toEqual(inputRepoName);
    });
  });

  describe('listJobSets', () => {
    it('should return a list of job sets', async () => {
      const {pachClient} = await createSandBox('listJobSets');

      const jobSets = await pachClient.pps.listJobSets();
      expect(jobSets).toHaveLength(2);
    });
  });

  describe('inspectJob', () => {
    it('should return details about the specified job', async () => {
      const {pachClient} = await createSandBox('inspectJob');
      const jobs = await pachClient.pps.listJobs({projectId: 'default'});
      const job = await pachClient.pps.inspectJob({
        id: jobs[1].job?.id || '',
        pipelineName: 'inspectJob',
        projectId: '',
      });

      expect(job).toMatchObject(jobs[1]);
    });
  });

  describe('getLogs', () => {
    it('should return logs about the cluster', async () => {
      const {pachClient} = await createSandBox('getLogs');
      const logs = await pachClient.pps.getLogs({projectId: 'default'});
      logs.map((log) =>
        expect(log).toEqual(
          expect.objectContaining({
            pipelineName: expect.any(String),
            jobId: expect.any(String),
            workerId: expect.any(String),
            datumId: expect.any(String),
            master: expect.any(Boolean),
            dataList: expect.any(Array),
            user: expect.any(Boolean),
            message: expect.any(String),
          }),
        ),
      );
    });
  });

  describe('createPipeline', () => {
    it('should create a pipeline', async () => {
      const {pachClient, inputRepoName} = await createSandBox('createPipeline');
      const pipelines = await pachClient.pps.listPipeline({projectIds: []});
      expect(pipelines).toHaveLength(2);

      const transform = new Transform()
        .setCmdList(['sh'])
        .setImage('alpine')
        .setStdinList([`cp /pfs/${inputRepoName}/*.dat /pfs/out/`]);
      const input = new Input();
      const pfsInput = new PFSInput().setGlob('/*').setRepo(inputRepoName);
      input.setPfs(pfsInput);

      await pachClient.pps.createPipeline({
        pipeline: {name: 'createPipeline2'},
        transform: transform.toObject(),
        input: input.toObject(),
        dryRun: false,
      });
      const updatedPipelines = await pachClient.pps.listPipeline({
        projectIds: [],
      });
      expect(updatedPipelines).toHaveLength(3);
    });
  });

  describe('createPipelineV2', () => {
    it('should create a pipeline', async () => {
      const {pachClient, inputRepoName} = await createSandBox(
        'createPipelineV2',
      );
      const pipelines = await pachClient.pps.listPipeline({projectIds: []});
      expect(pipelines).toHaveLength(2);

      const json = `{"pipeline":{"project":{"name":"default"}, "name":"test"}, "input": {"pfs": {"project": "default","name": "${inputRepoName}","repo": "${inputRepoName}","repoType": "user","branch": "master","glob": "/*"}},"transform": {"image": "alpine","cmd": ["sh"],"stdin": ["cp /pfs/${inputRepoName}/*.dat /pfs/out/"]}, "resourceRequests":{"cpu":1}}`;

      await pachClient.pps.setClusterDefaults({
        clusterDefaultsJson: `{"create_pipeline_request": {}}`,
      });

      const resp = await pachClient.pps.createPipelineV2({
        createPipelineRequestJson: json,
        reprocess: false,
        dryRun: false,
        update: false,
      });

      expect(JSON.parse(resp.effectiveCreatePipelineRequestJson)).toEqual(
        JSON.parse(
          `{"pipeline":{"project":{"name":"default"}, "name":"test"}, "transform":{"image":"alpine", "cmd":["sh"], "stdin":["cp /pfs/${inputRepoName}/*.dat /pfs/out/"]}, "resourceRequests":{"cpu":1}, "input":{"pfs":{"project":"default", "name":"${inputRepoName}", "repo":"${inputRepoName}", "repoType":"user", "branch":"master", "glob":"/*"}}}`,
        ),
      );

      expect(await pachClient.pps.listPipeline({projectIds: []})).toHaveLength(
        3,
      );
    });

    it('updates an existing pipeline', async () => {
      const {pachClient, inputRepoName} = await createSandBox(
        'createPipelineV2',
      );
      const pipelines = await pachClient.pps.listPipeline({projectIds: []});
      expect(pipelines).toHaveLength(2);

      const json = `{"pipeline":{"project":{"name":"default"}, "name":"test"}, "input": {"pfs": {"project": "default","name": "${inputRepoName}","repo": "${inputRepoName}","repoType": "user","branch": "master","glob": "/*"}},"transform": {"image": "alpine","cmd": ["sh"],"stdin": ["cp /pfs/${inputRepoName}/*.dat /pfs/out/"]}, "resourceRequests":{"cpu":1}}`;

      await pachClient.pps.setClusterDefaults({
        clusterDefaultsJson: `{"create_pipeline_request": {}}`,
      });

      await pachClient.pps.createPipelineV2({
        createPipelineRequestJson: json,
        reprocess: false,
        dryRun: false,
        update: false,
      });

      const resp = await pachClient.pps.createPipelineV2({
        createPipelineRequestJson: json,
        reprocess: false,
        dryRun: false,
        update: true,
      });

      expect(JSON.parse(resp.effectiveCreatePipelineRequestJson)).toEqual(
        JSON.parse(
          `{"update": true, "pipeline":{"project":{"name":"default"}, "name":"test"}, "transform":{"image":"alpine", "cmd":["sh"], "stdin":["cp /pfs/${inputRepoName}/*.dat /pfs/out/"]}, "resourceRequests":{"cpu":1}, "input":{"pfs":{"project":"default", "name":"${inputRepoName}", "repo":"${inputRepoName}", "repoType":"user", "branch":"master", "glob":"/*"}}}`,
        ),
      );

      expect(await pachClient.pps.listPipeline({projectIds: []})).toHaveLength(
        3,
      );
    });

    it('trying to update without update field returns errors', async () => {
      const {pachClient, inputRepoName} = await createSandBox(
        'createPipelineV2',
      );
      const pipelines = await pachClient.pps.listPipeline({projectIds: []});
      expect(pipelines).toHaveLength(2);

      const json = `{"pipeline":{"project":{"name":"default"}, "name":"test"}, "input": {"pfs": {"project": "default","name": "${inputRepoName}","repo": "${inputRepoName}","repoType": "user","branch": "master","glob": "/*"}},"transform": {"image": "alpine","cmd": ["sh"],"stdin": ["cp /pfs/${inputRepoName}/*.dat /pfs/out/"]}, "resourceRequests":{"cpu":1}}`;

      await pachClient.pps.setClusterDefaults({
        clusterDefaultsJson: `{"create_pipeline_request": {}}`,
      });

      await pachClient.pps.createPipelineV2({
        createPipelineRequestJson: json,
        reprocess: false,
        dryRun: false,
        update: false,
      });

      await expect(
        pachClient.pps.createPipelineV2({
          createPipelineRequestJson: json,
          reprocess: false,
          dryRun: false,
          update: false,
        }),
      ).rejects.toHaveProperty('code', status.ALREADY_EXISTS);
    });

    it('dryrun does not create a pipeline', async () => {
      const {pachClient, inputRepoName} = await createSandBox(
        'createPipelineV2',
      );
      const pipelines = await pachClient.pps.listPipeline({projectIds: []});
      expect(pipelines).toHaveLength(2);

      const json = `{"pipeline":{"project":{"name":"default"}, "name":"test"}, "input": {"pfs": {"project": "default","name": "${inputRepoName}","repo": "${inputRepoName}","repoType": "user","branch": "master","glob": "/*"}},"transform": {"image": "alpine","cmd": ["sh"],"stdin": ["cp /pfs/${inputRepoName}/*.dat /pfs/out/"]}, "resourceRequests":{"cpu":1}}`;

      await pachClient.pps.setClusterDefaults({
        clusterDefaultsJson: `{"create_pipeline_request": {}}`,
      });

      const resp = await pachClient.pps.createPipelineV2({
        createPipelineRequestJson: json,
        reprocess: false,
        dryRun: true,
        update: false,
      });

      expect(JSON.parse(resp.effectiveCreatePipelineRequestJson)).toEqual(
        JSON.parse(
          `{"pipeline":{"project":{"name":"default"}, "name":"test"}, "transform":{"image":"alpine", "cmd":["sh"], "stdin":["cp /pfs/${inputRepoName}/*.dat /pfs/out/"]}, "resourceRequests":{"cpu":1}, "input":{"pfs":{"project":"default", "name":"${inputRepoName}", "repo":"${inputRepoName}", "repoType":"user", "branch":"master", "glob":"/*"}}}`,
        ),
      );

      expect(await pachClient.pps.listPipeline({projectIds: []})).toHaveLength(
        2,
      );
    });

    it('bad input spec or no arguments gives an error', async () => {
      const client = apiClientRequestWrapper();
      const pps = client.pps;
      const pfs = client.pfs;
      await pps.deleteAll();
      await pfs.deleteAll();

      expect(await client.pps.listPipeline({projectIds: []})).toHaveLength(0);

      await expect(
        pps.createPipelineV2({
          createPipelineRequestJson: `{"p": {"p": "p"}, "p": "p", "p": {}}`,
        }),
      ).rejects.toHaveProperty('code', status.INVALID_ARGUMENT);

      await expect(pps.createPipelineV2({})).rejects.toHaveProperty(
        'code',
        status.INVALID_ARGUMENT,
      );
    });
  });

  describe('deletePipeline', () => {
    it('should delete a pipeline', async () => {
      const {pachClient} = await createSandBox('deletePipeline');
      const pipelines = await pachClient.pps.listPipeline({projectIds: []});
      expect(pipelines).toHaveLength(2);

      await pachClient.pps.deletePipeline({
        projectId: 'default',
        pipeline: {name: 'deletePipeline-2'},
        force: true,
      });

      const updatedPipelines = await pachClient.pps.listPipeline({
        projectIds: [],
      });
      expect(updatedPipelines).toHaveLength(1);
    });
  });

  describe('deletePipelines', () => {
    it('should delete pipelines', async () => {
      const {pachClient} = await createSandBox('deletePipelines');
      const pipelines = await pachClient.pps.listPipeline({projectIds: []});
      expect(pipelines).toHaveLength(2);
      await pachClient.pps.deletePipelines({
        projectIds: ['default'],
      });
      const updatedPipelines = await pachClient.pps.listPipeline({
        projectIds: [],
      });
      expect(updatedPipelines).toHaveLength(0);
    });
  });

  describe('getClusterDefaults', () => {
    it('should return cluster defaults', async () => {
      const client = apiClientRequestWrapper();
      const pps = client.pps;
      const pfs = client.pfs;
      await pps.deleteAll();
      await pfs.deleteAll();

      await pps.setClusterDefaults({
        clusterDefaultsJson: `{"createPipelineRequest": {"resourceRequests":{"cpu":1, "memory":"256Mi", "disk":"1Gi"}, "sidecarResourceRequests":{"cpu":1, "memory":"256Mi", "disk":"1Gi"}}}`,
      });

      const resp = await pps.getClusterDefaults();

      expect(JSON.parse(resp.clusterDefaultsJson)).toEqual(
        expect.objectContaining(
          JSON.parse(
            '{"createPipelineRequest":{"resourceRequests":{"cpu":1, "memory":"256Mi", "disk":"1Gi"}, "sidecarResourceRequests":{"cpu":1, "memory":"256Mi", "disk":"1Gi"}}}',
          ),
        ),
      );
    });
  });

  describe('setClusterDefaults', () => {
    it('should set empty cluster defaults', async () => {
      const client = apiClientRequestWrapper();
      const pps = client.pps;
      const pfs = client.pfs;
      await pps.deleteAll();
      await pfs.deleteAll();

      const resp = await pps.setClusterDefaults({
        clusterDefaultsJson: '{}',
        dryRun: false,
        regenerate: false,
        reprocess: true,
      });

      expect(resp).toEqual(
        expect.objectContaining({
          affectedPipelinesList: [],
        }),
      );

      expect(await pps.getClusterDefaults()).toEqual(
        expect.objectContaining({
          clusterDefaultsJson: '{}',
        }),
      );
    });
    it('should set empty create_pipeline_request cluster defaults', async () => {
      const client = apiClientRequestWrapper();
      const pps = client.pps;
      const pfs = client.pfs;
      await pps.deleteAll();
      await pfs.deleteAll();

      const resp = await pps.setClusterDefaults({
        clusterDefaultsJson: '{"create_pipeline_request": {}}',
        dryRun: false,
        regenerate: false,
        reprocess: true,
      });

      expect(resp).toEqual(
        expect.objectContaining({
          affectedPipelinesList: [],
        }),
      );

      expect(await pps.getClusterDefaults()).toEqual(
        expect.objectContaining({
          clusterDefaultsJson: '{"create_pipeline_request": {}}',
        }),
      );
    });
    it('should set cluster defaults', async () => {
      const client = apiClientRequestWrapper();
      const pps = client.pps;
      const pfs = client.pfs;
      await pps.deleteAll();
      await pfs.deleteAll();

      const resp = await pps.setClusterDefaults({
        clusterDefaultsJson:
          '{"create_pipeline_request": {"pipeline": {"project": {"name": "foobar"}}}}',
        dryRun: false,
        regenerate: false,
        reprocess: true,
      });

      expect(resp).toEqual(
        expect.objectContaining({
          affectedPipelinesList: [],
        }),
      );

      expect(await pps.getClusterDefaults()).toEqual(
        expect.objectContaining({
          clusterDefaultsJson:
            '{"create_pipeline_request": {"pipeline": {"project": {"name": "foobar"}}}}',
        }),
      );
    });
    it('dry run does not affect the defaults', async () => {
      const client = apiClientRequestWrapper();
      const pps = client.pps;
      const pfs = client.pfs;
      await pps.deleteAll();
      await pfs.deleteAll();

      await pps.setClusterDefaults({
        clusterDefaultsJson: '{}',
        dryRun: false,
        regenerate: false,
        reprocess: true,
      });

      const resp = await pps.setClusterDefaults({
        clusterDefaultsJson:
          '{"create_pipeline_request": {"pipeline": {"project": {"name": "foobar"}}}}',
        dryRun: true,
        regenerate: false,
        reprocess: true,
      });

      expect(resp).toEqual(
        expect.objectContaining({
          affectedPipelinesList: [],
        }),
      );

      expect(await pps.getClusterDefaults()).toEqual(
        expect.objectContaining({
          clusterDefaultsJson: '{}',
        }),
      );
    });
    it('regenerate returns correct pipelines', async () => {
      const {pachClient, inputRepoName} = await createSandBox(
        'createPipelineV2',
      );
      const pipelines = await pachClient.pps.listPipeline({projectIds: []});
      expect(pipelines).toHaveLength(2);

      const json = `{"pipeline":{"project":{"name":"default"}, "name":"test"}, "input": {"pfs": {"project": "default","name": "${inputRepoName}","repo": "${inputRepoName}","repoType": "user","branch": "master","glob": "/*"}},"transform": {"image": "alpine","cmd": ["sh"],"stdin": ["cp /pfs/${inputRepoName}/*.dat /pfs/out/"]}}`;

      await pachClient.pps.createPipelineV2({
        createPipelineRequestJson: json,
      });

      const resp = await pachClient.pps.setClusterDefaults({
        clusterDefaultsJson:
          '{"create_pipeline_request": {"resourceRequests":{"cpu":1}}}',
        dryRun: true,
        regenerate: true,
        reprocess: false,
      });

      resp.affectedPipelinesList.sort((a, b) => {
        if (a.name < b.name) {
          return -1;
        }
        if (a.name > b.name) {
          return 1;
        }
        return 0;
      });

      expect(resp).toEqual(
        expect.objectContaining({
          affectedPipelinesList: [
            {
              name: 'createPipelineV2',
              project: {
                name: 'default',
              },
            },
            {
              name: 'createPipelineV2-2',
              project: {
                name: 'default',
              },
            },
            {
              name: 'test',
              project: {
                name: 'default',
              },
            },
          ],
        }),
      );
    });
  });

  describe('clusterDefaults and createPipelineV2', () => {
    it('createPipelineV2 can add values', async () => {
      const {pachClient, inputRepoName} = await createSandBox(
        'createPipelineV2',
      );
      const pipelines = await pachClient.pps.listPipeline({projectIds: []});
      expect(pipelines).toHaveLength(2);

      await pachClient.pps.setClusterDefaults({
        clusterDefaultsJson: '{}',
        dryRun: false,
        regenerate: false,
        reprocess: false,
      });

      const json = `{"pipeline":{"project":{"name":"default"}, "name":"test"}, "input": {"pfs": {"project": "default","name": "${inputRepoName}","repo": "${inputRepoName}","repoType": "user","branch": "master","glob": "/*"}},"transform": {"image": "alpine","cmd": ["sh"],"stdin": ["cp /pfs/${inputRepoName}/*.dat /pfs/out/"]}}`;

      const resp = await pachClient.pps.createPipelineV2({
        createPipelineRequestJson: json,
        dryRun: true,
      });

      expect(JSON.parse(resp.effectiveCreatePipelineRequestJson)).toEqual(
        expect.objectContaining(JSON.parse(json)),
      );
    });
    it('createPipelineV2 can overwrite a cluster default value', async () => {
      const {pachClient, inputRepoName} = await createSandBox(
        'createPipelineV2',
      );
      const pipelines = await pachClient.pps.listPipeline({projectIds: []});
      expect(pipelines).toHaveLength(2);

      await pachClient.pps.setClusterDefaults({
        clusterDefaultsJson:
          '{"create_pipeline_request": {"resourceRequests":{"cpu":1}}}',
        dryRun: false,
        regenerate: false,
        reprocess: false,
      });

      const json = `{"pipeline":{"project":{"name":"default"}, "name":"test"}, "input": {"pfs": {"project": "default","name": "${inputRepoName}","repo": "${inputRepoName}","repoType": "user","branch": "master","glob": "/*"}},"transform": {"image": "alpine","cmd": ["sh"],"stdin": ["cp /pfs/${inputRepoName}/*.dat /pfs/out/"]}, "resourceRequests":{"cpu": 2}}`;

      const resp = await pachClient.pps.createPipelineV2({
        createPipelineRequestJson: json,
        dryRun: true,
      });

      expect(JSON.parse(resp.effectiveCreatePipelineRequestJson)).toEqual(
        expect.objectContaining(JSON.parse(json)),
      );
    });
    it('createPipelineV2 can unset a cluster default value', async () => {
      const {pachClient, inputRepoName} = await createSandBox(
        'createPipelineV2',
      );
      const pipelines = await pachClient.pps.listPipeline({projectIds: []});
      expect(pipelines).toHaveLength(2);

      await pachClient.pps.setClusterDefaults({
        clusterDefaultsJson:
          '{"create_pipeline_request": {"resourceRequests":{"cpu":1}}}',
        dryRun: false,
        regenerate: false,
        reprocess: false,
      });

      const json = `{"pipeline":{"project":{"name":"default"}, "name":"test"}, "input": {"pfs": {"project": "default","name": "${inputRepoName}","repo": "${inputRepoName}","repoType": "user","branch": "master","glob": "/*"}},"transform": {"image": "alpine","cmd": ["sh"],"stdin": ["cp /pfs/${inputRepoName}/*.dat /pfs/out/"]}, "resourceRequests":{"cpu": null}}`;

      const resp = await pachClient.pps.createPipelineV2({
        createPipelineRequestJson: json,
        dryRun: true,
      });

      expect(JSON.parse(resp.effectiveCreatePipelineRequestJson)).toEqual(
        expect.objectContaining(
          JSON.parse(
            `{"pipeline":{"project":{"name":"default"}, "name":"test"}, "input": {"pfs": {"project": "default","name": "${inputRepoName}","repo": "${inputRepoName}","repoType": "user","branch": "master","glob": "/*"}},"transform": {"image": "alpine","cmd": ["sh"],"stdin": ["cp /pfs/${inputRepoName}/*.dat /pfs/out/"]}, "resourceRequests":{}}`,
          ),
        ),
      );
    });
  });
});
