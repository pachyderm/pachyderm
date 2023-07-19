import {QueryHookOptions} from '@apollo/client';
import {GetLogsQuery, GetLogsQueryVariables} from '@graphqlTypes';

import {useGetLogsQuery} from '@dash-frontend/generated/hooks';

const useLogs = (
  baseOptions: QueryHookOptions<GetLogsQuery, GetLogsQueryVariables>,
) => {
  const logsQuery = useGetLogsQuery({
    ...baseOptions,
  });
  return {
    ...logsQuery,
    logs: logsQuery.data?.logs.items || [],
    cursor: logsQuery.data?.logs.cursor,
  };
};

export default useLogs;
