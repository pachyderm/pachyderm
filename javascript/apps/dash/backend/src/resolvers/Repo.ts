import formatBytes from '@dash-backend/lib/formatBytes';
import {QueryResolvers, RepoResolvers} from '@graphqlTypes';

import {pipelineInfoToGQLPipeline} from './builders/pps';

interface RepoResolver {
  Query: {
    repo: QueryResolvers['repo'];
  };
  Repo: RepoResolvers;
}

const NUM_COMMITS = 100;

const repoResolver: RepoResolver = {
  Query: {
    repo: async (_parent, {args: {id}}, {pachClient}) => {
      const repo = await pachClient.pfs().inspectRepo(id);

      return {
        branches: repo?.branchesList.map((branch) => ({
          id: branch.name,
          name: branch.name,
        })),
        commits: [],
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
    commits: async (repo, _args, {pachClient}) => {
      try {
        return (await pachClient.pfs().listCommit(repo.id, NUM_COMMITS))
          .map((commit) => ({
            branch: commit.commit?.branch && {
              id: commit.commit?.branch?.name,
              name: commit.commit?.branch?.name,
            },
            description: commit.description,
            finished: commit.finished?.seconds || 0,
            id: commit.commit?.id || '',
            started: commit.started?.seconds || 0,
            sizeBytes: commit.details?.sizeBytes || 0,
            sizeDisplay: formatBytes(commit.details?.sizeBytes || 0),
          }))
          .reverse();
      } catch (err) {
        return [];
      }
    },
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
