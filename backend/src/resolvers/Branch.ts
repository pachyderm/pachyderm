import {ApolloError} from 'apollo-server-errors';

import {MutationResolvers, QueryResolvers} from '@graphqlTypes';

import {branchInfoToGQLBranch} from './builders/pps';

interface BranchResolver {
  Query: {
    branch: QueryResolvers['branch'];
  };
  Mutation: {
    createBranch: MutationResolvers['createBranch'];
  };
}

const branchResolver: BranchResolver = {
  Query: {
    branch: async (_field, {args: {branch}}, {pachClient}) => {
      const branchInfo = await pachClient
        .pfs()
        .inspectBranch({name: branch.name, repo: branch.repo || undefined});

      if (branchInfo.branch) {
        return branchInfoToGQLBranch(branchInfo.branch);
      } else {
        throw new ApolloError(`branch not found`);
      }
    },
  },
  Mutation: {
    createBranch: async (
      _field,
      {args: {head, branch, provenance, newCommitSet}},
      {pachClient},
    ) => {
      if (branch) {
        const branchObject = {
          name: branch.name,
          repo: branch.repo || undefined,
        };
        const provenanceList = (provenance || []).map(
          (b) =>
            (b && {
              name: b.name,
              repo: b.repo || undefined,
            }) ||
            undefined,
        );
        await pachClient.pfs().createBranch({
          head: head || undefined,
          branch: branchObject || undefined,
          provenance: provenanceList,
          newCommitSet: newCommitSet || false,
        });

        const created = await pachClient.pfs().inspectBranch(branchObject);
        if (created.branch) {
          return branchInfoToGQLBranch(created.branch);
        }
      }
      throw new ApolloError(`failed to create branch`);
    },
  },
};

export default branchResolver;
