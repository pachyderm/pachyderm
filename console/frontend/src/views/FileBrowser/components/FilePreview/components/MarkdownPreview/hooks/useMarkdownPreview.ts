import {useCallback} from 'react';

import {useFetch} from 'hooks/useFetch';

const useMarkdownPreview = (url: string) => {
  const formatResponse = useCallback(async (res: Response) => {
    return await res.text();
  }, []);

  return useFetch({
    url,
    formatResponse,
  });
};

export default useMarkdownPreview;
