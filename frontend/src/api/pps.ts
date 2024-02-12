import isEqual from 'lodash/isEqual';

import {
  JobInfo,
  ListJobRequest,
  ListPipelineRequest,
  InspectPipelineRequest,
  PipelineInfo,
  InspectJobSetRequest,
  SetClusterDefaultsRequest,
  CreatePipelineV2Request,
  DeletePipelineRequest,
  RerunPipelineRequest,
  InspectJobRequest,
  InspectDatumRequest,
  GetProjectDefaultsRequest,
  SetProjectDefaultsRequest,
  ListDatumRequest,
  DatumState,
  GetLogsRequest,
  LogMessage,
  DatumInfo,
  StopJobRequest,
  StartPipelineRequest,
  StopPipelineRequest,
  RunCronRequest,
} from '@dash-frontend/generated/proto/pps/pps.pb';
import {getUnixSecondsFromISOString} from '@dash-frontend/lib/dateTime';
import {InternalJobSet} from '@dash-frontend/lib/types';
import {LOGS_PAGE_SIZE} from '@dash-frontend/views/DatumViewer/components/MiddleSection/components/LogsViewer/constants/logsViewersConstants';

import {pps} from './base/api';
import {isAbortSignalError} from './utils/error';
import {getRequestOptions} from './utils/requestHeaders';
import {jobsMapToJobSet} from './utils/transform';

export const inspectPipeline = async (req: InspectPipelineRequest) => {
  return await pps.InspectPipeline(req, getRequestOptions());
};

export const listPipeline = async (req: ListPipelineRequest) => {
  const pipelines: PipelineInfo[] = [];

  await pps.ListPipeline(
    req,
    (pipeline) => pipelines.push(pipeline),
    getRequestOptions(),
  );

  return pipelines;
};

export const listJob = async (req: ListJobRequest) => {
  const jobs: JobInfo[] = [];

  await pps.ListJob(req, (job) => jobs.push(job), getRequestOptions());

  return jobs;
};

export const inspectJob = async (req: InspectJobRequest) => {
  return await pps.InspectJob(req, getRequestOptions());
};

export const inspectJobSet = async (req: InspectJobSetRequest) => {
  const jobSet: JobInfo[] = [];

  await pps.InspectJobSet(req, (job) => jobSet.push(job), getRequestOptions());

  return jobSet;
};

export const listJobSets = async (
  req: Omit<ListJobRequest, 'number'>,
  limit: number,
): Promise<{countExceededLimit: boolean; jobSets: InternalJobSet[]}> => {
  const jobsMap = new Map<string, JobInfo[]>();
  const controller = new AbortController();

  let counter = 0;

  try {
    await pps.ListJob(
      req,
      (job) => {
        const addNewJobSets = jobsMap.size < limit;

        if (job?.job?.id) {
          const jobId = job.job.id;

          if (!addNewJobSets && !jobsMap.has(jobId)) {
            return controller.abort();
          }

          counter += 1;

          const jobsInJobSet = jobsMap.get(jobId) || [];

          jobsInJobSet.push(job);
          jobsMap.set(jobId, jobsInJobSet);
        }
      },
      {
        ...getRequestOptions(),
        signal: controller.signal,
      },
    );
  } catch (error) {
    if (isAbortSignalError(error)) {
      return {
        countExceededLimit: counter >= limit,
        jobSets: jobsMapToJobSet(jobsMap),
      };
    }

    throw error;
  }

  return {
    countExceededLimit: counter >= limit,
    jobSets: jobsMapToJobSet(jobsMap),
  };
};

type GetLogsPaging = {
  limit?: number;
  cursor?: LogMessage;
  since?: string;
};

export const getLogs = async (
  req: GetLogsRequest,
  {limit = LOGS_PAGE_SIZE, cursor, since}: GetLogsPaging,
) => {
  const logs: LogMessage[] = [];
  const controller = new AbortController();
  let foundCursor = false;
  let count = 0;
  const DEFAULT_TIME_BUFFER_IN_SECONDS = 3;

  // If datum id is not a full ID it will cause the API to not return anything and the parent datum object needs to be deleted to fix it.
  if (!req.datum?.id) delete req.datum;

  const calculateSince = (startTime?: string): string | undefined => {
    if (startTime) {
      const now = Math.floor(Date.now() / 1000);
      const unixStart = getUnixSecondsFromISOString(startTime);
      const nowStartDiff = now - unixStart;

      return `${nowStartDiff + DEFAULT_TIME_BUFFER_IN_SECONDS}s`;
    }
  };

  const handleLogStream = (log: LogMessage) => {
    if (count === limit) {
      controller.abort();
    } else {
      logs.push(log);
      count++;
    }
  };

  const requestInitalization = {
    ...getRequestOptions(),
    signal: controller.signal,
  };

  if (cursor) {
    try {
      const providedLogIsCursorLog = (log: LogMessage) =>
        // Timestamps are precise ("2023-12-04T17:29:53.104828172Z") and therefore unlikely (though possible) to collide on time.
        log.ts && cursor && log.ts === cursor.ts && isEqual(log, cursor);

      const cursorRequest = {...req, since: calculateSince(cursor.ts)};
      await pps.GetLogs(
        cursorRequest,
        (log) => {
          if (!foundCursor && providedLogIsCursorLog(log)) {
            foundCursor = true;
          } else if (foundCursor) {
            handleLogStream(log);
          }
        },
        requestInitalization,
      );
    } catch (e) {
      if (!isAbortSignalError(e)) throw e;
    }
    if (cursor && !foundCursor) {
      const pagingError = new Error(
        'Did not find provided cursor in logs stream.',
      );
      pagingError.name = 'PagingError';
      throw pagingError;
    }
  } else {
    try {
      const nonCursorRequest = {...req, since: calculateSince(since)};
      await pps.GetLogs(
        nonCursorRequest,
        handleLogStream,
        requestInitalization,
      );
    } catch (e) {
      if (!isAbortSignalError(e)) throw e;
    }
  }

  return logs;
};

export const listDatum = async (req: ListDatumRequest) => {
  const datums: DatumInfo[] = [];

  await pps.ListDatum(req, (datum) => datums.push(datum), getRequestOptions());

  return datums;
};

export const listDatumPaged = async (req: ListDatumRequest) => {
  // We do not want to show UNKNOWN and STARTING statuses in the datum list.
  // These two statuses are not being used by core so we will not allow users to ask for them.
  const DEFAULT_FILTERS = [
    DatumState.FAILED,
    DatumState.RECOVERED,
    DatumState.SKIPPED,
    DatumState.SUCCESS,
  ];

  const modifiedArgs = {
    ...req,
    filter: {
      state: req?.filter?.state ? req.filter.state : DEFAULT_FILTERS,
    },
    number: req?.number && String(Number(req.number) + 1),
  };

  const datums = await listDatum(modifiedArgs);

  let nextCursor = undefined;

  // If datums.length is not greater than limit there are no pages left
  if (req?.number && datums && datums?.length > Number(req.number)) {
    datums.pop(); //remove the extra datum from the response
    nextCursor = datums[datums.length - 1].datum?.id;
  }

  return {
    datums,
    cursor: nextCursor,
    hasNextPage: !!nextCursor,
  };
};

export const inspectDatum = async (req: InspectDatumRequest) => {
  return await pps.InspectDatum(req, getRequestOptions());
};

export const getClusterDefaults = async () => {
  return await pps.GetClusterDefaults({}, getRequestOptions());
};

export const setClusterDefaults = async (req: SetClusterDefaultsRequest) => {
  return await pps.SetClusterDefaults(req, getRequestOptions());
};

export const getProjectDefaults = async (req: GetProjectDefaultsRequest) => {
  return await pps.GetProjectDefaults(req, getRequestOptions());
};

export const setProjectDefaults = async (req: SetProjectDefaultsRequest) => {
  return await pps.SetProjectDefaults(req, getRequestOptions());
};

export const createPipeline = async (req: CreatePipelineV2Request) => {
  return await pps.CreatePipelineV2(req, getRequestOptions());
};

export const deletePipeline = async (req: DeletePipelineRequest) => {
  return await pps.DeletePipeline(req, getRequestOptions());
};

export const rerunPipeline = async (req: RerunPipelineRequest) => {
  return await pps.RerunPipeline(req, getRequestOptions());
};

export const stopPipeline = async (req: StopPipelineRequest) => {
  return await pps.StopPipeline(req, getRequestOptions());
};

export const startPipeline = async (req: StartPipelineRequest) => {
  return await pps.StartPipeline(req, getRequestOptions());
};

export const stopJob = async (req: StopJobRequest) => {
  return await pps.StopJob(req, getRequestOptions());
};

export const runCron = async (req: RunCronRequest) => {
  return await pps.RunCron(req, getRequestOptions());
};

// Export all of the pps types
export * from '@dash-frontend/generated/proto/pps/pps.pb';
