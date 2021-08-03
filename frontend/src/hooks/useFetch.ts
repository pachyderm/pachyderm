import {useEffect, useReducer} from 'react';

type FetchState<ResponseType> = {
  loading: boolean;
  error?: string;
  data?: ResponseType;
};

type FetchAction<ResponseType> =
  | {type: 'FETCHING'}
  | {type: 'FETCHED'; payload: ResponseType}
  | {type: 'ERROR'; payload: string};

interface useFetchParams<Type> {
  url: string;
  fetchFunc?: (
    input: RequestInfo,
    init?: RequestInit | undefined,
  ) => Promise<Response>;
  formatResponse: (data: Response) => Promise<Type> | Type;
}

export const useFetch = <ResponseType>({
  url,
  fetchFunc = fetch,
  formatResponse,
}: useFetchParams<ResponseType>) => {
  const initialState: FetchState<ResponseType> = {
    loading: false,
  };
  const [state, dispatch] = useReducer(
    (state: FetchState<ResponseType>, action: FetchAction<ResponseType>) => {
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
