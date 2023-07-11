import {CREATE_REPO_MUTATION} from '@dash-frontend/mutations/CreateRepo';
import {GET_REPO_QUERY} from '@dash-frontend/queries/GetRepoQuery';
import {GET_REPOS_QUERY} from '@dash-frontend/queries/GetReposQuery';
import {GET_REPOS_WITH_COMMIT_QUERY} from '@dash-frontend/queries/GetReposWithCommitQuery';
import {GET_REPO_WITH_COMMIT_QUERY} from '@dash-frontend/queries/GetRepoWithCommitQuery';
import {Status} from '@grpc/grpc-js/build/src/constants';

import {executeMutation, executeQuery} from '@dash-backend/testHelpers';
import {
  CreateRepoMutation,
  RepoQuery,
  ReposQuery,
  RepoWithCommitQuery,
  ReposWithCommitQuery,
} from '@graphqlTypes';

describe('resolvers/Repo', () => {
  const id = 'cron';
  const projectId = 'Solar-Power-Data-Logger-Team-Collab';

  describe('repo', () => {
    it('should return a repo for a given id and projectId', async () => {
      const {data, errors = []} = await executeQuery<RepoQuery>(
        GET_REPO_QUERY,
        {
          args: {id, projectId},
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.repo).toEqual(
        expect.objectContaining({
          id,
          name: id,
          description: 'cron job',
          sizeDisplay: '621.86 kB',
        }),
      );
    });

    it('should return an error if a repo cannot be found', async () => {
      const {data, errors = []} = await executeQuery<RepoQuery>(
        GET_REPO_QUERY,
        {
          args: {id: 'bogus', projectId},
        },
      );

      expect(errors).toHaveLength(1);
      expect(data).toBeNull();
      expect(errors[0].extensions.code).toBe('NOT_FOUND');
    });

    it('should return repo with last commit', async () => {
      const {data, errors = []} = await executeQuery<RepoWithCommitQuery>(
        GET_REPO_WITH_COMMIT_QUERY,
        {
          args: {id, projectId},
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.repo.id).toBe(id);
      expect(data?.repo?.lastCommit?.id).toBe(
        '9d5daa0918ac4c43a476b86e3bb5e88e',
      );
    });
  });

  describe('repos', () => {
    it('should return repo list', async () => {
      const {data} = await executeQuery<ReposQuery>(GET_REPOS_QUERY, {
        args: {projectId},
      });

      expect(data?.repos).toHaveLength(2);
      expect(data?.repos[0]?.id).toBe('cron');
      expect(data?.repos[1]?.id).toBe('processor');
    });

    it('should return repo list filtered by globalId', async () => {
      const {data} = await executeQuery<ReposQuery>(GET_REPOS_QUERY, {
        args: {projectId: 'Solar-Panel-Data-Sorting'},
      });
      expect(data?.repos).toHaveLength(3);
      expect(data?.repos[0]?.id).toBe('montage');
      expect(data?.repos[1]?.id).toBe('edges');
      expect(data?.repos[2]?.id).toBe('images');

      const {data: filteredData} = await executeQuery<ReposQuery>(
        GET_REPOS_QUERY,
        {
          args: {
            projectId: 'Solar-Panel-Data-Sorting',
            jobSetId: '33b9af7d5d4343219bc8e02ff44cd55a',
          },
        },
      );

      expect(filteredData?.repos).toHaveLength(1);
      expect(filteredData?.repos[0]?.id).toBe('montage');
    });

    it('should return repo list with last commit for each repo', async () => {
      const {data} = await executeQuery<ReposWithCommitQuery>(
        GET_REPOS_WITH_COMMIT_QUERY,
        {
          args: {projectId},
        },
      );

      expect(data?.repos).toHaveLength(2);
      expect(data?.repos[0]?.id).toBe('cron');
      expect(data?.repos[0]?.lastCommit?.id).toBe(
        '9d5daa0918ac4c43a476b86e3bb5e88e',
      );
      expect(data?.repos[1]?.id).toBe('processor');
      expect(data?.repos[1]?.lastCommit?.id).toBe(
        'f4e23cf347c342d98bd9015e4c3ad52a',
      );
    });
  });

  describe('createRepo', () => {
    it('should create a Repo', async () => {
      const {data: repo, errors: repoErrors = []} =
        await executeQuery<RepoQuery>(GET_REPO_QUERY, {
          args: {id: 'test', projectId},
        });
      expect(repo).toBeNull();
      expect(repoErrors[0].extensions.code).toBe('NOT_FOUND');

      const {data, errors = []} = await executeMutation<CreateRepoMutation>(
        CREATE_REPO_MUTATION,
        {
          args: {
            name: 'test',
            description: 'test repo',
            update: false,
            projectId,
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.createRepo.id).toBe('test');
      expect(data?.createRepo.name).toBe('test');
      expect(data?.createRepo.description).toBe('test repo');
      expect(data?.createRepo.sizeDisplay).toBe('0 B');
    });
    it('should return an error if a repo with that name already exists', async () => {
      const {errors = []} = await executeMutation<CreateRepoMutation>(
        CREATE_REPO_MUTATION,
        {
          args: {
            name: 'cron',
            description: 'test repo',
            update: false,
            projectId,
          },
        },
      );
      expect(errors).toHaveLength(1);
      expect(errors[0].extensions.grpcCode).toEqual(Status.ALREADY_EXISTS);
      expect(errors[0].extensions.details).toBe('repo cron already exists');
    });
    it('should update a repo', async () => {
      const {data: repo} = await executeQuery<RepoQuery>(GET_REPO_QUERY, {
        args: {id: 'cron', projectId},
      });
      expect(repo?.repo.description).toBe('cron job');

      const {data, errors = []} = await executeMutation<CreateRepoMutation>(
        CREATE_REPO_MUTATION,
        {
          args: {
            name: 'cron',
            description: 'test repo',
            update: true,
            projectId,
          },
        },
      );
      expect(errors).toHaveLength(0);
      expect(data?.createRepo.id).toBe('cron');
      expect(data?.createRepo.name).toBe('cron');
      expect(data?.createRepo.description).toBe('test repo');
    });
  });
});
