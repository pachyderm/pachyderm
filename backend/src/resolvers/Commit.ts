import {ApolloError} from 'apollo-server-errors';

import {
  DEFAULT_COMMITS_LIMIT,
  DEFAULT_FIND_COMMITS_LIMIT,
} from '@dash-backend/constants/limits';
import {UUID_WITHOUT_DASHES_REGEX} from '@dash-backend/constants/pachCore';
import formatDiff, {formatDiffOnlyTotals} from '@dash-backend/lib/formatDiff';
import {toProtoCommitOrigin} from '@dash-backend/lib/gqlEnumMappers';
import {NotFoundError} from '@dash-backend/lib/types';
import {CommitState} from '@dash-backend/proto';
import {MutationResolvers, QueryResolvers, Diff, Commit} from '@graphqlTypes';

import {commitInfoToGQLCommit, commitToGQLCommit} from './builders/pfs';
interface CommitResolver {
  Query: {
    commits: QueryResolvers['commits'];
    commit: QueryResolvers['commit'];
    commitDiff: QueryResolvers['commitDiff'];
    commitSearch: QueryResolvers['commitSearch'];
    findCommits: QueryResolvers['findCommits'];
  };
  Mutation: {
    startCommit: MutationResolvers['startCommit'];
    finishCommit: MutationResolvers['finishCommit'];
  };
}

const commitResolver: CommitResolver = {
  Query: {
    commit: async (
      _parent,
      {args: {projectId, id, repoName, branchName}},
      {pachClient},
    ) => {
      let commit: Commit;

      if (!id) {
        try {
          const commits = await pachClient.pfs.listCommit({
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
          await pachClient.pfs.inspectCommit({
            projectId,
            wait: CommitState.COMMIT_STATE_UNKNOWN,
            commit: {
              id,
              branch: {name: branchName || 'master', repo: {name: repoName}},
            },
          }),
        );
      }

      return {
        ...commit,
      };
    },
    commitDiff: async (
      _parent,
      {args: {projectId, commitId, repoName, branchName}},
      {pachClient},
    ) => {
      let diff: Diff | null = null;

      if (commitId && repoName && branchName) {
        const diffResponse = await pachClient.pfs.diffFile({
          projectId,
          newFileObject: {
            commitId: commitId,
            path: '/',
            branch: {
              name: branchName || 'master',
              repo: {name: repoName},
            },
          },
        });
        diff = formatDiff(diffResponse).diff;
      }

      return diff;
    },
    commits: async (
      _parent,
      {
        args: {
          projectId,
          repoName,
          branchName,
          commitIdCursor,
          number,
          originKind,
          cursor,
          reverse,
        },
      },
      {pachClient},
    ) => {
      number = number || DEFAULT_COMMITS_LIMIT;

      // You can only use one cursor in a query at a time.
      if (cursor && commitIdCursor) {
        throw new ApolloError(
          `received cursor and commitIdCursor arguments`,
          'INVALID_ARGUMENT',
        );
      }
      // The branch parameter only works when not using the cursor argument
      if (cursor && branchName) {
        throw new ApolloError(
          `can not specify cursor and branchName in query`,
          'INVALID_ARGUMENT',
        );
      }

      const commits = await pachClient.pfs.listCommit({
        projectId,
        repo: {
          name: repoName,
          project: {name: projectId},
        },
        number: number + 1,
        originKind: originKind ? toProtoCommitOrigin(originKind) : undefined,
        to:
          commitIdCursor || branchName
            ? {
                id: commitIdCursor || '',
                branch: {
                  name: branchName || '',
                  repo: {
                    name: repoName,
                    project: {name: projectId},
                  },
                },
              }
            : undefined,
        started_time: cursor || undefined,
        reverse: !!reverse,
      });

      let nextCursor = undefined;

      //If commits.length is not greater than limit there are no pages left
      if (commits.length > number) {
        commits.pop(); //remove the extra commit from the response
        nextCursor = commits[commits.length - 1];
      }

      return {
        items: commits.map((c) => commitInfoToGQLCommit(c)),
        cursor: nextCursor && nextCursor.started,
        parentCommit: nextCursor && nextCursor.parentCommit?.id,
      };
    },
    commitSearch: async (
      _parent,
      {args: {projectId, id, repoName}},
      {pachClient},
    ) => {
      if (!UUID_WITHOUT_DASHES_REGEX.test(id)) {
        throw new ApolloError(`invalid commit id`, 'INVALID_ARGUMENT');
      }
      try {
        const commit = await pachClient.pfs.inspectCommit({
          projectId,
          wait: CommitState.COMMIT_STATE_UNKNOWN,
          // We want to specify the repo name but not the branch name
          // as we are searching with only commit id.
          commit: {
            id,
            branch: {name: '', repo: {name: repoName}},
          },
        });

        return commitInfoToGQLCommit(commit);
      } catch (e) {
        if (e instanceof NotFoundError) {
          return null;
        }
        throw e;
      }
    },
    findCommits: async (
      _parent,
      {args: {projectId, repoId, commitId, branchId, filePath, limit}},
      {pachClient},
    ) => {
      if (!commitId && !branchId) {
        throw new ApolloError(
          `Need to specify commitId and or branchId`,
          'INVALID_ARGUMENT',
        );
      }

      const commits = await pachClient.pfs.findCommits({
        commit: {
          id: commitId || '',
          branch: {
            name: branchId || '',
            repo: {name: repoId, project: {name: projectId}},
          },
        },
        path: filePath,
        limit: limit || DEFAULT_FIND_COMMITS_LIMIT,
      });

      const foundCommitIds: string[] = [];
      let lastCommitSearched = '';

      commits.forEach((c) => {
        if (c.foundCommit) {
          foundCommitIds.push(c.foundCommit.id);
        } else if (c.lastSearchedCommit) {
          lastCommitSearched = c.lastSearchedCommit.id;
        }
      });

      const foundCommits = await Promise.all(
        foundCommitIds.map(async (commit) => {
          const commitInfo = commitInfoToGQLCommit(
            await pachClient.pfs.inspectCommit({
              projectId,
              wait: CommitState.COMMIT_STATE_UNKNOWN,
              commit: {
                id: commit,
                branch: {name: '', repo: {name: repoId}},
              },
            }),
          );

          let diffTotals;
          if (commitInfo.finished !== -1) {
            const diffResponse = await pachClient.pfs.diffFile({
              projectId,
              newFileObject: {
                commitId: commitInfo.id || '',
                path: filePath || '/',
                branch: {
                  name: branchId || '',
                  repo: {name: repoId},
                },
              },
            });
            diffTotals = formatDiffOnlyTotals(diffResponse);
          }

          return {
            id: commitInfo.id,
            started: commitInfo.started,
            description: commitInfo.description,
            commitAction: diffTotals && Object.values(diffTotals)[0],
          };
        }),
      );

      let cursor = '';
      if (lastCommitSearched) {
        const lastCommitInfo = await pachClient.pfs.inspectCommit({
          projectId,
          wait: CommitState.COMMIT_STATE_UNKNOWN,
          commit: {
            id: lastCommitSearched,
            branch: {name: '', repo: {name: repoId}},
          },
        });
        cursor = lastCommitInfo.parentCommit?.id || '';
      }

      return {
        commits: foundCommits,
        cursor: cursor,
        hasNextPage: !!cursor,
      };
    },
  },
  Mutation: {
    startCommit: async (
      _field,
      {args: {projectId, branchName, repoName}},
      {pachClient},
    ) => {
      const commit = await pachClient.pfs.startCommit({
        projectId,
        branch: {repo: {name: repoName}, name: branchName},
      });

      return commitToGQLCommit(commit);
    },
    finishCommit: async (_field, {args: {projectId, commit}}, {pachClient}) => {
      await pachClient.pfs.finishCommit({projectId, commit});
      return true;
    },
  },
};

export default commitResolver;
