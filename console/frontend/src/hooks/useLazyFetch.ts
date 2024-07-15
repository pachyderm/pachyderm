import {useReducer, useCallback} from 'react';

type FetchState<ResponseType> = {
  loading: boolean;
  error?: string;
  data?: ResponseType;
};

type FetchAction<ResponseType> =
  | {type: 'FETCHING'}
  | {type: 'FETCHED'; payload: ResponseType}
  | {type: 'ERROR'; payload: string}
  | {type: 'RESET'};

interface fetchParams {
  method?: string;
  body?: BodyInit;
  headers?: Record<string, string>;
}

export interface useFetchParams<Type> extends fetchParams {
  url: string;
  fetchFunc?: (
    input: RequestInfo,
    init?: RequestInit | undefined,
  ) => Promise<Response>;
  formatResponse: (data: Response) => Promise<Type> | Type;
  mode?: RequestMode;
  credentials?: RequestCredentials;
  onComplete?: (data: Type) => void;
  onError?: (err: unknown) => void;
  skip?: boolean;
}

export const useLazyFetch = <ResponseType>({
  url,
  fetchFunc = fetch,
  formatResponse,
  method = 'GET',
  mode = 'cors',
  credentials = 'include',
  body,
  onComplete,
  onError,
}: useFetchParams<ResponseType>): [
  (arg0?: fetchParams) => Promise<void>,
  FetchState<ResponseType> & {reset: () => void},
] => {
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
        case 'RESET':
          return initialState;
        default:
          return state;
      }
    },
    initialState,
  );

  const execFetch = useCallback(
    async (params = {}) => {
      dispatch({type: 'FETCHING'});

      try {
        const res = await fetchFunc(url, {
          credentials,
          mode,
          headers: {
            'Content-Type': 'application/json',
          },
          method,
          body,
          ...params,
        });

        if (res.ok) {
          const data = await formatResponse(res);
          if (onComplete) onComplete(data);
          dispatch({type: 'FETCHED', payload: data});
        } else {
          const data = await res.json();
          dispatch({type: 'ERROR', payload: data.errors[0].detail});
        }
      } catch (error) {
        if (onError) onError(error);
        dispatch({
          type: 'ERROR',
          payload: 'Something went wrong. Please try again.',
        });
      }
    },
    [
      fetchFunc,
      formatResponse,
      url,
      body,
      method,
      credentials,
      mode,
      onError,
      onComplete,
    ],
  );

  const reset = () => dispatch({type: 'RESET'});

  return [execFetch, {...state, reset}];
};
