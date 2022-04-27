import {load} from 'js-yaml';
import {useCallback} from 'react';

import {useFetch} from 'hooks/useFetch';

const useYAMLPreview = (downloadLink: string) => {
  const formatResponse = useCallback(async (res: Response) => {
    const body = await res.text();
    return load(body);
  }, []);
  const {data, loading, error} = useFetch({
    url: downloadLink,
    formatResponse,
  });

  return {
    data,
    loading,
    error,
  };
};

export default useYAMLPreview;
