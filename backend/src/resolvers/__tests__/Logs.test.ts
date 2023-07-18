import {
  GET_LOGS_QUERY,
  GET_WORKSPACE_LOGS_QUERY,
} from '@dash-frontend/queries/GetLogsQuerys';

import {executeQuery} from '@dash-backend/testHelpers';
import {GetLogsQuery, GetWorkspaceLogsQuery} from '@graphqlTypes';

describe('Logs resolver', () => {
  const projectId = 'Solar-Panel-Data-Sorting';

  describe('workspace query resolver', () => {
    it('should resolve workspace logs', async () => {
      const {data, errors = []} = await executeQuery<GetWorkspaceLogsQuery>(
        GET_WORKSPACE_LOGS_QUERY,
        {
          args: {projectId: 'default'},
        },
      );
      const workspaceLogs = data?.workspaceLogs;
      expect(errors).toHaveLength(0);
      expect(workspaceLogs).toHaveLength(5);
      expect(workspaceLogs?.[0]?.message).toContain(
        'auth.API.GetPermissionsForPrincipal',
      );
      expect(workspaceLogs?.[0]?.timestamp?.seconds).toBe(1623264952);
      expect(workspaceLogs?.[1]?.message).toContain(
        'pfs-over-HTTP - TLS disabled',
      );
      expect(workspaceLogs?.[1]?.timestamp?.seconds).toBe(1623264953);
      expect(workspaceLogs?.[2]?.message).toContain('auth.API.WhoAmI');
      expect(workspaceLogs?.[2]?.timestamp?.seconds).toBe(1623264954);
      expect(workspaceLogs?.[3]?.message).toContain('pps.API.GetLogs');
      expect(workspaceLogs?.[3]?.timestamp?.seconds).toBe(1623264955);
      expect(workspaceLogs?.[4]?.message).toBe(
        'PPS master: processing event for "edges"',
      );
      expect(workspaceLogs?.[4]?.timestamp).toBeNull();
    });
  });

  describe('logs query resolver', () => {
    it('should resolve pipeline logs', async () => {
      const {data, errors = []} = await executeQuery<GetLogsQuery>(
        GET_LOGS_QUERY,
        {
          args: {pipelineName: 'montage', start: 1616533099, projectId},
        },
      );
      const pipelineLogs = data?.logs;
      expect(errors).toHaveLength(0);
      expect(pipelineLogs?.items).toHaveLength(6);
      expect(pipelineLogs?.items[0]?.message).toBe('started datum task');
      expect(pipelineLogs?.items[1]?.message).toBe(
        'beginning to run user code',
      );
      expect(pipelineLogs?.items[2]?.message).toContain(
        'UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.',
      );
      expect(pipelineLogs?.items[3]?.message).toBe('finished datum task');
    });

    it('should resolve job logs', async () => {
      const {data, errors = []} = await executeQuery<GetLogsQuery>(
        GET_LOGS_QUERY,
        {
          args: {
            pipelineName: 'montage',
            jobId: '23b9af7d5d4343219bc8e02ff44cd55a',
            start: 1616533099,
            projectId,
          },
        },
      );
      const jobLogs = data?.logs;
      expect(errors).toHaveLength(0);
      expect(jobLogs?.items).toHaveLength(6);
      expect(jobLogs?.items[0]?.message).toBe('started datum task');
      expect(jobLogs?.items[0]?.timestamp?.seconds).toBe(1616533099);
      expect(jobLogs?.items[5]?.message).toBe('finished datum task');
      expect(jobLogs?.items[5]?.timestamp?.seconds).toBe(1616533220);
    });

    it('should resolve datum logs', async () => {
      const {data, errors = []} = await executeQuery<GetLogsQuery>(
        GET_LOGS_QUERY,
        {
          args: {
            pipelineName: 'montage',
            jobId: '23b9af7d5d4343219bc8e02ff44cd55a',
            datumId:
              '0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
            start: 1616533099,
            projectId,
          },
        },
      );
      const datumLogs = data?.logs;
      expect(errors).toHaveLength(0);
      expect(datumLogs?.items).toHaveLength(4);
      expect(datumLogs?.items[0]?.message).toBe('started datum task');
      expect(datumLogs?.items[0]?.timestamp?.seconds).toBe(1616533099);
      expect(datumLogs?.items[3]?.message).toBe('finished datum task');
      expect(datumLogs?.items[3]?.timestamp?.seconds).toBe(1616533106);
    });

    it('should resolve master logs', async () => {
      const {data, errors = []} = await executeQuery<GetLogsQuery>(
        GET_LOGS_QUERY,
        {
          args: {
            pipelineName: 'montage',
            projectId,
            master: true,
            start: 1614126189,
          },
        },
      );
      const masterLogs = data?.logs;
      expect(errors).toHaveLength(0);
      expect(masterLogs?.items).toHaveLength(1);
      expect(masterLogs?.items[0]?.message).toBe('started datum task');
      expect(masterLogs?.items[0]?.timestamp?.seconds).toBe(1614126189);
    });

    it('should page logs request', async () => {
      const {data, errors = []} = await executeQuery<GetLogsQuery>(
        GET_LOGS_QUERY,
        {
          args: {
            pipelineName: 'montage',
            jobId: '23b9af7d5d4343219bc8e02ff44cd55a',
            projectId,
            cursor: {
              message: 'started datum task',
              timestamp: {
                seconds: 1616533099,
                nanos: 0,
              },
            },
            limit: 3,
          },
        },
      );
      const jogLogs = data?.logs;
      expect(errors).toHaveLength(0);
      expect(jogLogs?.items).toHaveLength(3);
      expect(jogLogs?.items[0]?.message).toBe('beginning to run user code');
      expect(jogLogs?.items[2]?.message).toBe('finished datum task');
    });

    it('should return a cursor to the next page of logs', async () => {
      const {data, errors = []} = await executeQuery<GetLogsQuery>(
        GET_LOGS_QUERY,
        {
          args: {
            pipelineName: 'montage',
            jobId: '23b9af7d5d4343219bc8e02ff44cd55a',
            projectId,
            cursor: {
              message: 'started datum task',
              timestamp: {
                seconds: 1616533099,
                nanos: 0,
              },
            },
            limit: 3,
          },
        },
      );
      const jogLogs = data?.logs;
      expect(errors).toHaveLength(0);
      expect(jogLogs?.items).toHaveLength(3);
      expect(jogLogs?.items[0]?.message).toBe('beginning to run user code');
      expect(jogLogs?.items[2]?.message).toBe('finished datum task');
      expect(jogLogs?.cursor).toEqual(
        expect.objectContaining({
          __typename: 'LogCursor',
          message: 'finished datum task',
          timestamp: {
            __typename: 'Timestamp',
            nanos: 0,
            seconds: 1616533106,
          },
        }),
      );
    });

    it('should skip a cursor withought a timestamp', async () => {
      const {data, errors = []} = await executeQuery<GetLogsQuery>(
        GET_LOGS_QUERY,
        {
          args: {
            pipelineName: 'montage',
            jobId: '23b9af7d5d4343219bc8e02ff44cd55a',
            projectId,
            cursor: {
              message: 'started datum task',
              timestamp: {
                seconds: 1616533099,
                nanos: 0,
              },
            },
            limit: 2,
          },
        },
      );
      const jogLogs = data?.logs;
      expect(errors).toHaveLength(0);
      expect(jogLogs?.items).toHaveLength(2);
      expect(jogLogs?.items[0]?.message).toBe('beginning to run user code');
      expect(jogLogs?.items[1]?.message).toBe(
        '/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.',
      );
      expect(jogLogs?.cursor).toEqual(
        expect.objectContaining({
          __typename: 'LogCursor',
          message: 'beginning to run user code',
          timestamp: {
            __typename: 'Timestamp',
            nanos: 0,
            seconds: 1616533100,
          },
        }),
      );
    });
  });
});
