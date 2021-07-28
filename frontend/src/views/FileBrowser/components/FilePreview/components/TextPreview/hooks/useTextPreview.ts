import {useCallback, useMemo} from 'react';

import {useFetch} from 'hooks/useFetch';

const useJSONPreview = (downloadLink: string) => {
  const formatResponse = useCallback(async (res: Response) => res.text(), []);

  const {data, loading, error} = useFetch<string>({
    url: downloadLink,
    formatResponse,
  });

  const formattedData = useMemo(() => {
    // use leading empty string as UI padding on index 0
    return ['', ...String(data).split('\n')];
  }, [data]);

  return {
    data: formattedData,
    loading,
    error,
  };
};

export default useJSONPreview;
