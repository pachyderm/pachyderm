import {useCallback, useEffect, useMemo, useState} from 'react';
import {useHistory} from 'react-router';

import {JobState} from '@dash-frontend/api/pps';
import {
  restJobStateToNodeState,
  restPipelineStateToNodeState,
} from '@dash-frontend/api/utils/nodeStateMappers';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {NodeState} from '@dash-frontend/lib/types';
import {pipelineRoute} from '@dash-frontend/views/Project/utils/routes';

import {useSearchResults} from '../../../../../hooks/useSearchResults';

const useDropdown = () => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();
  const [search, setSearch] = useState('');
  const lowercaseQuery = search?.toLowerCase();

  const [selectedTab, setSelectedTab] = useState<
    'job' | 'pipeline' | undefined
  >();

  const pipelineOnClick = useCallback(
    (pipelineName: string, tabId?: string) => {
      browserHistory.push(
        pipelineRoute({
          projectId,
          pipelineId: pipelineName,
          tabId: tabId,
        }),
      );
    },
    [browserHistory, projectId],
  );

  const {
    loading,
    searchResults: {pipelines: searchResults},
  } = useSearchResults(projectId, lowercaseQuery);

  const filteredSubJobs = useMemo(
    () =>
      searchResults?.filter(
        (pipeline) =>
          restJobStateToNodeState(pipeline.lastJobState) === NodeState.ERROR &&
          pipeline.lastJobState !== JobState.JOB_UNRUNNABLE,
      ),
    [searchResults],
  );

  const filteredPipelines = useMemo(
    () =>
      searchResults?.filter(
        (pipeline) =>
          restPipelineStateToNodeState(pipeline.state) === NodeState.ERROR,
      ),
    [searchResults],
  );

  const [initialtab, setInitialTab] = useState('');
  useEffect(() => {
    if (
      !initialtab &&
      searchResults &&
      (filteredSubJobs.length !== 0 || filteredPipelines.length !== 0)
    ) {
      setInitialTab(filteredSubJobs.length > 0 ? 'job' : 'pipeline');
    }
  }, [filteredPipelines, filteredSubJobs, initialtab, searchResults]);

  const subJobsHeading = `${filteredSubJobs.length} Subjob ${
    (filteredSubJobs.length || 0) === 1 ? 'Failed' : 'Failures'
  }`;

  const pipelinesHeading = `${filteredPipelines.length} Pipeline ${
    (filteredPipelines.length || 0) === 1 ? 'Failed' : 'Failures'
  }`;
  return {
    search,
    setSearch,
    setSelectedTab,
    initialtab,
    filteredSubJobs,
    subJobsHeading,
    filteredPipelines,
    pipelinesHeading,
    searchResults: selectedTab === 'job' ? filteredSubJobs : filteredPipelines,
    lowercaseQuery,
    pipelineOnClick,
    loading,
    selectedTab,
  };
};

export default useDropdown;
