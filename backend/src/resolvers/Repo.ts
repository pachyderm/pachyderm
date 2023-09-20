import getJobsFromJobSet from '@dash-backend/lib/getJobsFromJobSet';
import {JobInfo} from '@dash-backend/proto';
import {MutationResolvers, QueryResolvers, RepoResolvers} from '@graphqlTypes';

import {commitInfoToGQLCommit} from './builders/pfs';
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
    repo: async (_parent, {args: {projectId, id}}, {pachClient}) => {
      return repoInfoToGQLRepo(
        await pachClient.pfs.inspectRepo({projectId, name: id}),
      );
    },
    repos: async (_parent, {args: {projectId, jobSetId}}, {pachClient}) => {
      if (jobSetId) {
        const jobs = await getJobsFromJobSet({
          jobSet: await pachClient.pps.inspectJobSet({id: jobSetId, projectId}),
          projectId,
          pachClient,
        });

        const jobPipelines = jobs.reduce((acc, job) => {
          if (job.job?.pipeline?.name) {
            return {...acc, [job.job?.pipeline?.name]: job};
          }
          return acc;
        }, {} as {[key: string]: JobInfo.AsObject});

        return (await pachClient.pfs.listRepo({projectIds: [projectId]}))
          .filter((repo) => repo.repo?.name && jobPipelines[repo.repo?.name])
          .map((repo) => repoInfoToGQLRepo(repo));
      }

      return (await pachClient.pfs.listRepo({projectIds: [projectId]})).map(
        (repo) => repoInfoToGQLRepo(repo),
      );
    },
  },
  Repo: {
    linkedPipeline: async (repo, _args, {pachClient}) => {
      try {
        return pipelineInfoToGQLPipeline(
          await pachClient.pps.inspectPipeline({
            pipelineId: repo.id,
            projectId: repo.projectId,
          }),
        );
      } catch (err) {
        return null;
      }
    },
    lastCommit: async (repo, _args, {pachClient}) => {
      try {
        const commits = await pachClient.pfs.listCommit({
          repo: {
            name: repo.name,
          },
          projectId: repo.projectId,
          number: 1,
        });

        return commitInfoToGQLCommit(commits[0]);
      } catch (err) {
        return null;
      }
    },
  },
  Mutation: {
    createRepo: async (
      _parent,
      {args: {projectId, name, description, update}},
      {pachClient},
    ) => {
      await pachClient.pfs.createRepo({
        projectId,
        repo: {name},
        description: description || undefined,
        update: update || false,
      });
      const repo = await pachClient.pfs.inspectRepo({projectId, name});

      return repoInfoToGQLRepo(repo);
    },
    deleteRepo: async (
      _parent,
      {args: {projectId, repo, force}},
      {pachClient},
    ) => {
      const response = await pachClient.pfs.deleteRepo({
        projectId,
        repo: {name: repo.name},
        force: force || undefined,
      });

      return response.deleted;
    },
  },
};

export default repoResolver;
