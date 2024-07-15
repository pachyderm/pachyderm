import merge from 'lodash/merge';
import {rest} from 'msw';

import {Empty} from '@dash-frontend/api/googleTypes';
import {GetLogsRequest, LogMessage} from '@dash-frontend/api/pps';

export const buildLog = (log: Partial<LogMessage>): LogMessage => {
  const defaultLog: LogMessage = {
    message: '',
    user: false,
    ts: undefined,
    __typename: 'LogMessage',
  };
  return merge(defaultLog, log);
};

const MOCK_LOGS: LogMessage[] = [
  {
    ts: '2023-12-01T21:29:44.047823092Z',
    user: false,
    message: 'started process datum set task',
  },
  {
    ts: '2023-12-01T21:29:44.047823092Z',
    user: false,
    message: 'beginning to run user code',
  },
  {
    ts: '2023-12-01T21:29:44.047823092Z',
    user: true,
    message:
      "montage: unable to open image '/pfs/edges/Screenshot': No such file or directory @ error/blob.c/OpenBlob/3527.",
  },
  {
    ts: '2023-12-01T21:29:44.047823092Z',
    user: true,
    message:
      "montage: no decode delegate for this image format `' @ error/constitute.c/ReadImage/740.",
  },
  {
    ts: '2023-12-01T21:29:44.047823092Z',
    user: true,
    message:
      "montage: unable to open image 'test': No such file or directory @ error/blob.c/OpenBlob/3527.",
  },
  {
    ts: '2023-12-01T21:29:44.047823092Z',
    user: false,
    message: 'finished running user code',
  },
  {
    ts: '2023-12-01T21:30:44.047823092Z',
    user: false,
    message: 'finished process datum set task',
  },
];

export const mockEmptyGetLogs = () =>
  rest.post<GetLogsRequest, Empty, LogMessage[]>(
    '/api/pps_v2.API/GetLogs',
    (_req, res, ctx) => {
      return res(ctx.json([]));
    },
  );

export const mockGetLogs = () =>
  rest.post<GetLogsRequest, Empty, LogMessage[]>(
    '/api/pps_v2.API/GetLogs',
    async (req, res, ctx) => {
      const body = await req.json();
      if (
        body.job?.id === '5c1aa9bc87dd411ba5a1be0c80a3ebc2' &&
        body.pipeline?.name === 'montage'
      ) {
        return res(ctx.json(MOCK_LOGS));
      }
      return res(ctx.json([]));
    },
  );

export const mockGetServiceOrSpoutLogs = () =>
  rest.post<GetLogsRequest, Empty, LogMessage[]>(
    '/api/pps_v2.API/GetLogs',
    async (req, res, ctx) => {
      const body = await req.json();
      if (!body.job) {
        return res(ctx.json(MOCK_LOGS));
      }
      return res(ctx.json([]));
    },
  );
