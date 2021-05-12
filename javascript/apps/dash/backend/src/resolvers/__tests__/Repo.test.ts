import {executeQuery} from '@dash-backend/testHelpers';
import {GET_REPO_QUERY} from '@dash-frontend/queries/GetRepoQuery';
import {RepoQuery} from '@graphqlTypes';

describe('resolvers/Repo', () => {
  const id = 'cron';
  const projectId = '3';

  it('should return a repo for a given id and projectId', async () => {
    const {data, errors = []} = await executeQuery<RepoQuery>(GET_REPO_QUERY, {
      args: {id, projectId},
    });

    expect(errors.length).toBe(0);
    expect(data?.repo.id).toBe(id);
    expect(data?.repo.name).toBe(id);
    expect(data?.repo.description).toBe('');
    expect(data?.repo.sizeDisplay).toBe('607.28 KB');
    expect(data?.repo.commits).toHaveLength(3);
    expect(data?.repo.linkedPipeline).toBeNull();
  });

  it('should return an error if a repo cannot be found', async () => {
    const {data, errors = []} = await executeQuery<RepoQuery>(GET_REPO_QUERY, {
      args: {id: 'bogus', projectId},
    });

    expect(errors.length).toBe(1);
    expect(data).toBeNull();
    expect(errors[0].extensions.code).toBe('NOT_FOUND');
  });
});
