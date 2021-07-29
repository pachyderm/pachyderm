import {LogsArgs} from '@graphqlTypes';

import {useGetLogsQuery} from '@dash-frontend/generated/hooks';

const useLogs = (args: LogsArgs) => {
  const {data, error, loading} = useGetLogsQuery({
    variables: {args},
  });

  return {
    error,
    logs: data?.logs || [],
    loading,
  };
};

export default useLogs;
