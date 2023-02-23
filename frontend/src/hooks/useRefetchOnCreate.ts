import {useEffect} from 'react';

import usePreviousValue from '@pachyderm/components/hooks/usePreviousValue';

const useRefetchOnCreate = <T extends () => void>({
  refetch,
  loading,
}: {
  refetch: T;
  loading: boolean;
}) => {
  const wasLoading = usePreviousValue(loading);

  useEffect(() => {
    if (wasLoading && !loading) {
      refetch();
    }
  }, [loading, refetch, wasLoading]);
};

export default useRefetchOnCreate;
