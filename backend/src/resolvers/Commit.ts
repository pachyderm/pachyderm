import {toProtoCommitOrigin} from '@dash-backend/lib/gqlEnumMappers';
import {PachClient} from '@dash-backend/lib/types';
import {QueryResolvers} from '@graphqlTypes';

import {commitInfoToGQLCommit} from './builders/pfs';

interface CommitResolver {
  Query: {
    commits: QueryResolvers['commits'];
  };
}

const getJobSetIds = async (pachClient: PachClient) => {
  try {
    const jobSets = await pachClient.pps().listJobSets({details: false});
    return jobSets.reduce((memo, {jobSet}) => {
      memo.add(jobSet?.id);
      return memo;
    }, new Set());
  } catch (err) {
    return Promise.resolve(new Set());
  }
};

const commitResolver: CommitResolver = {
  Query: {
    commits: async (
      _parent,
      {args: {repoName, branchName, number, originKind, pipelineName}},
      {pachClient},
    ) => {
      const jobSetIds = await getJobSetIds(pachClient);
      return (
        await pachClient.pfs().listCommit({
          repo: {
            name: repoName,
          },
          number: number || 100,
          originKind: originKind ? toProtoCommitOrigin(originKind) : undefined,
          to: branchName
            ? {
                id: '',
                branch: {
                  name: branchName,
                  repo: {
                    name: repoName,
                  },
                },
              }
            : undefined,
        })
      ).map((c) => {
        const gqlCommit = commitInfoToGQLCommit(c);
        gqlCommit.hasLinkedJob = c.commit?.id
          ? jobSetIds.has(c.commit?.id)
          : false;
        return gqlCommit;
      });
    },
  },
};

export default commitResolver;
