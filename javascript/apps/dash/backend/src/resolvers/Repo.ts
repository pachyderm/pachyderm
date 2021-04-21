import formatBytes from '@dash-backend/lib/formatBytes';
import {QueryResolvers, RepoResolvers} from '@graphqlTypes';

import {pipelineInfoToGQLPipeline} from './builders/pps';

interface RepoResolver {
  Query: {
    repo: QueryResolvers['repo'];
  };
  Repo: RepoResolvers;
}

const repoResolver: RepoResolver = {
  Query: {
    repo: async (_parent, {args: {id}}, {pachClient}) => {
      const repo = await pachClient.pfs().inspectRepo(id);

      return {
        createdAt: repo?.created?.seconds || 0,
        description: repo?.description || '',
        id: repo?.repo?.name || '',
        name: repo?.repo?.name || '',
        sizeBytes: repo?.sizeBytes || 0,
        sizeDisplay: formatBytes(repo?.sizeBytes || 0),
      };
    },
  },
  Repo: {
    linkedPipeline: async (repo, _args, {pachClient}) => {
      try {
        return pipelineInfoToGQLPipeline(
          await pachClient.pps().inspectPipeline(repo.id),
        );
      } catch (err) {
        return null;
      }
    },
  },
};

export default repoResolver;
