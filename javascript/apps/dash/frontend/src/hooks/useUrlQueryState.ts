import {useCallback, useMemo} from 'react';
import {useLocation} from 'react-router';

import {DagDirection} from '@graphqlTypes';

interface UrlState {
  dagDirection?: DagDirection;
}

const useUrlQueryState = () => {
  const {pathname, search} = useLocation();

  const searchParams = useMemo(() => {
    return new URLSearchParams(search);
  }, [search]);

  const encodedViewState = searchParams.get('view');

  const viewState = useMemo(() => {
    let decodedViewState: UrlState = {};

    if (encodedViewState) {
      try {
        decodedViewState = JSON.parse(atob(encodedViewState));
      } catch (e) {
        decodedViewState = {};
      }
    }

    return decodedViewState;
  }, [encodedViewState]);

  const setUrlFromViewState = useCallback(
    (newState: Partial<UrlState>) => {
      searchParams.set(
        'view',
        btoa(JSON.stringify({...viewState, ...newState})),
      );

      return `${pathname}?${searchParams}`;
    },
    [pathname, searchParams, viewState],
  );

  return {
    viewState,
    setUrlFromViewState,
  };
};

export default useUrlQueryState;
