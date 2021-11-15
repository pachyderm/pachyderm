import {CREATE_REPO_MUTATION} from '@dash-frontend/mutations/CreateRepo';
import {GET_REPO_QUERY} from '@dash-frontend/queries/GetRepoQuery';
import {Status} from '@grpc/grpc-js/build/src/constants';

import {executeMutation, executeQuery} from '@dash-backend/testHelpers';
import {CreateRepoMutation, RepoQuery} from '@graphqlTypes';

describe('resolvers/Repo', () => {
  const id = 'cron';
  const projectId = '3';

  describe('repo', () => {
    it('should return a repo for a given id and projectId', async () => {
      const {data, errors = []} = await executeQuery<RepoQuery>(
        GET_REPO_QUERY,
        {
          args: {id, projectId},
        },
      );

      expect(errors.length).toBe(0);
      expect(data?.repo.id).toBe(id);
      expect(data?.repo.name).toBe(id);
      expect(data?.repo.description).toBe('');
      expect(data?.repo.sizeDisplay).toBe('607.28 KB');
      expect(data?.repo.linkedPipeline).toBeNull();
    });

    it('should return an error if a repo cannot be found', async () => {
      const {data, errors = []} = await executeQuery<RepoQuery>(
        GET_REPO_QUERY,
        {
          args: {id: 'bogus', projectId},
        },
      );

      expect(errors.length).toBe(1);
      expect(data).toBeNull();
      expect(errors[0].extensions.code).toBe('NOT_FOUND');
    });
  });

  describe('createRepo', () => {
    it('should create a Repo', async () => {
      const {data: repo, errors: repoErrors = []} =
        await executeQuery<RepoQuery>(GET_REPO_QUERY, {
          args: {id: 'test', projectId},
        });
      expect(repo).toBeNull();
      expect(repoErrors[0].extensions.code).toEqual('NOT_FOUND');

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

      expect(errors.length).toBe(0);
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
      expect(errors.length).toBe(1);
      expect(errors[0].extensions.grpcCode).toEqual(Status.ALREADY_EXISTS);
      expect(errors[0].extensions.details).toEqual('repo cron already exists');
    });
    it('should update a repo', async () => {
      const {data: repo} = await executeQuery<RepoQuery>(GET_REPO_QUERY, {
        args: {id: 'cron', projectId},
      });
      expect(repo?.repo.description).toEqual('');

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
      expect(errors.length).toBe(0);
      expect(data?.createRepo.id).toBe('cron');
      expect(data?.createRepo.name).toBe('cron');
      expect(data?.createRepo.description).toBe('test repo');
    });
  });
});
