import {useEffect, useReducer, useCallback} from 'react';

type FetchState = {
  loading: boolean;
  error?: string;
  data?: unknown;
};

type FetchAction =
  | {type: 'FETCHING'}
  | {type: 'FETCHED'; payload: unknown}
  | {type: 'ERROR'; payload: string};

interface useFetchParams<Type> {
  url: string;
  fetchFunc?: (
    input: RequestInfo,
    init?: RequestInit | undefined,
  ) => Promise<Response>;
  formatResponse?: (data: Response) => Promise<Type>;
}

const initialState: FetchState = {
  loading: false,
  data: {},
};

export const useFetch = <ResponseType>({
  url,
  fetchFunc = fetch,
  formatResponse,
}: useFetchParams<ResponseType>) => {
  const [state, dispatch] = useReducer(
    (state: FetchState, action: FetchAction) => {
      switch (action.type) {
        case 'FETCHING':
          return {...initialState, loading: true};
        case 'FETCHED':
          return {...initialState, data: action.payload};
        case 'ERROR':
          return {...initialState, error: action.payload};
        default:
          return state;
      }
    },
    initialState,
  );

  useEffect(() => {
    const fetchURL = async () => {
      dispatch({type: 'FETCHING'});

      try {
        const res = await fetchFunc(url, {
          credentials: 'include',
          mode: 'cors',
        });

        if (res.ok) {
          const data = formatResponse ? await formatResponse(res) : res;
          dispatch({type: 'FETCHED', payload: data});
        } else {
          const data = await res.json();
          dispatch({type: 'ERROR', payload: data.errors[0].detail});
        }
      } catch (error) {
        dispatch({
          type: 'ERROR',
          payload: 'Something went wrong. Please try again.',
        });
      }
    };
    fetchURL();
  }, [fetchFunc, formatResponse, url]);

  return state;
};
