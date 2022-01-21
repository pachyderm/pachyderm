import {MutationResolvers, QueryResolvers, RepoResolvers} from '@graphqlTypes';

import {pipelineInfoToGQLPipeline, repoInfoToGQLRepo} from './builders/pps';

interface RepoResolver {
  Query: {
    repo: QueryResolvers['repo'];
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
