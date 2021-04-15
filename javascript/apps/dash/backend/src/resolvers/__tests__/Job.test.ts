import jobs from '@dash-backend/mock/fixtures/jobs';
import {executeOperation} from '@dash-backend/testHelpers';
import {GetJobsQuery} from '@graphqlTypes';

describe('Jobs resolver', () => {
  it('should find jobs for a given project', async () => {
    const {data, errors = []} = await executeOperation<GetJobsQuery>(
      'getJobs',
      {
        args: {
          projectId: '1',
        },
      },
    );

    expect(errors.length).toBe(0);
    expect(data?.jobs.length).toBe(jobs['1'].length);
  });

  it('should find jobs for a given pipelineId', async () => {
    const {data, errors = []} = await executeOperation<GetJobsQuery>(
      'getJobs',
      {
        args: {
          projectId: '1',
          pipelineId: 'montage',
        },
      },
    );

    const expectedJobs = jobs['1'].filter(
      (jobs) => jobs.getPipeline()?.getName() === 'montage',
    );

    expect(errors.length).toBe(0);
    expect(data?.jobs.length).toBe(expectedJobs.length);
    expect(data?.jobs[0].id).toBe(expectedJobs[0].getJob()?.getId());
  });

  it('should return an empty set if no records exist', async () => {
    const {data, errors = []} = await executeOperation<GetJobsQuery>(
      'getJobs',
      {
        args: {
          projectId: '1',
          pipelineId: 'bogus',
        },
      },
    );

    expect(errors.length).toBe(0);
    expect(data?.jobs.length).toBe(0);
  });
});
