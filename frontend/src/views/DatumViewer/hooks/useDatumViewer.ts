import {useCallback} from 'react';
import {useHistory} from 'react-router';

import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useModal} from '@pachyderm/components';

const useDatumViewer = (onCloseRoute: string) => {
  const {closeModal, isOpen} = useModal(true);
  const {jobId, datumId} = useUrlState();
  const {updateViewState} = useUrlQueryState();
  const browserHistory = useHistory();

  const onClose = useCallback(() => {
    closeModal();
    setTimeout(() => browserHistory.push(onCloseRoute), 500);
    updateViewState({datumFilters: []});
  }, [browserHistory, closeModal, onCloseRoute, updateViewState]);

  return {
    onClose,
    isOpen,
    jobId,
    datumId,
  };
};

export default useDatumViewer;
