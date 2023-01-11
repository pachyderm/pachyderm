import {JOB_SET_QUERY} from '@dash-frontend/queries/GetJobsetQuery';
import {JOB_SETS_QUERY} from '@dash-frontend/queries/GetJobSetsQuery';
import {JOBS_QUERY} from '@dash-frontend/queries/GetJobsQuery';

import jobs from '@dash-backend/mock/fixtures/jobs';
import {executeQuery} from '@dash-backend/testHelpers';
import {JobSetQuery, JobSetsQuery, JobsQuery, JobState} from '@graphqlTypes';

describe('Jobs', () => {
  describe('Pipeline Jobs', () => {
    it('should find jobs for a given project', async () => {
      const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
        args: {
          projectId: '1',
        },
      });

      expect(errors).toHaveLength(0);
      expect(data?.jobs).toHaveLength(jobs['1'].length);
    });

    it('should return a specified number of jobs when a limit is given', async () => {
      const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
        args: {
          projectId: '1',
          limit: 1,
        },
      });

      expect(errors).toHaveLength(0);
      expect(data?.jobs).toHaveLength(1);
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

      expect(errors).toHaveLength(0);
      expect(data?.jobs).toHaveLength(expectedJobs.length);
      expect(data?.jobs[0].id).toBe(expectedJobs[0].getJob()?.getId());
    });

    it('should return an empty set if no records exist', async () => {
      const {data, errors = []} = await executeQuery<JobsQuery>(JOBS_QUERY, {
        args: {
          projectId: '1',
          pipelineId: 'bogus',
        },
      });

      expect(errors).toHaveLength(0);
      expect(data?.jobs).toHaveLength(0);
    });
  });

  describe('JobSets', () => {
    it('should return a jobset by id', async () => {
      const {data, errors = []} = await executeQuery<JobSetQuery>(
        JOB_SET_QUERY,
        {
          args: {
            projectId: '2',
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
            projectId: '2',
            id: 'bogus',
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.jobSet.createdAt).toBe(null);
      expect(data?.jobSet.id).toBe('bogus');
      expect(data?.jobSet.state).toBe(JobState.JOB_SUCCESS);
      expect(data?.jobSet.jobs).toHaveLength(0);
    });

    it('should return all jobsets for a given project', async () => {
      const {data, errors = []} = await executeQuery<JobSetsQuery>(
        JOB_SETS_QUERY,
        {
          args: {
            projectId: '1',
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.jobSets).toHaveLength(4);
      expect(data?.jobSets[0].id).toBe('23b9af7d5d4343219bc8e02ff44cd55a');
      expect(data?.jobSets[1].id).toBe('33b9af7d5d4343219bc8e02ff44cd55a');
      expect(data?.jobSets[2].id).toBe('7798fhje5d4343219bc8e02ff4acd33a');
      expect(data?.jobSets[3].id).toBe('o90du4js5d4343219bc8e02ff4acd33a');
    });
  });
});
