import {useCallback} from 'react';
import {useHistory} from 'react-router';

import {useCurrentPipeline} from '@dash-frontend/hooks/useCurrentPipeline';
import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useModal} from '@pachyderm/components';

const useDatumViewer = (onCloseRoute: string) => {
  const {closeModal, isOpen} = useModal(true);
  const {projectId, pipelineId, jobId, datumId} = useUrlState();
  const {updateSearchParamsAndGo} = useUrlQueryState();
  const browserHistory = useHistory();
  const {isServiceOrSpout} = useCurrentPipeline();

  const {
    job,
    loading: jobLoading,
    error,
  } = useJob(
    {
      id: jobId,
      pipelineName: pipelineId,
      projectId,
    },
    !!jobId,
  );

  const onClose = useCallback(() => {
    closeModal();
    setTimeout(() => browserHistory.push(onCloseRoute), 500);
    updateSearchParamsAndGo({datumFilters: []});
  }, [browserHistory, closeModal, onCloseRoute, updateSearchParamsAndGo]);

  return {
    onClose,
    isOpen,
    pipelineId,
    jobId,
    datumId,
    job,
    jobLoading,
    error,
    isServiceOrSpout,
  };
};

export default useDatumViewer;
