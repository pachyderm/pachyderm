import {FINISH_COMMIT_MUTATION} from '@dash-frontend/mutations/FinishCommit';
import {START_COMMIT_MUTATION} from '@dash-frontend/mutations/StartCommit';
import {GET_COMMIT_QUERY} from '@dash-frontend/queries/GetCommitQuery';
import {COMMIT_SEARCH_QUERY} from '@dash-frontend/queries/GetCommitSearchQuery';
import {GET_COMMITS_QUERY} from '@dash-frontend/queries/GetCommitsQuery';
import {FIND_COMMITS_QUERY} from '@dash-frontend/queries/GetFindCommitsQuery';

import {executeMutation, executeQuery} from '@dash-backend/testHelpers';
import {
  FinishCommitMutation,
  GetCommitsQuery,
  CommitQuery,
  StartCommitMutation,
  CommitSearchQuery,
  FindCommitsQuery,
} from '@graphqlTypes';

describe('resolvers/Commits', () => {
  describe('commit', () => {
    it('should inspect a given commit id', async () => {
      const projectId = 'Solar-Power-Data-Logger-Team-Collab';
      const repoName = 'cron';
      const id = '0918ac9d5daa76b86e3bb5e88e4c43a4';
      const {data, errors = []} = await executeQuery<CommitQuery>(
        GET_COMMIT_QUERY,
        {
          args: {
            projectId,
            repoName,
            id,
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.commit?.id).toBe(id);
      expect(data?.commit?.repoName).toBe(repoName);
      expect(data?.commit?.branch?.name).toBe('master');
      expect(data?.commit?.description).toBe('added mako');
      expect(data?.commit?.originKind).toBe('AUTO');
      expect(data?.commit?.started).toBe(1614136389);
      expect(data?.commit?.finished).toBe(1614136391);
      expect(data?.commit?.sizeBytes).toBe(44276);
      expect(data?.commit?.sizeDisplay).toBe('44.28 kB');
    });

    it('should inspect the latest commit if not given commit id', async () => {
      const projectId = 'Solar-Power-Data-Logger-Team-Collab';
      const repoName = 'cron';
      const {data, errors = []} = await executeQuery<CommitQuery>(
        GET_COMMIT_QUERY,
        {
          args: {
            projectId,
            repoName,
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.commit?.id).toBe('9d5daa0918ac4c43a476b86e3bb5e88e');
      expect(data?.commit?.repoName).toBe(repoName);
    });
  });

  describe('commitSearch', () => {
    it('should return a commit based on search criteria', async () => {
      const projectId = 'Solar-Power-Data-Logger-Team-Collab';
      const repoName = 'cron';
      const id = '9d5daa0918ac4c43a476b86e3bb5e88e';

      const {data, errors = []} = await executeQuery<CommitSearchQuery>(
        COMMIT_SEARCH_QUERY,
        {
          args: {
            id,
            repoName,
            projectId,
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.commitSearch).toMatchObject({
        id,
        repoName,
        branch: {name: 'master'},
        description: 'added shinra hq building specs',
      });
    });
    it('should return error if id is not the correct format', async () => {
      const projectId = 'Solar-Power-Data-Logger-Team-Collab';
      const repoName = 'cron';
      const id = '0918zzzd5daa76b86e3bb5e88e4c43a4';

      const {data, errors = []} = await executeQuery<CommitSearchQuery>(
        COMMIT_SEARCH_QUERY,
        {
          args: {
            id,
            repoName,
            projectId,
          },
        },
      );

      expect(errors).toHaveLength(1);
      expect(errors[0].extensions.code).toBe('INVALID_ARGUMENT');
      expect(data?.commitSearch).toBeUndefined();
    });

    it('should return null if commit does not exist', async () => {
      const projectId = 'Solar-Power-Data-Logger-Team-Collab';
      const repoName = 'cron';
      const id = '2d5daa0918ac4c43a412386e3bb5e88e';

      const {data, errors = []} = await executeQuery<CommitSearchQuery>(
        COMMIT_SEARCH_QUERY,
        {
          args: {
            id,
            repoName,
            projectId,
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.commitSearch).toBeNull();
    });
  });

  describe('findCommits', () => {
    it('should return error if request is missing both commitId and branchId parameters', async () => {
      const projectId = 'Solar-Power-Data-Logger-Team-Collab';
      const repoId = 'cron';

      const {data, errors = []} = await executeQuery<FindCommitsQuery>(
        FIND_COMMITS_QUERY,
        {
          args: {
            projectId,
            repoId,
            filePath: '/kitten.png',
          },
        },
      );

      expect(errors).toHaveLength(1);
      expect(errors[0].extensions.code).toBe('INVALID_ARGUMENT');
      expect(data?.findCommits.commits).toBeUndefined();
    });

    it('should search by commitId', async () => {
      const projectId = 'Solar-Power-Data-Logger-Team-Collab';
      const repoId = 'cron';
      const branchId = 'master';

      const {errors = []} = await executeQuery<FindCommitsQuery>(
        FIND_COMMITS_QUERY,
        {
          args: {
            projectId,
            repoId,
            branchId,
            filePath: '/kitten.png',
          },
        },
      );

      expect(errors).toHaveLength(0);
    });
    it('should search by branchId', async () => {
      const projectId = 'Solar-Power-Data-Logger-Team-Collab';
      const repoId = 'cron';
      const commitId = '2d5daa0918ac4c43a412386e3bb5e88e';

      const {errors = []} = await executeQuery<FindCommitsQuery>(
        FIND_COMMITS_QUERY,
        {
          args: {
            projectId,
            repoId,
            commitId,
            filePath: '/kitten.png',
          },
        },
      );

      expect(errors).toHaveLength(0);
    });
  });

  describe('commits', () => {
    it('should return commits for a given repo', async () => {
      const projectId = 'Solar-Power-Data-Logger-Team-Collab';
      const repo = 'cron';

      const {data, errors = []} = await executeQuery<GetCommitsQuery>(
        GET_COMMITS_QUERY,
        {
          args: {projectId, repoName: repo, number: 100},
        },
      );

      expect(errors).toHaveLength(0);
      const commits = data?.commits.items;
      expect(commits).toHaveLength(6);
    });

    it('should return commits for a given branch', async () => {
      const projectId = 'Data-Cleaning-Process';
      const repo = 'training';
      const branch = 'master';
      const {data, errors = []} = await executeQuery<GetCommitsQuery>(
        GET_COMMITS_QUERY,
        {
          args: {projectId, repoName: repo, number: 100, branchName: branch},
        },
      );

      expect(errors).toHaveLength(0);
      const commits = data?.commits.items;
      expect(commits).toHaveLength(1);
    });

    describe('paging', () => {
      it('should return the first page if no cursor is specified', async () => {
        const projectId = 'Solar-Power-Data-Logger-Team-Collab';
        const repo = 'cron';
        const {data, errors = []} = await executeQuery<GetCommitsQuery>(
          GET_COMMITS_QUERY,
          {
            args: {projectId, repoName: repo, number: 3},
          },
        );

        expect(errors).toHaveLength(0);
        const commits = data?.commits.items;
        expect(commits).toHaveLength(3);
        expect(data?.commits.cursor).toEqual(
          expect.objectContaining({
            seconds: 1614136389,
            nanos: 0,
          }),
        );
      });

      it('should return the next page if cursor is specified', async () => {
        const projectId = 'Solar-Power-Data-Logger-Team-Collab';
        const repo = 'cron';
        const {data, errors = []} = await executeQuery<GetCommitsQuery>(
          GET_COMMITS_QUERY,
          {
            args: {
              projectId,
              repoName: repo,
              number: 3,
              cursor: {
                seconds: 1614136289,
                nanos: 0,
              },
            },
          },
        );

        expect(errors).toHaveLength(0);
        const commits = data?.commits.items;
        expect(errors).toHaveLength(0);
        expect(commits).toHaveLength(3);
        expect(data?.commits.cursor).toEqual(
          expect.objectContaining({
            seconds: 1614133389,
            nanos: 3,
          }),
        );
      });

      it('should return no cursor if there are no more commits after the requested page', async () => {
        const projectId = 'Solar-Power-Data-Logger-Team-Collab';
        const repo = 'cron';
        const {data, errors = []} = await executeQuery<GetCommitsQuery>(
          GET_COMMITS_QUERY,
          {
            args: {
              projectId,
              repoName: repo,
              number: 3,
              cursor: {
                seconds: 1614133389,
                nanos: 3,
              },
            },
          },
        );

        expect(errors).toHaveLength(0);
        const commits = data?.commits.items;
        expect(commits).toHaveLength(1);
        expect(data?.commits.cursor).toBeNull();
      });

      it('should return error if both cursors are passed to the query', async () => {
        const projectId = 'Solar-Power-Data-Logger-Team-Collab';
        const repoName = 'cron';
        const id = '0918zzzd5daa76b86e3bb5e88e4c43a4';

        const {data, errors = []} = await executeQuery<GetCommitsQuery>(
          GET_COMMITS_QUERY,
          {
            args: {
              projectId,
              repoName,
              number: 3,
              cursor: {
                seconds: 1614133389,
                nanos: 3,
              },
              commitIdCursor: id,
            },
          },
        );

        expect(errors).toHaveLength(1);
        expect(errors[0].extensions.code).toBe('INVALID_ARGUMENT');
        expect(data?.commits).toBeUndefined();
      });

      it('should return error if both cursor and branch are passed to the query', async () => {
        const projectId = 'Solar-Power-Data-Logger-Team-Collab';
        const repoName = 'cron';
        const branchName = 'test';

        const {data, errors = []} = await executeQuery<GetCommitsQuery>(
          GET_COMMITS_QUERY,
          {
            args: {
              projectId,
              repoName,
              number: 3,
              cursor: {
                seconds: 1614133389,
                nanos: 3,
              },
              branchName,
            },
          },
        );

        expect(errors).toHaveLength(1);
        expect(errors[0].extensions.code).toBe('INVALID_ARGUMENT');
        expect(data?.commits).toBeUndefined();
      });

      it('should return the reverse order if reverse is true', async () => {
        const projectId = 'Solar-Power-Data-Logger-Team-Collab';
        const repo = 'cron';
        const {data, errors = []} = await executeQuery<GetCommitsQuery>(
          GET_COMMITS_QUERY,
          {
            args: {projectId, repoName: repo, number: 3, reverse: true},
          },
        );

        expect(errors).toHaveLength(0);
        const commits = data?.commits.items;
        expect(commits).toHaveLength(3);
        expect(data?.commits.cursor).toEqual(
          expect.objectContaining({
            seconds: 1614133389,
            nanos: 0,
          }),
        );
        expect(data?.commits.items[0]).toEqual(
          expect.objectContaining({
            __typename: 'Commit',
            description: 'in progress 3',
            id: '0518ac9d5daa76b86e3bb5e88e4c43a5',
            started: 1614133389,
          }),
        );
      });
    });
  });
  describe('startCommit', () => {
    it('should start a commit', async () => {
      const projectId = 'Solar-Power-Data-Logger-Team-Collab';
      const repo = 'cron';

      const {data: initialData} = await executeQuery<GetCommitsQuery>(
        GET_COMMITS_QUERY,
        {
          args: {projectId, repoName: repo, number: 100},
        },
      );
      const initialCommits = initialData?.commits.items;
      expect(initialCommits).toHaveLength(6);

      const {errors = []} = await executeMutation<StartCommitMutation>(
        START_COMMIT_MUTATION,
        {
          args: {projectId, repoName: repo, branchName: 'master'},
        },
      );

      expect(errors).toHaveLength(0);

      const {data: updatedData} = await executeQuery<GetCommitsQuery>(
        GET_COMMITS_QUERY,
        {
          args: {projectId, repoName: repo, number: 100},
        },
      );

      const updatedCommits = updatedData?.commits.items;
      expect(updatedCommits).toHaveLength(7);
    });
  });

  describe('finishCommit', () => {
    it('should finish a commit', async () => {
      const projectId = 'Solar-Power-Data-Logger-Team-Collab';
      const repo = 'cron';

      const {data} = await executeMutation<StartCommitMutation>(
        START_COMMIT_MUTATION,
        {
          args: {projectId, repoName: repo, branchName: 'master'},
        },
      );

      const {data: initialData} = await executeQuery<GetCommitsQuery>(
        GET_COMMITS_QUERY,
        {
          args: {projectId, repoName: repo, number: 100},
        },
      );

      const initialCommits = initialData?.commits.items;
      const startCommit = initialCommits?.find(
        (commit) => commit.id === data?.startCommit.id,
      );

      expect(startCommit?.finished).toBe(-1);

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

      const {data: updatedData} = await executeQuery<GetCommitsQuery>(
        GET_COMMITS_QUERY,
        {
          args: {projectId, repoName: repo, number: 100},
        },
      );

      const updatedCommits = updatedData?.commits.items;
      const finishedCommit = updatedCommits?.find(
        (commit) => commit.id === data?.startCommit.id,
      );

      expect(finishedCommit?.finished).not.toBe(-1);
    });
  });
});
