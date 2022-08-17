import {FINISH_COMMIT_MUTATION} from '@dash-frontend/mutations/FinishCommit';
import {START_COMMIT_MUTATION} from '@dash-frontend/mutations/StartCommit';
import {GET_COMMIT_QUERY} from '@dash-frontend/queries/GetCommitQuery';
import {GET_COMMITS_QUERY} from '@dash-frontend/queries/GetCommitsQuery';

import {executeMutation, executeQuery} from '@dash-backend/testHelpers';
import {
  FinishCommitMutation,
  GetCommitsQuery,
  CommitQuery,
  StartCommitMutation,
} from '@graphqlTypes';

describe('resolvers/Commits', () => {
  describe('commit', () => {
    it('should inspect a given commit id', async () => {
      const projectId = '3';
      const repoName = 'cron';
      const id = '0918ac9d5daa76b86e3bb5e88e4c43a4';
      const {data, errors = []} = await executeQuery<CommitQuery>(
        GET_COMMIT_QUERY,
        {
          args: {
            projectId,
            repoName,
            id,
            withDiff: true,
          },
        },
      );

      expect(errors.length).toBe(0);
      expect(data?.commit.id).toBe(id);
      expect(data?.commit.repoName).toBe(repoName);
      expect(data?.commit.branch?.name).toBe('master');
      expect(data?.commit.description).toBe('added mako');
      expect(data?.commit.originKind).toBe('AUTO');
      expect(data?.commit.hasLinkedJob).toBeFalsy();
      expect(data?.commit.started).toBe(1614136389);
      expect(data?.commit.finished).toBe(1614136391);
      expect(data?.commit.sizeBytes).toBe(44276);
      expect(data?.commit.sizeDisplay).toBe('44.28 kB');
    });
  });
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
      expect(data?.commits.length).toBe(6);
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
  describe('startCommit', () => {
    it('should start a commit', async () => {
      const projectId = '3';
      const repo = 'cron';

      const {data: initialCommits} = await executeQuery<GetCommitsQuery>(
        GET_COMMITS_QUERY,
        {
          args: {projectId, repoName: repo, number: 100},
        },
      );

      expect(initialCommits?.commits.length).toBe(6);

      const {errors = []} = await executeMutation<StartCommitMutation>(
        START_COMMIT_MUTATION,
        {
          args: {projectId, repoName: repo, branchName: 'master'},
        },
      );

      expect(errors).toHaveLength(0);

      const {data: updatedCommits} = await executeQuery<GetCommitsQuery>(
        GET_COMMITS_QUERY,
        {
          args: {projectId, repoName: repo, number: 100},
        },
      );

      expect(updatedCommits?.commits.length).toBe(7);
    });
  });

  describe('finishCommit', () => {
    it('should finish a commit', async () => {
      const projectId = '3';
      const repo = 'cron';

      const {data} = await executeMutation<StartCommitMutation>(
        START_COMMIT_MUTATION,
        {
          args: {projectId, repoName: repo, branchName: 'master'},
        },
      );

      const {data: initialCommits} = await executeQuery<GetCommitsQuery>(
        GET_COMMITS_QUERY,
        {
          args: {projectId, repoName: repo, number: 100},
        },
      );

      const startCommit = initialCommits?.commits.find(
        (commit) => commit.id === data?.startCommit.id,
      );

      expect(startCommit?.finished).toEqual(-1);

      const {errors = []} = await executeMutation<FinishCommitMutation>(
        FINISH_COMMIT_MUTATION,
        {
          args: {
            projectId,
            commit: {
              branch: {
                name: data?.startCommit.branch.name,
                repo: {
                  name: data?.startCommit.branch.repo?.name,
                  type: data?.startCommit.branch.repo?.name,
                },
              },
              id: data?.startCommit.id,
            },
          },
        },
      );

      expect(errors).toHaveLength(0);

      const {data: updatedCommits} = await executeQuery<GetCommitsQuery>(
        GET_COMMITS_QUERY,
        {
          args: {projectId, repoName: repo, number: 100},
        },
      );

      const finishedCommit = updatedCommits?.commits.find(
        (commit) => commit.id === data?.startCommit.id,
      );

      expect(finishedCommit?.finished).not.toEqual(-1);
    });
  });
});
