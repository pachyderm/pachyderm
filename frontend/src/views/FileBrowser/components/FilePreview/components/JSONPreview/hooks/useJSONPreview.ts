import {useCallback} from 'react';

import {useFetch} from 'hooks/useFetch';

const useJSONPreview = (downloadLink: string) => {
  const formatResponse = useCallback(
    async (res: Response) => await res.json(),
    [],
  );
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

export default useJSONPreview;
