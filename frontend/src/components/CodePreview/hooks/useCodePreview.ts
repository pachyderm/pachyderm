import {useCallback} from 'react';

import {useFetch} from 'hooks/useFetch';

const useCodePreview = (url?: string, source?: string) => {
  const formatResponse = useCallback(async (res: Response) => {
    return await res.text();
  }, []);
  const {data, loading, error, reset} = useFetch({
    url: url || '',
    formatResponse,
    skip: !url,
  });

  return {
    data: source || data,
    loading: !url ? false : loading,
    error: !url ? undefined : error,
    reset,
  };
};

export default useCodePreview;
