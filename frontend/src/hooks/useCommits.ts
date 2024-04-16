import {useQuery} from '@tanstack/react-query';

import {listCommitsPaged} from '@dash-frontend/api/pfs';
import {DEFAULT_COMMITS_LIMIT} from '@dash-frontend/constants/limits';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const COMMIT_LIMIT = 100;

type UseCommitsArgs = {
  branchName?: string;
  cursor?: string;
  commitIdCursor?: string;
  number?: number;
  reverse?: boolean;
};

type UseCommits = {
  projectName: string;
  repoName: string;
  args: UseCommitsArgs;
};

export const useCommits = (
  {args, projectName, repoName}: UseCommits,
  skip = false,
) => {
  const {
    data,
    isLoading: loading,
    error,
    refetch,
    isRefetching: refetching,
  } = useQuery({
    queryKey: queryKeys.commits<UseCommitsArgs>({
      projectId: projectName,
      repoId: repoName,
      args,
    }),
    enabled: !skip,
    queryFn: () => {
      // You can only use one cursor in a query at a time.
      if (args.cursor && args.commitIdCursor) {
        throw new Error(
          `INVALID_ARGUMENT: received cursor and commitIdCursor arguments`,
        );
      }
      // The branch parameter only works when not using the cursor argument
      if (args.cursor && args.branchName) {
        throw new Error(
          `INVALID_ARGUMENT: can not specify cursor and branchName in query`,
        );
      }

      const repo = {
        name: repoName,
        type: 'user',
        project: {name: projectName},
      };

      return listCommitsPaged({
        repo,
        to:
          args.commitIdCursor || args.branchName
            ? {
                id: args.commitIdCursor || '',
                branch: {
                  name: args.branchName || '',
                  repo: repo,
                },
              }
            : undefined,
        startedTime: args.cursor || undefined,
        number: String(args.number ?? DEFAULT_COMMITS_LIMIT),
        reverse: args.reverse,
      });
    },
  });

  return {
    ...data,
    loading,
    error: getErrorMessage(error),
    refetch,
    refetching,
  };
};
