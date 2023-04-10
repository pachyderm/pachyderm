import crypto from 'crypto';

import client from '../../client';
import {
  DatumState,
  Input,
  PFSInput,
  PipelineState,
  Transform,
} from '../../proto/pps/pps_pb';

describe('services/pps', () => {
  afterAll(async () => {
    const pachClient = client({ssl: false, pachdAddress: 'localhost:30650'});
    const pps = pachClient.pps();
    const pfs = pachClient.pfs();
    await pps.deleteAll();
    await pfs.deleteAll();
  });

  const createSandBox = async (name: string) => {
    const pachClient = client({ssl: false, pachdAddress: 'localhost:30650'});
    const pps = pachClient.pps();
    const pfs = pachClient.pfs();
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
    });

    const input2 = new Input();
    const pfsInput2 = new PFSInput().setGlob('/*').setRepo(name);
    input2.setPfs(pfsInput2);

    await pps.createPipeline({
      pipeline: {name: `${name}-2`},
      transform: transform.toObject(),
      input: input2.toObject(),
    });

    return {pachClient, inputRepoName};
  };

  describe('listPipeline', () => {
    it('should return a list of pipelines in the pachyderm cluster', async () => {
      const {pachClient} = await createSandBox('listPipeline');

      const pipelines = await pachClient.pps().listPipeline({projectIds: []});

      expect(pipelines).toHaveLength(2);
      expect(pipelines[1].pipeline?.name).toBe('listPipeline');
    });
  });

  describe('listJobs', () => {
    it('should return a list of jobs', async () => {
      const {pachClient} = await createSandBox('listJobs');

      const jobs = await pachClient.pps().listJobs({projectId: 'default'});
      expect(jobs).toHaveLength(2);
    });

    it('should return a list of jobs given a pipeline id', async () => {
      const {pachClient} = await createSandBox('listJobs');

      const jobs = await pachClient
        .pps()
        .listJobs({projectId: 'default', pipelineId: 'listJobs'});
      expect(jobs).toHaveLength(1);
    });

    it('should return a list of jobs given a jq filter', async () => {
      const {pachClient} = await createSandBox('listJobs');

      const jobs = await pachClient.pps().listJobs({
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

      const pipeline = await pachClient
        .pps()
        .inspectPipeline({pipelineId: 'inspectPipeline', projectId: ''});

      expect(pipeline.pipeline?.name).toBe('inspectPipeline');
      expect(pipeline.state).toEqual(PipelineState.PIPELINE_STARTING);
      expect(pipeline.details?.input?.pfs?.repo).toEqual(inputRepoName);
    });
  });

  describe('listJobSets', () => {
    it('should return a list of job sets', async () => {
      const {pachClient} = await createSandBox('listJobSets');

      const jobSets = await pachClient.pps().listJobSets();
      expect(jobSets).toHaveLength(2);
    });
  });

  describe('inspectJob', () => {
    it('should return details about the specified job', async () => {
      const {pachClient} = await createSandBox('inspectJob');
      const jobs = await pachClient.pps().listJobs({projectId: 'default'});
      const job = await pachClient.pps().inspectJob({
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
      const logs = await pachClient.pps().getLogs({projectId: 'default'});
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
      const pipelines = await pachClient.pps().listPipeline({projectIds: []});
      expect(pipelines).toHaveLength(2);

      const transform = new Transform()
        .setCmdList(['sh'])
        .setImage('alpine')
        .setStdinList([`cp /pfs/${inputRepoName}/*.dat /pfs/out/`]);
      const input = new Input();
      const pfsInput = new PFSInput().setGlob('/*').setRepo(inputRepoName);
      input.setPfs(pfsInput);

      await pachClient.pps().createPipeline({
        pipeline: {name: 'createPipeline2'},
        transform: transform.toObject(),
        input: input.toObject(),
      });
      const updatedPipelines = await pachClient
        .pps()
        .listPipeline({projectIds: []});
      expect(updatedPipelines).toHaveLength(3);
    });
  });

  describe('deletePipeline', () => {
    it('should delete a pipeline', async () => {
      const {pachClient} = await createSandBox('deletePipeline');
      const pipelines = await pachClient.pps().listPipeline({projectIds: []});
      expect(pipelines).toHaveLength(2);

      await pachClient.pps().deletePipeline({
        projectId: 'default',
        pipeline: {name: 'deletePipeline-2'},
        force: true,
      });

      const updatedPipelines = await pachClient
        .pps()
        .listPipeline({projectIds: []});
      expect(updatedPipelines).toHaveLength(1);
    });
  });

  describe('deletePipelines', () => {
    it('should delete pipelines', async () => {
      const {pachClient} = await createSandBox('deletePipelines');
      const pipelines = await pachClient.pps().listPipeline({projectIds: []});
      expect(pipelines).toHaveLength(2);
      await pachClient.pps().deletePipelines({
        projectIds: ['default'],
      });
      const updatedPipelines = await pachClient
        .pps()
        .listPipeline({projectIds: []});
      expect(updatedPipelines).toHaveLength(0);
    });
  });

  describe('listDatums + inspectDatum', () => {
    jest.setTimeout(60000);

    it('should inspect a datum for a pipeline job', async () => {
      const {pachClient, inputRepoName} = await createSandBox('listDatums');
      const commit = await pachClient.pfs().startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: inputRepoName}},
      });

      const fileClient = await pachClient.pfs().modifyFile();

      await fileClient
        .setCommit(commit)
        .putFileFromBytes('dummyData.csv', Buffer.from('a,b,c'))
        .end();

      await pachClient.pfs().finishCommit({projectId: 'default', commit});
      const jobs = await pachClient.pps().listJobs({projectId: 'default'});

      const jobId = jobs[0]?.job?.id;
      expect(jobId).toBeDefined();

      await pachClient.pps().inspectJob({
        id: jobId || '',
        pipelineName: 'listDatums',
        wait: true,
        projectId: '',
      });

      const datums = await pachClient.pps().listDatums({
        projectId: 'default',
        jobId: jobId || '',
        pipelineName: 'listDatums',
      });

      expect(datums).toHaveLength(1);

      const datum = await pachClient.pps().inspectDatum({
        projectId: 'default',
        id: datums[0]?.datum?.id || '',
        jobId: jobId || '',
        pipelineName: 'listDatums',
      });

      const datumObject = datum.toObject();

      expect(datumObject.state).toEqual(DatumState.SUCCESS);
      expect(datumObject.dataList[0]?.file?.path).toBe('/dummyData.csv');
      expect(datumObject.dataList[0]?.sizeBytes).toBe(5);
    });

    it('should list datums for a pipeline job', async () => {
      const {pachClient, inputRepoName} = await createSandBox('listDatums');
      const commit = await pachClient.pfs().startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: inputRepoName}},
      });

      const fileClient = await pachClient.pfs().modifyFile();

      await fileClient
        .setCommit(commit)
        .putFileFromBytes('dummyData.csv', Buffer.from('a,b,c'))
        .end();

      await pachClient.pfs().finishCommit({projectId: 'default', commit});
      const jobs = await pachClient.pps().listJobs({projectId: 'default'});

      const jobId = jobs[0]?.job?.id;
      expect(jobId).toBeDefined();

      await pachClient.pps().inspectJob({
        id: jobId || '',
        pipelineName: 'listDatums',
        wait: true,
        projectId: '',
      });

      const datums = await pachClient.pps().listDatums({
        projectId: 'default',
        jobId: jobId || '',
        pipelineName: 'listDatums',
      });

      expect(datums).toHaveLength(1);
      expect(datums[0].state).toEqual(DatumState.SUCCESS);
      expect(datums[0].dataList[0]?.file?.path).toBe('/dummyData.csv');
      expect(datums[0].dataList[0]?.sizeBytes).toBe(5);
    });

    it('should filter datum list', async () => {
      const {pachClient, inputRepoName} = await createSandBox('listDatums');
      const commit = await pachClient.pfs().startCommit({
        projectId: 'default',
        branch: {name: 'master', repo: {name: inputRepoName}},
      });

      const fileClient = await pachClient.pfs().modifyFile();

      await fileClient
        .setCommit(commit)
        .putFileFromBytes('dummyData.csv', Buffer.from('a,b,c'))
        .end();

      await pachClient.pfs().finishCommit({projectId: 'default', commit});
      const jobs = await pachClient.pps().listJobs({projectId: 'default'});

      const jobId = jobs[0]?.job?.id;
      expect(jobId).toBeDefined();

      await pachClient.pps().inspectJob({
        id: jobId || '',
        pipelineName: 'listDatums',
        wait: true,
        projectId: '',
      });

      let datums = await pachClient.pps().listDatums({
        projectId: 'default',
        jobId: jobId || '',
        pipelineName: 'listDatums',
        filter: [DatumState.FAILED],
      });

      expect(datums).toHaveLength(0);

      datums = await pachClient.pps().listDatums({
        projectId: 'default',
        jobId: jobId || '',
        pipelineName: 'listDatums',
        filter: [DatumState.FAILED, DatumState.SUCCESS],
      });

      expect(datums).toHaveLength(1);
    });
  });
});
