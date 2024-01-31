import isEmpty from 'lodash/isEmpty';
import parse from 'parse-duration';
import {useCallback, useMemo} from 'react';

import {DatumState, Input} from '@dash-frontend/api/pps';
import {restJobStateToNodeState} from '@dash-frontend/api/utils/nodeStateMappers';
import {useJobOrJobs} from '@dash-frontend/hooks/useJobOrJobs';
import useLogsNavigation from '@dash-frontend/hooks/useLogsNavigation';
import {useStopJob} from '@dash-frontend/hooks/useStopJob';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  calculateJobTotalRuntime,
  formatDurationFromSeconds,
  getStandardDateFromISOString,
} from '@dash-frontend/lib/dateTime';
import {DatumFilter, NodeState} from '@dash-frontend/lib/types';

import omitByDeep from './omitByDeep';

const useInfoPanel = () => {
  const {jobId, projectId, pipelineId} = useUrlState();
  const {getPathToDatumLogs} = useLogsNavigation();
  const {getUpdatedSearchParams, searchParams} = useUrlQueryState();

  const {job, loading: jobLoading} = useJobOrJobs({
    id: searchParams.globalIdFilter || jobId,
    pipelineName: pipelineId,
    projectId,
  });

  const {stopJob, loading: stopJobLoading} = useStopJob();

  const canStopJob = NodeState.RUNNING === restJobStateToNodeState(job?.state);

  const started = getStandardDateFromISOString(job?.started);

  const created = getStandardDateFromISOString(job?.created);

  const totalRuntime = calculateJobTotalRuntime(job);

  const cumulativeTime = formatDurationFromSeconds(
    (parse(job?.stats?.downloadTime ?? '', 's') ?? 0) +
      (parse(job?.stats?.processTime ?? '', 's') ?? 0) +
      (parse(job?.stats?.uploadTime ?? '', 's') ?? 0),
  );

  const datumMetrics = useMemo<
    {
      value: string | undefined;
      label: string;
      filter: DatumFilter;
    }[]
  >(() => {
    return [
      {
        value: job?.dataProcessed,
        label: 'Success',
        filter: DatumState.SUCCESS,
      },
      {
        value: job?.dataSkipped,
        label: 'Skipped',
        filter: DatumState.SKIPPED,
      },
      {
        value: job?.dataFailed,
        label: 'Failed',
        filter: DatumState.FAILED,
      },
      {
        value: job?.dataRecovered,
        label: 'Recovered',
        filter: DatumState.RECOVERED,
      },
    ];
  }, [job]);

  let logsDatumRoute = '';
  if ((job?.job?.id || jobId) && pipelineId) {
    logsDatumRoute = getPathToDatumLogs(
      {
        projectId,
        jobId: job?.job?.id || jobId,
        pipelineId: pipelineId,
      },
      [],
    );
  }

  const addLogsQueryParams = useCallback(
    (path: string, filter: DatumFilter) => {
      return `${path}&${getUpdatedSearchParams(
        {
          datumFilters: [filter],
        },
        true,
      )}`;
    },
    [getUpdatedSearchParams],
  );

  const runtimeMetrics = useMemo(() => {
    return [
      {
        duration: formatDurationFromSeconds(
          parse(job?.stats?.downloadTime ?? '', 's'),
        ),
        bytes: Number(job?.stats?.downloadBytes),
        label: 'Download',
      },
      {
        duration: formatDurationFromSeconds(
          parse(job?.stats?.uploadTime ?? '', 's'),
        ),
        bytes: Number(job?.stats?.uploadBytes),
        label: 'Upload',
      },
      {
        duration: formatDurationFromSeconds(
          parse(job?.stats?.processTime ?? '', 's'),
        ),
        label: 'Processing',
        bytes: undefined,
      },
    ];
  }, [job]);

  const inputs = useMemo(() => {
    const getInputRepos = (input: Input) => {
      const inputs: {projectId: string; name: string}[] = [];
      if (input.pfs?.repo && input.pfs.project) {
        inputs.push({
          projectId: input.pfs.project,
          name: input.pfs.repo,
        });
      }
      input.join?.forEach((i) => {
        inputs.push(...getInputRepos(i));
      });
      input.group?.forEach((i) => {
        inputs.push(...getInputRepos(i));
      });
      input.cross?.forEach((i) => {
        inputs.push(...getInputRepos(i));
      });
      input.union?.forEach((i) => {
        inputs.push(...getInputRepos(i));
      });

      return inputs;
    };

    if (job?.details?.input) {
      return getInputRepos(job?.details?.input);
    }
    return [];
  }, [job]);

  const jobConfig = useMemo(() => {
    if (!job) return null;
    else {
      const simplifiedSpec = omitByDeep(
        {
          pipelineVersion: job?.pipelineVersion,
          dataTotal: job?.dataTotal,
          dataFailed: job?.dataFailed,
          stats: job?.stats,
          salt: job?.details?.salt,
          datumTries: job?.details?.datumTries,
        },
        (val, _) => !val || (typeof val === 'object' && isEmpty(val)),
      );

      return {
        input: job?.details?.input || {},
        transform: job?.details?.transform || {},
        details: simplifiedSpec || {},
      };
    }
  }, [job]);

  return {
    job,
    totalRuntime,
    cumulativeTime,
    jobLoading,
    started,
    created,
    datumMetrics,
    runtimeMetrics,
    inputs,
    jobConfig,
    logsDatumRoute,
    addLogsQueryParams,
    stopJob,
    stopJobLoading,
    canStopJob,
  };
};

export default useInfoPanel;
