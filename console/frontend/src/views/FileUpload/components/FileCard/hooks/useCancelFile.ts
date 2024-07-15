import {useCallback, useEffect} from 'react';

import {useLazyFetch} from '@dash-frontend/hooks/useLazyFetch';

type UseCancelFileProps = {
  uploadId?: string;
  onCancel: () => void;
  onError: () => void;
  fileId: string;
};

const useCancelFile = ({
  onCancel,
  fileId,
  uploadId,
  onError,
}: UseCancelFileProps) => {
  const [execFetch, {data: cancelData, loading}] = useLazyFetch({
    url: '/upload/cancel',
    method: 'POST',
    formatResponse: async (data) => await data.json(),
    onError,
  });

  const cancelFile = useCallback(
    /* This is a little odd, we need to allow an argument in case the 
      upload is still happening because the recursive upload won't be aware
      of the state update */
    () => {
      execFetch({
        body: JSON.stringify({
          uploadId,
          fileId,
        }),
      });
    },
    [uploadId, fileId, execFetch],
  );

  useEffect(() => {
    if (cancelData) {
      onCancel();
    }
  }, [cancelData, onCancel]);

  return {
    cancelFile,
    loading,
  };
};

export default useCancelFile;
