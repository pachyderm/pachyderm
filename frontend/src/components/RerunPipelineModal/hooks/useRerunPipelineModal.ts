import {useCallback, useState} from 'react';
import {useForm} from 'react-hook-form';

import {useRerunPipeline} from '@dash-frontend/hooks/useRerunPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';

interface FormValues {
  reprocess?: string;
}

const useRerunPipelineModal = (pipelineId: string, onHide: () => void) => {
  const {projectId} = useUrlState();
  const [successMessage, setSuccessMessage] = useState('');

  const onCompleted = () => {
    setSuccessMessage('Pipeline was rerun successfully');
    setTimeout(onHide, 1000);
  };

  const formCtx = useForm<FormValues>({mode: 'onChange'});
  const reprocessFormValue = formCtx.getValues('reprocess');
  const isFormComplete = Boolean(reprocessFormValue);

  const {
    rerunPipeline,
    loading: updating,
    error,
  } = useRerunPipeline(onCompleted);

  const onRerunPipeline = useCallback(async () => {
    try {
      await rerunPipeline({
        variables: {
          args: {
            projectId,
            pipelineId,
            reprocess: reprocessFormValue === 'true',
          },
        },
      });
    } catch (e) {
      return;
    }
  }, [rerunPipeline, projectId, pipelineId, reprocessFormValue]);

  return {
    projectId,
    formCtx,
    onRerunPipeline,
    updating,
    disabled: !!successMessage || !isFormComplete,
    error,
    successMessage,
  };
};

export default useRerunPipelineModal;
