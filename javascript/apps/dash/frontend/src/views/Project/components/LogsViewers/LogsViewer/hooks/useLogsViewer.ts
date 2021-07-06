import {useModal} from '@pachyderm/components';
import {useCallback, useEffect, useState} from 'react';

import useLogs from '@dash-frontend/hooks/useLogs';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import {LOGS_DEFAULT_DROPDOWN_OPTIONS} from '../../constants/logsViewersConstants';

const timeNowInSeconds = Math.floor(Date.now() / 1000);

export const defaultValues: {[key: string]: number} = {
  'Last 30 Minutes': timeNowInSeconds - 1800,
  'Last 24 Hours': timeNowInSeconds - 86400,
  'Last 3 Days': timeNowInSeconds - 86400 * 3,
};

const useLogsViewer = (
  onCloseCallback: () => void,
  dropdownLabel: string,
  startTime?: number,
) => {
  const {closeModal, isOpen} = useModal(true);
  const {projectId, pipelineId, jobId} = useUrlState();
  const [highlightUserLogs, setHighlightUserLogs] = useState(false);
  const [selectedLogsMap, setSelectedLogsMap] = useState<{
    [key: number]: boolean;
  }>({});

  const [selectedTime, setSelectedTime] = useState(dropdownLabel);

  const dropdownValues: {[key: string]: number | undefined} = {
    [dropdownLabel]: startTime,
    ...defaultValues,
  };

  const dropdownOptions = [
    {
      id: dropdownLabel,
      content: dropdownLabel,
      closeOnClick: true,
    },
    ...LOGS_DEFAULT_DROPDOWN_OPTIONS,
  ];

  const {logs, loading} = useLogs({
    projectId: projectId,
    pipelineName: pipelineId,
    jobId: jobId,
    start: dropdownValues[selectedTime],
    reverse: true,
  });

  useEffect(() => {
    setSelectedLogsMap({});
  }, [selectedTime]);

  const onClose = useCallback(() => {
    closeModal();
    onCloseCallback();
  }, [closeModal, onCloseCallback]);

  return {
    onClose,
    isOpen,
    loading,
    logs,
    highlightUserLogs,
    setHighlightUserLogs,
    selectedLogsMap,
    setSelectedLogsMap,
    selectedTime,
    dropdownOptions,
    setSelectedTime,
  };
};
export default useLogsViewer;
