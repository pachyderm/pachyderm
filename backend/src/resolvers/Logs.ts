import {default as reverseArray} from 'lodash/reverse';
import {v4 as uuid} from 'uuid';

import withStream from '@dash-backend/lib/withStream';
import {LogMessage} from '@dash-backend/proto';
import {Log, Maybe, QueryResolvers, SubscriptionResolvers} from '@graphqlTypes';

import {logMessageToGQLLog} from './builders/pps';
interface LogsResolver {
  Query: {
    workspaceLogs: QueryResolvers['workspaceLogs'];
    logs: QueryResolvers['logs'];
  };
  Subscription: {
    workspaceLogs: SubscriptionResolvers['workspaceLogs'];
    logs: SubscriptionResolvers['logs'];
  };
}

const WORKSPACE_MESSAGE_REGEX = /INFO|WARNING|ERROR/;
const DEFAULT_TIME_BUFFER_IN_SECONDS = 3;
const parseWorkspaceLog = (log: LogMessage.AsObject) => {
  const splitLog = log.message.split(WORKSPACE_MESSAGE_REGEX);
  if (splitLog.length === 2) {
    const parsedTs = Date.parse(splitLog[0].trim());
    if (!isNaN(parsedTs) && typeof parsedTs === 'number') {
      return {
        user: log.user,
        timestamp: {
          seconds: Math.floor(parsedTs / 1000),
          nanos: (parsedTs % 1000) * 1000000,
        },
        message: splitLog[1].trim(),
      };
    }
  }
  return logMessageToGQLLog(log);
};

const calculateSince = (startTime?: Maybe<number>) => {
  if (startTime) {
    return (
      Math.floor(Date.now() / 1000) - startTime + DEFAULT_TIME_BUFFER_IN_SECONDS
    );
  }
};

const logsResolver: LogsResolver = {
  Query: {
    workspaceLogs: async (_field, {args: {projectId, start}}, {pachClient}) => {
      const logs = await pachClient
        .pps()
        .getLogs({projectId, since: calculateSince(start)});

      return logs.map(parseWorkspaceLog);
    },
    logs: async (
      _field,
      {
        args: {
          projectId,
          pipelineName,
          jobId,
          datumId,
          start,
          reverse = false,
          master = false,
        },
      },
      {pachClient},
    ) => {
      const logs = await pachClient.pps().getLogs({
        projectId,
        pipelineName: pipelineName,
        since: calculateSince(start),
        jobId: jobId || undefined,
        datumId: datumId || undefined,
        master: master || undefined,
      });

      return reverse
        ? reverseArray(logs.map(logMessageToGQLLog))
        : logs.map(logMessageToGQLLog);
    },
  },

  Subscription: {
    workspaceLogs: {
      subscribe: async (_field, {args: {projectId, start}}, {pachClient}) => {
        const stream = await pachClient
          .pps()
          .getLogsStream({projectId, since: calculateSince(start)});

        return {
          [Symbol.asyncIterator]: () =>
            withStream<Log, LogMessage>({
              triggerName: `${uuid()}_WORKSPACE_LOGS`,
              stream: stream,
              onData: (chunk) => parseWorkspaceLog(chunk.toObject()),
            }),
        };
      },
      resolve: (result: Log) => {
        return result;
      },
    },

    logs: {
      subscribe: async (
        _field,
        {args: {projectId, pipelineName, jobId, start}},
        {pachClient},
      ) => {
        const stream = await pachClient.pps().getLogsStream({
          projectId,
          pipelineName: pipelineName,
          jobId: jobId || undefined,
          since: calculateSince(start),
        });

        const triggerName =
          pipelineName && jobId
            ? `${uuid()}_${pipelineName}_${jobId}_LOGS`
            : `${uuid()}_${pipelineName}_LOGS`;
        return {
          [Symbol.asyncIterator]: () =>
            withStream<Log, LogMessage>({
              triggerName: triggerName,
              stream: stream,
              onData: (chunk) => logMessageToGQLLog(chunk.toObject()),
            }),
        };
      },
      resolve: (result: Log) => {
        return result;
      },
    },
  },
};

export default logsResolver;
