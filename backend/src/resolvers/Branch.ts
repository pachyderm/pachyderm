import {ApolloError} from 'apollo-server-errors';

import {MutationResolvers, QueryResolvers} from '@graphqlTypes';

import {branchInfoToGQLBranch} from './builders/pps';

interface BranchResolver {
  Query: {
    branch: QueryResolvers['branch'];
    branches: QueryResolvers['branches'];
  };
  Mutation: {
    createBranch: MutationResolvers['createBranch'];
  };
}

const branchResolver: BranchResolver = {
  Query: {
    branch: async (_field, {args: {projectId, branch}}, {pachClient}) => {
      const branchInfo = await pachClient.pfs().inspectBranch({
        projectId,
        name: branch.name,
        repo: branch.repo || undefined,
      });

      if (branchInfo.branch) {
        return branchInfoToGQLBranch(branchInfo.branch);
      } else {
        throw new ApolloError(`branch not found`);
      }
    },
    branches: async (_field, {args: {projectId, repoName}}, {pachClient}) => {
      const branches = await pachClient.pfs().listBranch({repoName, projectId});
      return branches.map((branch) => {
        if (branch.branch) {
          return branchInfoToGQLBranch(branch.branch);
        } else {
          throw new ApolloError(`branch not found`);
        }
      });
    },
  },
  Mutation: {
    createBranch: async (
      _field,
      {args: {projectId, head, branch, provenance, newCommitSet}},
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
        const headObject = head
          ? {
              id: head?.id,
              branch: head?.branch
                ? {
                    name: head.branch.name,
                    repo: head.branch.repo,
                  }
                : undefined,
            }
          : undefined;

        await pachClient.pfs().createBranch({
          head: headObject,
          branch: branchObject || undefined,
          provenance: provenanceList,
          newCommitSet: newCommitSet || false,
        });

        const created = await pachClient
          .pfs()
          .inspectBranch({projectId, ...branchObject});
        if (created.branch) {
          return branchInfoToGQLBranch(created.branch);
        }
      }
      throw new ApolloError(`failed to create branch`);
    },
  },
};

export default branchResolver;
