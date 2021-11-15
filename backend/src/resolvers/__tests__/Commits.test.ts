import {GET_COMMITS_QUERY} from '@dash-frontend/queries/GetCommitsQuery';

import {executeQuery} from '@dash-backend/testHelpers';
import {GetCommitsQuery} from '@graphqlTypes';

describe('resolvers/Commits', () => {
  describe('commits', () => {
    it('should return commits for a given repo', async () => {
      const projectId = '3';
      const repo = 'cron';
      const {data, errors = []} = await executeQuery<GetCommitsQuery>(
        GET_COMMITS_QUERY,
        {
          args: {projectId, repoName: repo, number: 100},
        },
      );

      expect(errors.length).toBe(0);
      expect(data?.commits.length).toBe(3);
      expect(data?.commits[0].hasLinkedJob).toBeFalsy();
    });

    it('should return commits for a given branch', async () => {
      const projectId = '2';
      const repo = 'training';
      const branch = 'master';
      const {data, errors = []} = await executeQuery<GetCommitsQuery>(
        GET_COMMITS_QUERY,
        {
          args: {projectId, repoName: repo, number: 100, branchName: branch},
        },
      );

      expect(errors.length).toBe(0);
      expect(data?.commits.length).toBe(1);
    });
  });
});
