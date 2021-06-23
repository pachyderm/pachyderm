import formatBytes from '@dash-backend/lib/formatBytes';
import {QueryResolvers, RepoResolvers} from '@graphqlTypes';

import {pipelineInfoToGQLPipeline, repoInfoToGQLRepo} from './builders/pps';

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
      return repoInfoToGQLRepo(await pachClient.pfs().inspectRepo(id));
    },
  },
  Repo: {
    commits: async (repo, _args, {pachClient}) => {
      try {
        return (await pachClient.pfs().listCommit(repo.id, NUM_COMMITS))
          .map((commit) => ({
            repoName: repo.name,
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
