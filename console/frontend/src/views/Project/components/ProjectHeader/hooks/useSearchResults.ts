import keyBy from 'lodash/keyBy';
import {useMemo} from 'react';

import {PipelineInfo} from '@dash-frontend/api/pps';
import {
  restJobStateToNodeState,
  restPipelineStateToNodeState,
} from '@dash-frontend/api/utils/nodeStateMappers';
import {usePipelines} from '@dash-frontend/hooks/usePipelines';
import {useRepos} from '@dash-frontend/hooks/useRepos';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import hasRepoReadPermissions from '@dash-frontend/lib/hasRepoReadPermissions';
import {NodeState} from '@dash-frontend/lib/types';

export const useSearchResults = (projectId: string, query: string) => {
  const lowercaseQuery = query?.toLowerCase();

  const {
    repos,
    loading: reposLoading,
    error: reposError,
  } = useRepos(projectId, true, 2000);
  const {
    pipelines,
    loading: pipelinesLoading,
    error: pipelinesError,
  } = usePipelines(projectId, true, 2000);

  const filteredRepos = useMemo(() => {
    const filteredRepos = repos?.filter(
      (r) =>
        r.repo?.name?.toLowerCase().startsWith(lowercaseQuery) &&
        hasRepoReadPermissions(r.authInfo?.permissions),
    );

    filteredRepos?.sort(
      (a, b) =>
        a.repo?.name?.localeCompare(b.repo?.name || '', 'en', {
          numeric: true,
        }) || 0,
    );
    return filteredRepos;
  }, [lowercaseQuery, repos]);

  const filteredPipelines = useMemo(() => {
    const authorizedRepoMap = keyBy(filteredRepos, (r) => r.repo?.name || '');
    const pipelinesWithAccessAndNameMatch = pipelines?.filter(
      (p) =>
        authorizedRepoMap[p.pipeline?.name || ''] &&
        p.pipeline?.name?.toLowerCase().startsWith(lowercaseQuery),
    );

    pipelinesWithAccessAndNameMatch?.sort((a, b) => {
      const getRank = (pipeline: PipelineInfo) => {
        const pipelineFailure =
          restPipelineStateToNodeState(pipeline.state) === NodeState.ERROR;
        const subjobFailure =
          restJobStateToNodeState(pipeline.lastJobState) === NodeState.ERROR;

        if (pipelineFailure && subjobFailure) return 3;
        if (pipelineFailure) return 2;
        if (subjobFailure) return 1;
        return 0;
      };

      const rankA = getRank(a);
      const rankB = getRank(b);

      if (rankA !== rankB) {
        return rankB - rankA; // Sort in descending order of rank
      }

      // otherwise sort alphanumerically
      return (
        a.pipeline?.name?.localeCompare(b.pipeline?.name || '', 'en', {
          numeric: true,
        }) || 0
      );
    });

    return pipelinesWithAccessAndNameMatch;
  }, [filteredRepos, pipelines, lowercaseQuery]);

  return {
    loading: reposLoading || pipelinesLoading,
    error: getErrorMessage(reposError) || getErrorMessage(pipelinesError),
    searchResults: {
      pipelines: filteredPipelines || [],
      repos: filteredRepos || [],
    },
  };
};
