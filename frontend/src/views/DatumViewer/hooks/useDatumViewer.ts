import {useCallback} from 'react';
import {useHistory} from 'react-router';

import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useModal} from '@pachyderm/components';

const useDatumViewer = (onCloseRoute: string) => {
  const {closeModal, isOpen} = useModal(true);
  const {projectId, pipelineId, jobId, datumId} = useUrlState();
  const {updateSearchParamsAndGo} = useUrlQueryState();
  const browserHistory = useHistory();

  const {
    job,
    loading: jobLoading,
    error,
  } = useJob({
    id: jobId,
    pipelineName: pipelineId,
    projectId,
  });

  const onClose = useCallback(() => {
    closeModal();
    setTimeout(() => browserHistory.push(onCloseRoute), 500);
    updateSearchParamsAndGo({datumFilters: []});
  }, [browserHistory, closeModal, onCloseRoute, updateSearchParamsAndGo]);

  return {
    onClose,
    isOpen,
    jobId,
    datumId,
    job,
    jobLoading,
    error,
  };
};

export default useDatumViewer;
