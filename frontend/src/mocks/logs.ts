import {Log, mockGetLogsQuery} from '@graphqlTypes';
import merge from 'lodash/merge';

export const buildLog = (log: Partial<Log>): Log => {
  const defaultLog = {
    message: '',
    user: false,
    timestamp: null,
    __typename: 'Log',
  };
  return merge(defaultLog, log);
};

export const mockEmptyGetLogs = () =>
  mockGetLogsQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        logs: {
          items: [],
          cursor: null,
          __typename: 'PageableLogs',
        },
      }),
    );
  });
