import {JobInfo} from '@pachyderm/node-pachyderm';

import formatBytes from '@dash-backend/lib/formatBytes';
import getSizeBytes from '@dash-backend/lib/getSizeBytes';
import {PachClient} from '@dash-backend/lib/types';
import {QueryResolvers, RepoResolvers} from '@graphqlTypes';

import {pipelineInfoToGQLPipeline, repoInfoToGQLRepo} from './builders/pps';

interface RepoResolver {
  Query: {
    repo: QueryResolvers['repo'];
  };
  Repo: RepoResolvers;
}

const NUM_COMMITS = 100;

const getJobs = (pipelineId: string, pachClient: PachClient) => {
  try {
    return pachClient.pps().listJobs({pipelineId, limit: NUM_COMMITS});
  } catch (err) {
    return [];
  }
};

const repoResolver: RepoResolver = {
  Query: {
    repo: async (_parent, {args: {id}}, {pachClient}) => {
      return repoInfoToGQLRepo(await pachClient.pfs().inspectRepo(id));
    },
  },
  Repo: {
    commits: async (repo, _args, {pachClient}) => {
      let jobs: JobInfo.AsObject[] = [];

      try {
        jobs = await getJobs(repo.id, pachClient);
      } catch (err) {
        jobs = [];
      }

      try {
        return (
          await pachClient.pfs().listCommit({
            repo: {name: repo.id},
            number: NUM_COMMITS,
            reverse: true,
          })
        )
          .map((commit) => ({
            repoName: repo.name,
            branch: commit.commit?.branch && {
              id: commit.commit?.branch?.name,
              name: commit.commit?.branch?.name,
            },
            description: commit.description,
            finished: commit.finished?.seconds || 0,
            hasLinkedJob: jobs.some((job) => job.job?.id === commit.commit?.id),
            id: commit.commit?.id || '',
            started: commit.started?.seconds || 0,
            sizeBytes: getSizeBytes(commit),
            sizeDisplay: formatBytes(getSizeBytes(commit)),
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
