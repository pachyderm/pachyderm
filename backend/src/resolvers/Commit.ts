import formatDiff from '@dash-backend/lib/formatDiff';
import {toProtoCommitOrigin} from '@dash-backend/lib/gqlEnumMappers';
import {PachClient} from '@dash-backend/lib/types';
import {CommitState} from '@dash-backend/proto';
import {
  MutationResolvers,
  OriginKind,
  QueryResolvers,
  Diff,
  Commit,
} from '@graphqlTypes';

import {commitInfoToGQLCommit, commitToGQLCommit} from './builders/pfs';

const DEFAULT_LIMIT = 100;
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
      let diff: Diff | undefined;
      let commit: Commit;
      const jobSetIds = await getJobSetIds({pachClient, projectId});

      if (!id) {
        try {
          const commits = await pachClient.pfs().listCommit({
            projectId,
            repo: {
              name: repoName,
            },
            number: 1,
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
          });
          commit = commitInfoToGQLCommit(commits[0]);
        } catch (err) {
          return null;
        }
      } else {
        commit = commitInfoToGQLCommit(
          await pachClient.pfs().inspectCommit({
            projectId,
            wait: CommitState.COMMIT_STATE_UNKNOWN,
            commit: {
              id,
              branch: {name: branchName || 'master', repo: {name: repoName}},
            },
          }),
        );
      }
      if (
        withDiff &&
        commit.originKind !== OriginKind.ALIAS &&
        commit.finished !== -1
      ) {
        const diffResponse = await pachClient.pfs().diffFile({
          projectId,
          newFileObject: {
            commitId: id || commit.id,
            path: '/',
            branch: {
              name: commit.branch?.name || 'master',
              repo: {name: repoName},
            },
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
      {args: {projectId, repoName, branchName, number, originKind, cursor}},
      {pachClient},
    ) => {
      number = number || DEFAULT_LIMIT;

      const jobSetIds = await getJobSetIds({projectId, pachClient});
      const commits = await pachClient.pfs().listCommit({
        projectId,
        repo: {
          name: repoName,
          project: {name: projectId},
        },
        number: number + 1,
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
        started_time: cursor || undefined,
      });

      let nextCursor = undefined;

      //If commits.length is not greater than limit there are no pages left
      if (commits.length > number) {
        commits.pop(); //remove the extra commit from the response
        nextCursor = commits[commits.length - 1].started;
      }

      return {
        items: commits.map((c) => {
          const gqlCommit = commitInfoToGQLCommit(c);
          gqlCommit.hasLinkedJob = c.commit?.id
            ? jobSetIds.has(c.commit?.id)
            : false;
          return gqlCommit;
        }),
        cursor: nextCursor,
        hasNextPage: !!nextCursor,
      };
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
