import {QueryFunctionOptions} from '@apollo/client';
import {LogsArgs} from '@graphqlTypes';

import {useGetLogsQuery} from '@dash-frontend/generated/hooks';

const useLogs = (args: LogsArgs, opts?: QueryFunctionOptions) => {
  const {data, error, loading} = useGetLogsQuery({
    variables: {args},
    skip: opts?.skip,
  });

  return {
    error,
    logs: data?.logs || [],
    loading,
  };
};

export default useLogs;
