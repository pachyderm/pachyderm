import {JOBS_BY_PIPELINE_QUERY} from '@dash-frontend/queries/GetJobsByPipelineQuery';
import {JOB_SET_QUERY} from '@dash-frontend/queries/GetJobsetQuery';
import {JOB_SETS_QUERY} from '@dash-frontend/queries/GetJobSetsQuery';
import {JOBS_QUERY} from '@dash-frontend/queries/GetJobsQuery';

import jobs from '@dash-backend/mock/fixtures/jobs';
import {executeQuery} from '@dash-backend/testHelpers';
import {
  JobSetQuery,
  JobSetsQuery,
  JobsQuery,
  JobState,
  JobsByPipelineQuery,
} from '@graphqlTypes';

describe('Jobs', () => {
  describe('Pipeline Jobs', () => {
    it('should find jobs for a given project', async () => {
      const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
        args: {
          projectId: 'Solar-Panel-Data-Sorting',
        },
      });

      expect(errors).toHaveLength(0);
      expect(data?.jobs.items).toHaveLength(
        jobs['Solar-Panel-Data-Sorting'].length,
      );
    });

    it('should return a specified number of jobs when a limit is given', async () => {
      const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
        args: {
          projectId: 'Solar-Panel-Data-Sorting',
          limit: 1,
        },
      });

      expect(errors).toHaveLength(0);
      expect(data?.jobs.items).toHaveLength(1);
    });

    it('should find jobs for a given pipelineId', async () => {
      const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
        args: {
          projectId: 'Solar-Panel-Data-Sorting',
          pipelineId: 'montage',
        },
      });

      const expectedJobs = jobs['Solar-Panel-Data-Sorting'].filter(
        (jobs) => jobs.getJob()?.getPipeline()?.getName() === 'montage',
      );

      expect(errors).toHaveLength(0);
      expect(data?.jobs.items).toHaveLength(expectedJobs.length);
      expect(data?.jobs.items[0].id).toBe(expectedJobs[0].getJob()?.getId());
    });

    it('should find jobs for multiple given pipelineIds', async () => {
      const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
        args: {projectId: 'Solar-Panel-Data-Sorting', pipelineIds: ['montage']},
      });
      const expectedJobs = jobs['Solar-Panel-Data-Sorting'].filter(
        (jobs) => jobs.getJob()?.getPipeline()?.getName() === 'montage',
      );

      expect(errors).toHaveLength(0);
      expect(data?.jobs.items).toHaveLength(expectedJobs.length);
      expect(data?.jobs.items[0].id).toBe(expectedJobs[0].getJob()?.getId());
    });

    it('should find jobs per pipeline for multiple given pipelineIds', async () => {
      const {data, errors = []} = await executeQuery<JobsByPipelineQuery>(
        JOBS_BY_PIPELINE_QUERY,
        {
          args: {
            projectId: 'Solar-Panel-Data-Sorting',
            pipelineIds: ['montage', 'edges'],
            limit: 1,
          },
        },
      );
      const montageJob = jobs['Solar-Panel-Data-Sorting'].find(
        (jobs) => jobs.getJob()?.getPipeline()?.getName() === 'montage',
      );
      const edgesJob = jobs['Solar-Panel-Data-Sorting'].find(
        (jobs) => jobs.getJob()?.getPipeline()?.getName() === 'edges',
      );
      const expectedJobs = [montageJob, edgesJob];

      expect(errors).toHaveLength(0);
      expect(data?.jobsByPipeline).toHaveLength(expectedJobs.length);
      expect(data?.jobsByPipeline[0].id).toBe(
        expectedJobs[0] && expectedJobs[0].getJob()?.getId(),
      );
      expect(data?.jobsByPipeline[1].id).toBe(
        expectedJobs[1] && expectedJobs[1].getJob()?.getId(),
      );
    });

    it('should find jobs for multiple given jobSetIds', async () => {
      const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
        args: {
          projectId: 'Solar-Panel-Data-Sorting',
          jobSetIds: ['23b9af7d5d4343219bc8e02ff44cd55a'],
        },
      });

      const expectedJobs = jobs['Solar-Panel-Data-Sorting'].filter(
        (jobs) => jobs.getJob()?.getId() === '23b9af7d5d4343219bc8e02ff44cd55a',
      );

      expect(errors).toHaveLength(0);
      expect(data?.jobs.items).toHaveLength(expectedJobs.length);
      expect(data?.jobs.items[0].id).toBe(expectedJobs[0].getJob()?.getId());
    });

    it('should return an empty set if no records exist', async () => {
      const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
        args: {
          projectId: 'Solar-Panel-Data-Sorting',
          pipelineId: 'bogus',
        },
      });

      expect(errors).toHaveLength(0);
      expect(data?.jobs.items).toHaveLength(0);
    });
  });

  describe('JobSets', () => {
    it('should return a jobset by id', async () => {
      const {data, errors = []} = await executeQuery<JobSetQuery>(
        JOB_SET_QUERY,
        {
          args: {
            projectId: 'Data-Cleaning-Process',
            id: '23b9af7d5d4343219bc8e02ff4acd33a',
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.jobSet.createdAt).toBe(1614136189);
      expect(data?.jobSet.id).toBe('23b9af7d5d4343219bc8e02ff4acd33a');
      expect(data?.jobSet.state).toBe(JobState.JOB_FAILURE);

      // assert order
      expect(data?.jobSet.jobs[0].pipelineName).toBe('likelihoods');
      expect(data?.jobSet?.jobs[1].pipelineName).toBe('models');
      expect(data?.jobSet?.jobs[2].pipelineName).toBe('joint_call');
      expect(data?.jobSet?.jobs[3].pipelineName).toBe('split');
      expect(data?.jobSet?.jobs[4].pipelineName).toBe('test');
    });

    it('should return an empty jobset if a jobset is not found', async () => {
      const {data, errors = []} = await executeQuery<JobSetQuery>(
        JOB_SET_QUERY,
        {
          args: {
            projectId: 'Data-Cleaning-Process',
            id: 'bogus',
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.jobSet.createdAt).toBeNull();
      expect(data?.jobSet.id).toBe('bogus');
      expect(data?.jobSet.state).toBe(JobState.JOB_SUCCESS);
      expect(data?.jobSet.jobs).toHaveLength(0);
    });

    it('should return all jobsets for a given project', async () => {
      const {data, errors = []} = await executeQuery<JobSetsQuery>(
        JOB_SETS_QUERY,
        {
          args: {
            projectId: 'Solar-Panel-Data-Sorting',
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.jobSets.items).toHaveLength(4);
      expect(data?.jobSets.items[0].id).toBe(
        '23b9af7d5d4343219bc8e02ff44cd55a',
      );
      expect(data?.jobSets.items[1].id).toBe(
        '33b9af7d5d4343219bc8e02ff44cd55a',
      );
      expect(data?.jobSets.items[2].id).toBe(
        '7798fhje5d4343219bc8e02ff4acd33a',
      );
      expect(data?.jobSets.items[3].id).toBe(
        'o90du4js5d4343219bc8e02ff4acd33a',
      );
    });

    it('should return the earliest created and latest finished job times', async () => {
      const {data} = await executeQuery<JobSetsQuery>(JOB_SETS_QUERY, {
        args: {
          projectId: '1',
        },
      });

      expect(data?.jobSets.items[0].createdAt).toBe(1614126189);
      expect(data?.jobSets.items[0].finishedAt).toBe(1616533103);
      expect(data?.jobSets.items[0].inProgress).toBe(false);
    });
  });

  describe('Jobs Paging', () => {
    const projectId = 'Solar-Panel-Data-Sorting';

    it('should return the first page if no cursor is specified', async () => {
      const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
        args: {projectId, limit: 3},
      });

      expect(errors).toHaveLength(0);
      expect(data?.jobs.items).toHaveLength(3);
      expect(data?.jobs.cursor).toEqual(
        expect.objectContaining({
          seconds: 1614126191,
          nanos: 100,
        }),
      );
      expect(data?.jobs.hasNextPage).toBe(true);
      expect(data?.jobs.items[0]).toEqual(
        expect.objectContaining({
          __typename: 'Job',
          startedAt: 1616533100,
          createdAt: 1616533099,
          finishedAt: 1616533103,
          id: '23b9af7d5d4343219bc8e02ff44cd55a',
        }),
      );
    });

    it('should return the next page if cursor is specified', async () => {
      const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
        args: {
          projectId,
          limit: 2,
          cursor: {
            seconds: 1614126190,
            nanos: 100,
          },
        },
      });

      expect(errors).toHaveLength(0);
      expect(data?.jobs.items).toHaveLength(2);
      expect(data?.jobs.cursor).toEqual(
        expect.objectContaining({
          seconds: 1614125000,
          nanos: 100,
        }),
      );
      expect(data?.jobs.hasNextPage).toBe(true);
    });

    it('should return no cursor if there are no more jobs after the requested page', async () => {
      const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
        args: {
          projectId,
          limit: 3,
          cursor: {
            seconds: 1614126191,
            nanos: 100,
          },
        },
      });

      expect(errors).toHaveLength(0);
      expect(data?.jobs.items).toHaveLength(2);
      expect(data?.jobs.cursor).toBeNull();
      expect(data?.jobs.hasNextPage).toBe(false);
    });

    it('should return the reverse order is reverse is true', async () => {
      const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
        args: {
          projectId,
          limit: 2,
          reverse: true,
        },
      });

      expect(errors).toHaveLength(0);
      expect(data?.jobs.items).toHaveLength(2);
      expect(data?.jobs.cursor).toEqual(
        expect.objectContaining({
          seconds: 1614125000,
          nanos: 100,
        }),
      );
      expect(data?.jobs.hasNextPage).toBe(true);
      expect(data?.jobs.items[0]).toEqual(
        expect.objectContaining({
          __typename: 'Job',
          startedAt: 1614123000,
          createdAt: 1614123000,
          id: 'o90du4js5d4343219bc8e02ff4acd33a',
        }),
      );
    });
  });
});
