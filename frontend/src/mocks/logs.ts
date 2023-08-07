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

const MOCK_LOGS: Log[] = [
  {
    timestamp: {
      seconds: 1690919093,
      nanos: 971781303,
      __typename: 'Timestamp',
    },
    user: false,
    message: 'started process datum set task',
    __typename: 'Log',
  },
  {
    timestamp: {
      seconds: 1690919094,
      nanos: 112230845,
      __typename: 'Timestamp',
    },
    user: false,
    message: 'beginning to run user code',
    __typename: 'Log',
  },
  {
    timestamp: {
      seconds: 1690919094,
      nanos: 695166387,
      __typename: 'Timestamp',
    },
    user: true,
    message:
      "montage: unable to open image '/pfs/edges/Screenshot': No such file or directory @ error/blob.c/OpenBlob/3527.",
    __typename: 'Log',
  },
  {
    timestamp: {
      seconds: 1690919094,
      nanos: 695190928,
      __typename: 'Timestamp',
    },
    user: true,
    message:
      "montage: no decode delegate for this image format `' @ error/constitute.c/ReadImage/740.",
    __typename: 'Log',
  },
  {
    timestamp: {
      seconds: 1690919094,
      nanos: 695193053,
      __typename: 'Timestamp',
    },
    user: true,
    message:
      "montage: unable to open image 'test': No such file or directory @ error/blob.c/OpenBlob/3527.",
    __typename: 'Log',
  },
  {
    timestamp: {
      seconds: 1690919094,
      nanos: 695197342,
      __typename: 'Timestamp',
    },
    user: false,
    message: 'finished running user code',
    __typename: 'Log',
  },
  {
    timestamp: {
      seconds: 1690919095,
      nanos: 955660095,
      __typename: 'Timestamp',
    },
    user: false,
    message: 'finished process datum set task',
    __typename: 'Log',
  },
];

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

export const mockGetLogs = () =>
  mockGetLogsQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        logs: {
          items: MOCK_LOGS,
          cursor: null,
          __typename: 'PageableLogs',
        },
      }),
    );
  });
