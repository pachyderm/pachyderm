import formatDiff from '@dash-backend/lib/formatDiff';
import {toProtoCommitOrigin} from '@dash-backend/lib/gqlEnumMappers';
import {PachClient} from '@dash-backend/lib/types';
import {CommitState} from '@dash-backend/proto';
import {MutationResolvers, OriginKind, QueryResolvers} from '@graphqlTypes';

import {commitInfoToGQLCommit, commitToGQLCommit} from './builders/pfs';

interface CommitResolver {
  Query: {
    commits: QueryResolvers['commits'];
    commit: QueryResolvers['commit'];
  };
  Mutation: {
    startCommit: MutationResolvers['startCommit'];
    finishCommit: MutationResolvers['finishCommit'];
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
    commit: async (
      _parent,
      {args: {id, repoName, branchName, withDiff}},
      {pachClient},
    ) => {
      let diff = undefined;
      const jobSetIds = await getJobSetIds(pachClient);

      const commit = commitInfoToGQLCommit(
        await pachClient.pfs().inspectCommit({
          wait: CommitState.COMMIT_STATE_UNKNOWN,
          commit: {
            id,
            branch: {name: branchName || 'master', repo: {name: repoName}},
          },
        }),
      );

      if (
        withDiff &&
        commit.originKind !== OriginKind.ALIAS &&
        commit.finished !== -1
      ) {
        const diffResponse = await pachClient.pfs().diffFile({
          commitId: id || 'master',
          path: '/',
          branch: {name: branchName || 'master', repo: {name: repoName}},
        });
        diff = formatDiff(diffResponse).diff;
      }

      return {
        ...commit,
        diff,
        hasLinkedJob: commit.id ? jobSetIds.has(commit.id) : false,
      };
    },
    commits: async (
      _parent,
      {args: {repoName, branchName, number, originKind}},
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
  Mutation: {
    startCommit: async (
      _field,
      {args: {branchName, repoName}},
      {pachClient},
    ) => {
      const commit = await pachClient
        .pfs()
        .startCommit({branch: {repo: {name: repoName}, name: branchName}});

      return commitToGQLCommit(commit);
    },
    finishCommit: async (_field, {args: {commit}}, {pachClient}) => {
      await pachClient.pfs().finishCommit({commit: commit});
      return true;
    },
  },
};

export default commitResolver;
