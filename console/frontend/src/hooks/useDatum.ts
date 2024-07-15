import {useQuery} from '@tanstack/react-query';

import {InspectDatumRequest, inspectDatum} from '@dash-frontend/api/pps';
import {isErrorWithMessage, isUnknown} from '@dash-frontend/api/utils/error';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useDatum = (req: InspectDatumRequest, enabled = true) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.datum({
      projectId: req.datum?.job?.pipeline?.project?.name,
      pipelineId: req.datum?.job?.pipeline?.name,
      jobId: req.datum?.job?.id,
      datumId: req.datum?.id,
    }),
    throwOnError: (e) =>
      !(
        isUnknown(e) &&
        isErrorWithMessage(e) &&
        e?.message.includes('not found in job')
      ),
    queryFn: () => inspectDatum(req),
    enabled,
  });

  return {datum: data, loading, error: getErrorMessage(error)};
};
