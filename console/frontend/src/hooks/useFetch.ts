import {useEffect} from 'react';

import {useLazyFetch, useFetchParams} from './useLazyFetch';

export const useFetch = <ResponseType>(
  fetchProps: useFetchParams<ResponseType>,
) => {
  const [exec, state] = useLazyFetch(fetchProps);

  useEffect(() => {
    if (fetchProps.skip) return;

    exec();
  }, [exec, fetchProps.skip]);

  return state;
};
