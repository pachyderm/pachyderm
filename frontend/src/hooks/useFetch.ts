import {useEffect} from 'react';

import {useLazyFetch, useFetchParams} from './useLazyFetch';

export const useFetch = <ResponseType>(
  fetchProps: useFetchParams<ResponseType>,
) => {
  const [exec, state] = useLazyFetch(fetchProps);

  useEffect(() => {
    exec();
  }, [exec]);

  return state;
};
