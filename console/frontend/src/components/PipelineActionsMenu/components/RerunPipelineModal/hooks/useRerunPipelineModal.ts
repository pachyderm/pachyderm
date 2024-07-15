import {useCallback} from 'react';
import {useForm} from 'react-hook-form';

import {useRerunPipeline} from '@dash-frontend/hooks/useRerunPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useNotificationBanner} from '@pachyderm/components';

interface FormValues {
  reprocess?: string;
}

const useRerunPipelineModal = (pipelineId: string, onHide: () => void) => {
  const {projectId} = useUrlState();
  const {add: triggerNotification} = useNotificationBanner();

  const onCompleted = () => {
    onHide();
    triggerNotification(`Pipeline "${pipelineId}" was rerun successfully`);
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
      rerunPipeline({
        pipeline: {
          name: pipelineId,
          project: {
            name: projectId,
          },
        },
        reprocess: reprocessFormValue === 'true',
      });
    } catch (e) {
      return;
    }
  }, [pipelineId, rerunPipeline, projectId, reprocessFormValue]);

  return {
    projectId,
    formCtx,
    onRerunPipeline,
    updating,
    disabled: !isFormComplete,
    error,
  };
};

export default useRerunPipelineModal;
