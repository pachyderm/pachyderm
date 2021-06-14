import jobs from '@dash-backend/mock/fixtures/jobs';
import {executeQuery} from '@dash-backend/testHelpers';
import {JOBS_QUERY} from '@dash-frontend/queries/GetJobsQuery';
import {JobsQuery} from '@graphqlTypes';

describe('Pipeline Job Resolver', () => {
  it('should find jobs for a given project', async () => {
    const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
      args: {
        projectId: '1',
      },
    });

    expect(errors.length).toBe(0);
    expect(data?.jobs.length).toBe(jobs['1'].length);
  });

  it('should find jobs for a given pipelineId', async () => {
    const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
      args: {
        projectId: '1',
        pipelineId: 'montage',
      },
    });

    const expectedJobs = jobs['1'].filter(
      (jobs) => jobs.getJob()?.getPipeline()?.getName() === 'montage',
    );

    expect(errors.length).toBe(0);
    expect(data?.jobs.length).toBe(expectedJobs.length);
    expect(data?.jobs[0].id).toBe(expectedJobs[0].getJob()?.getId());
  });

  it('should return an empty set if no records exist', async () => {
    const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
      args: {
        projectId: '1',
        pipelineId: 'bogus',
      },
    });

    expect(errors.length).toBe(0);
    expect(data?.jobs.length).toBe(0);
  });
});
