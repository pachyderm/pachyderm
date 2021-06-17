import {useEffect, useReducer} from 'react';

type FetchState = {
  loading: boolean;
  error?: string;
  data?: unknown;
};

type FetchAction =
  | {type: 'FETCHING'}
  | {type: 'FETCHED'; payload: unknown}
  | {type: 'ERROR'; payload: string};

type useFetchParams = {
  url: string;
  fetchFunc?: (
    input: RequestInfo,
    init?: RequestInit | undefined,
  ) => Promise<Response>;
  formatResponse?: (data: Response) => Promise<unknown>;
};

const initialState: FetchState = {
  loading: false,
  data: {},
};

export const useFetch = ({
  url,
  fetchFunc = fetch,
  formatResponse = async (res) => res,
}: useFetchParams) => {
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
          const data = await formatResponse(res);
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
