import client from '@dash-backend/grpc/client';
import {QueryResolvers} from '@graphqlTypes';

interface RepoResolver {
  Query: {
    repos: QueryResolvers['repos'];
  };
}

const repoResolver: RepoResolver = {
  Query: {
    repos: async (_parent, _args, {pachdAddress = '', authToken = ''}) => {
      const repos = await client(pachdAddress, authToken).pfs().listRepo();

      return repos.map((repo) => ({
        createdAt: repo.created?.seconds || 0,
        description: repo.description,
        isPipelineOutput: false, // TODO: How do we derive this?
        name: repo.repo?.name || '',
        sizeInBytes: repo.sizeBytes,
      }));
    },
  },
};

export default repoResolver;
