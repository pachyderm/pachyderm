import keyBy from 'lodash/keyBy';
import {useMemo} from 'react';

import {UUID_WITHOUT_DASHES_REGEX} from '@dash-frontend/constants/pachCore';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import hasRepoReadPermissions from '@dash-frontend/lib/hasRepoReadPermissions';

import {useJobSet} from './useJobSet';
import {usePipelines} from './usePipelines';
import {useRepos} from './useRepos';

export const useSearchResults = (
  projectId: string,
  query: string,
  limit?: number,
) => {
  const enabled = !!query;
  const lowercaseQuery = query?.toLowerCase();

  const {
    repos,
    loading: reposLoading,
    error: reposError,
  } = useRepos(
    projectId,
    enabled && !UUID_WITHOUT_DASHES_REGEX.test(lowercaseQuery),
  );
  const {
    pipelines,
    loading: pipelinesLoading,
    error: pipelinesError,
  } = usePipelines(
    projectId,
    enabled && !UUID_WITHOUT_DASHES_REGEX.test(lowercaseQuery),
  );

  const {
    jobSet,
    loading: jobSetLoading,
    error: jobSetError,
  } = useJobSet(query, enabled && UUID_WITHOUT_DASHES_REGEX.test(query));

  const filteredRepos = useMemo(
    () =>
      repos?.filter(
        (r) =>
          r.repo?.name?.toLowerCase().startsWith(lowercaseQuery) &&
          hasRepoReadPermissions(r.authInfo?.permissions),
      ),
    [lowercaseQuery, repos],
  );

  const filteredPipelines = useMemo(() => {
    const authorizedRepoMap = keyBy(filteredRepos, (r) => r.repo?.name || '');
    const filteredPipelines = pipelines?.filter(
      (p) =>
        authorizedRepoMap[p.pipeline?.name || ''] &&
        p.pipeline?.name?.toLowerCase().startsWith(lowercaseQuery),
    );
    return filteredPipelines;
  }, [filteredRepos, pipelines, lowercaseQuery]);

  return {
    loading: reposLoading || pipelinesLoading || jobSetLoading,
    error:
      getErrorMessage(reposError) ||
      getErrorMessage(pipelinesError) ||
      getErrorMessage(jobSetError),
    searchResults: {
      pipelines: filteredPipelines?.slice(0, limit) ?? [],
      repos: filteredRepos?.slice(0, limit) ?? [],
      jobSet,
    },
  };
};
