import pipelineJobs from '@dash-backend/mock/fixtures/pipelineJobs';
import {executeQuery} from '@dash-backend/testHelpers';
import {PIPELINE_JOBS_QUERY} from '@dash-frontend/queries/GetPipelineJobsQuery';
import {PipelineJobsQuery} from '@graphqlTypes';

describe('Pipeline Job Resolver', () => {
  it('should find jobs for a given project', async () => {
    const {data, errors = []} = await executeQuery<PipelineJobsQuery>(
      PIPELINE_JOBS_QUERY,
      {
        args: {
          projectId: '1',
        },
      },
    );

    expect(errors.length).toBe(0);
    expect(data?.pipelineJobs.length).toBe(pipelineJobs['1'].length);
  });

  it('should find jobs for a given pipelineId', async () => {
    const {data, errors = []} = await executeQuery<PipelineJobsQuery>(
      PIPELINE_JOBS_QUERY,
      {
        args: {
          projectId: '1',
          pipelineId: 'montage',
        },
      },
    );

    const expectedJobs = pipelineJobs['1'].filter(
      (jobs) => jobs.getPipeline()?.getName() === 'montage',
    );

    expect(errors.length).toBe(0);
    expect(data?.pipelineJobs.length).toBe(expectedJobs.length);
    expect(data?.pipelineJobs[0].id).toBe(
      expectedJobs[0].getPipelineJob()?.getId(),
    );
  });

  it('should return an empty set if no records exist', async () => {
    const {data, errors = []} = await executeQuery<PipelineJobsQuery>(
      PIPELINE_JOBS_QUERY,
      {
        args: {
          projectId: '1',
          pipelineId: 'bogus',
        },
      },
    );

    expect(errors.length).toBe(0);
    expect(data?.pipelineJobs.length).toBe(0);
  });
});
