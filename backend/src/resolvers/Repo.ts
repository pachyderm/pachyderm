import {JobInfo} from '@pachyderm/node-pachyderm';

import getJobsFromJobSet from '@dash-backend/lib/getJobsFromJobSet';
import {MutationResolvers, QueryResolvers, RepoResolvers} from '@graphqlTypes';

import {pipelineInfoToGQLPipeline, repoInfoToGQLRepo} from './builders/pps';

interface RepoResolver {
  Query: {
    repo: QueryResolvers['repo'];
    repos: QueryResolvers['repos'];
  };
  Repo: RepoResolvers;
  Mutation: {
    createRepo: MutationResolvers['createRepo'];
    deleteRepo: MutationResolvers['deleteRepo'];
  };
}

const repoResolver: RepoResolver = {
  Query: {
    repo: async (_parent, {args: {id}}, {pachClient}) => {
      return repoInfoToGQLRepo(await pachClient.pfs().inspectRepo(id));
    },
    repos: async (_parent, {args: {projectId, jobSetId}}, {pachClient}) => {
      if (jobSetId) {
        const jobs = await getJobsFromJobSet({
          jobSet: await pachClient
            .pps()
            .inspectJobSet({id: jobSetId, projectId}),
          projectId,
          pachClient,
        });

        const jobPipelines = jobs.reduce((acc, job) => {
          if (job.job?.pipeline?.name) {
            return {...acc, [job.job?.pipeline?.name]: job};
          }
          return acc;
        }, {} as {[key: string]: JobInfo.AsObject});

        return (await pachClient.pfs().listRepo())
          .filter((repo) => repo.repo?.name && jobPipelines[repo.repo?.name])
          .map((repo) => repoInfoToGQLRepo(repo));
      }

      return (await pachClient.pfs().listRepo()).map((repo) =>
        repoInfoToGQLRepo(repo),
      );
    },
  },
  Repo: {
    linkedPipeline: async (repo, _args, {pachClient}) => {
      try {
        if (repo.linkedPipeline) {
          return pipelineInfoToGQLPipeline(
            await pachClient.pps().inspectPipeline(repo.id),
          );
        }
        return null;
      } catch (err) {
        return null;
      }
    },
  },
  Mutation: {
    createRepo: async (
      _parent,
      {args: {name, description, update}},
      {pachClient},
    ) => {
      await pachClient.pfs().createRepo({
        repo: {name},
        description: description || undefined,
        update: update || false,
      });
      const repo = await pachClient.pfs().inspectRepo(name);

      return repoInfoToGQLRepo(repo);
    },
    deleteRepo: async (_parent, {args: {repo, force}}, {pachClient}) => {
      await pachClient.pfs().deleteRepo({
        repo: {name: repo.name},
        force: force || undefined,
      });

      return true;
    },
  },
};

export default repoResolver;
