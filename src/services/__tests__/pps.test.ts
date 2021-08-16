import crypto from 'crypto';

import {
  Input,
  PFSInput,
  PipelineState,
  Transform,
} from '@pachyderm/proto/pb/pps/pps_pb';

import client from 'client';

describe('services/pps', () => {
  afterAll(async () => {
    const pachClient = client({ssl: false, pachdAddress: 'localhost:30650'});
    const pps = pachClient.pps();
    await pps.deleteAll();
  });

  const createSandBox = async (name: string) => {
    const pachClient = client({ssl: false, pachdAddress: 'localhost:30650'});
    const pps = pachClient.pps();
    const pfs = pachClient.pfs();
    await pps.deleteAll();
    await pfs.deleteAll();
    const inputRepoName = crypto.randomBytes(5).toString('hex');
    await pfs.createRepo({repo: {name: inputRepoName}});
    const transform = new Transform()
      .setCmdList(['sh'])
      .setImage('alpine')
      .setStdinList([`cp /pfs/${inputRepoName}/*.dat /pfs/out/`]);
    const input = new Input();
    const pfsInput = new PFSInput().setGlob('/*').setRepo(inputRepoName);
    input.setPfs(pfsInput);

    await pps.createPipeline({
      name,
      transform: transform.toObject(),
      input: input.toObject(),
    });

    return {pachClient, inputRepoName};
  };

  describe('listPipeline', () => {
    it('should return a list of pipelines in the pachyderm cluster', async () => {
      const {pachClient} = await createSandBox('listPipeline');

      const pipelines = await pachClient.pps().listPipeline();

      expect(pipelines).toHaveLength(1);
      expect(pipelines[0].pipeline?.name).toEqual('listPipeline');
    });
  });

  describe('listJobs', () => {
    it('should return a list of jobs', async () => {
      const {pachClient} = await createSandBox('listJobs');

      const jobs = await pachClient.pps().listJobs();
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
        .inspectPipeline('inspectPipeline');

      expect(pipeline.pipeline?.name).toEqual('inspectPipeline');
      expect(pipeline.state).toEqual(PipelineState.PIPELINE_STARTING);
      expect(pipeline.details?.input?.pfs?.repo).toEqual(inputRepoName);
    });
  });

  describe('listJobSets', () => {
    it('should return a list of job sets', async () => {
      const {pachClient} = await createSandBox('listJobSets');

      const jobSets = await pachClient.pps().listJobSets();
      expect(jobSets).toHaveLength(1);
    });
  });

  describe('inspectJob', () => {
    it('should return details about the specified job', async () => {
      const {pachClient} = await createSandBox('inspectJob');
      const jobs = await pachClient.pps().listJobs();
      const job = await pachClient.pps().inspectJob({
        id: jobs[0].job?.id || '',
        pipelineName: 'inspectJob',
        projectId: '',
      });

      expect(job).toMatchObject(jobs[0]);
    });
  });

  describe('getLogs', () => {
    it('should return logs about the cluster', async () => {
      const {pachClient} = await createSandBox('getLogs');
      const logs = await pachClient.pps().getLogs({});
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
      const pipelines = await pachClient.pps().listPipeline();
      expect(pipelines).toHaveLength(1);

      const transform = new Transform()
        .setCmdList(['sh'])
        .setImage('alpine')
        .setStdinList([`cp /pfs/${inputRepoName}/*.dat /pfs/out/`]);
      const input = new Input();
      const pfsInput = new PFSInput().setGlob('/*').setRepo(inputRepoName);
      input.setPfs(pfsInput);

      await pachClient.pps().createPipeline({
        name: 'createPipeline2',
        transform: transform.toObject(),
        input: input.toObject(),
      });
      const updatedPipelines = await pachClient.pps().listPipeline();
      expect(updatedPipelines).toHaveLength(2);
    });
  });
});
