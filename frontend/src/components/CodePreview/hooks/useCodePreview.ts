import {useCallback} from 'react';

import {useFetch} from 'hooks/useFetch';

const useCodePreview = (url = '', source = '') => {
  const formatResponse = useCallback(async (res: Response) => {
    return await res.text();
  }, []);
  const {data, loading, error, reset} = useFetch({
    url,
    formatResponse,
    skip: !!source,
  });

  return {
    data: source || data,
    loading: source ? false : loading,
    error: source ? undefined : error,
    reset,
  };
};

export default useCodePreview;
