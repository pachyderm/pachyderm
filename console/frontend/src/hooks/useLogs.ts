import {useQuery} from '@tanstack/react-query';

import {GetLogsRequest, LogMessage, getLogs} from '@dash-frontend/api/pps';
import {isErrorWithMessage, isUnknown} from '@dash-frontend/api/utils/error';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

/**
 * The provided `since` should be an ISO string. `since` for the network request is a protobuf duration.
 * The provided `since` ISO string will be converted to the correct duration for the network request.
 */
export const useLogs = (
  req: Omit<GetLogsRequest, 'since'>,
  {
    cursor,
    limit,
    enabled = true,
    refetchInterval = false,
    since,
  }: {
    since?: string;
    cursor?: LogMessage;
    limit?: number;
    enabled?: boolean;
    refetchInterval?: false | number;
  },
) => {
  const isOverLokiQueryLimit = (error: unknown) =>
    isUnknown(error) &&
    isErrorWithMessage(error) &&
    error?.message.includes('the query time range exceeds the limit');

  const {
    data,
    isLoading: loading,
    error,
    refetch,
  } = useQuery({
    queryKey: queryKeys.log({
      projectName: req.pipeline?.project?.name || '',
      pipelineName: req.pipeline?.name || '',
      jobId: req.job?.id || '',
      datumId: req.datum?.id || '',
      req,
      args: {
        cursor,
        limit,
        since,
      },
    }),
    queryFn: () => getLogs(req, {cursor, limit, since}),
    throwOnError: (e) =>
      // This error will occur when you request a time frame longer than loki is storing
      !isOverLokiQueryLimit(e),
    enabled,
    refetchInterval,
  });

  const isPagingError = error?.name === 'PagingError';

  return {
    loading,
    logs: data,
    error: getErrorMessage(error),
    refetch,
    isOverLokiQueryLimit: isOverLokiQueryLimit(error),
    isPagingError,
  };
};
