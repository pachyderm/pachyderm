import {toProtoCommitOrigin} from '@dash-backend/lib/gqlEnumMappers';
import {PachClient} from '@dash-backend/lib/types';
import {QueryResolvers} from '@graphqlTypes';

import {commitInfoToGQLCommit} from './builders/pfs';

interface CommitResolver {
  Query: {
    commits: QueryResolvers['commits'];
  };
}

const getJobs = (pipelineId: string, limit: number, pachClient: PachClient) => {
  try {
    return pachClient.pps().listJobs({pipelineId, limit});
  } catch (err) {
    return [];
  }
};

const commitResolver: CommitResolver = {
  Query: {
    commits: async (
      _parent,
      {args: {repoName, branchName, number, desc, originKind}},
      {pachClient},
    ) => {
      const jobs = await getJobs(repoName, number || 100, pachClient);
      return (
        await pachClient.pfs().listCommit({
          repo: {
            name: repoName,
          },
          number: number || 100,
          reverse: desc === false ? false : true,
          originKind: originKind ? toProtoCommitOrigin(originKind) : undefined,
          to: branchName
            ? {
                id: '',
                branch: {
                  name: branchName,
                },
              }
            : undefined,
        })
      ).map((c) => {
        const gqlCommit = commitInfoToGQLCommit(c);
        gqlCommit.hasLinkedJob = jobs.some(
          (jobInfo) => jobInfo.job?.id === c.commit?.id,
        );
        return gqlCommit;
      });
    },
  },
};

export default commitResolver;
