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

const getJobSetIds = async ({
  projectId,
  pachClient,
}: {
  projectId: string;
  pachClient: PachClient;
}) => {
  try {
    const jobSets = await pachClient
      .pps()
      .listJobSets({projectIds: [projectId], details: false});
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
      {args: {projectId, id, repoName, branchName, withDiff}},
      {pachClient},
    ) => {
      let diff = undefined;
      const jobSetIds = await getJobSetIds({pachClient, projectId});

      const commit = commitInfoToGQLCommit(
        await pachClient.pfs().inspectCommit({
          projectId,
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
          projectId,
          newFileObject: {
            commitId: id || 'master',
            path: '/',
            branch: {name: branchName || 'master', repo: {name: repoName}},
          },
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
      {args: {projectId, repoName, branchName, number, originKind}},
      {pachClient},
    ) => {
      const jobSetIds = await getJobSetIds({projectId, pachClient});
      return (
        await pachClient.pfs().listCommit({
          projectId,
          repo: {
            name: repoName,
            project: {name: projectId},
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
                    project: {name: projectId},
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
      {args: {projectId, branchName, repoName}},
      {pachClient},
    ) => {
      const commit = await pachClient.pfs().startCommit({
        projectId,
        branch: {repo: {name: repoName}, name: branchName},
      });

      return commitToGQLCommit(commit);
    },
    finishCommit: async (_field, {args: {projectId, commit}}, {pachClient}) => {
      await pachClient.pfs().finishCommit({projectId, commit});
      return true;
    },
  },
};

export default commitResolver;
