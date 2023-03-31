import {CREATE_BRANCH_MUTATION} from '@dash-frontend/mutations/CreateBranch';
import {GET_BRANCHES_QUERY} from '@dash-frontend/queries/GetBranchesQuery';
import {GET_COMMITS_QUERY} from '@dash-frontend/queries/GetCommitsQuery';
import {GET_REPO_QUERY} from '@dash-frontend/queries/GetRepoQuery';

import {executeMutation, executeQuery} from '@dash-backend/testHelpers';
import {
  RepoQuery,
  CreateBranchMutation,
  GetCommitsQuery,
  Commit,
  GetBranchesQuery,
} from '@graphqlTypes';
describe('resolvers/Branch', () => {
  describe('createBranch', () => {
    const projectId = 'Solar-Power-Data-Logger-Team-Collab';
    it('should create a branch', async () => {
      const {data: response} = await executeQuery<RepoQuery>(GET_REPO_QUERY, {
        args: {id: 'cron', projectId},
      });
      expect(response?.repo.branches).toHaveLength(1);

      const {data, errors = []} = await executeMutation<CreateBranchMutation>(
        CREATE_BRANCH_MUTATION,
        {
          args: {
            branch: {name: 'dev', repo: {name: 'cron'}},
            projectId,
          },
        },
      );
      expect(errors).toHaveLength(0);
      expect(data?.createBranch?.name).toBe('dev');
      const {data: response2} = await executeQuery<RepoQuery>(GET_REPO_QUERY, {
        args: {id: 'cron', projectId},
      });
      expect(response2?.repo.branches).toHaveLength(2);
    });

    it('should update a branch head', async () => {
      const {data: response} = await executeQuery<GetCommitsQuery>(
        GET_COMMITS_QUERY,
        {
          args: {repoName: 'cron', projectId},
        },
      );

      const commits = response?.commits.items as Commit[];
      expect(commits).toHaveLength(6);

      const {data, errors = []} = await executeMutation<CreateBranchMutation>(
        CREATE_BRANCH_MUTATION,
        {
          args: {
            branch: {name: 'master', repo: {name: 'cron'}},
            head: {id: '0918ac4c43a476b86e3bb5e88e9d5daa'},
            projectId,
          },
        },
      );
      expect(errors).toHaveLength(0);
      expect(data?.createBranch?.name).toBe('master');
      const {data: response2} = await executeQuery<GetCommitsQuery>(
        GET_COMMITS_QUERY,
        {
          args: {repoName: 'cron', projectId},
        },
      );
      const commits2 = response2?.commits.items as Commit[];
      expect(commits2).toHaveLength(2);
    });
  });

  describe('branches', () => {
    it('should return a list of branches on a repo', async () => {
      const projectId = 'default';
      const repo = 'training';

      const {data, errors = []} = await executeQuery<GetBranchesQuery>(
        GET_BRANCHES_QUERY,
        {
          args: {projectId, repoName: repo},
        },
      );

      expect(errors).toHaveLength(0);
      const branches = data?.branches;
      expect(branches).toHaveLength(3);
      expect(branches?.[0]?.name).toBe('master');
      expect(branches?.[1]?.name).toBe('develop');
      expect(branches?.[2]?.name).toBe('test');
    });
  });
});
