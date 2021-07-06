import {useGetLogsQuery} from '@dash-frontend/generated/hooks';
import {LogsArgs} from '@graphqlTypes';

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
