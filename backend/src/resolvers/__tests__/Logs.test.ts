import {
  GET_LOGS_QUERY,
  GET_WORKSPACE_LOGS_QUERY,
} from '@dash-frontend/queries/GetLogsQuerys';

import {executeQuery} from '@dash-backend/testHelpers';
import {GetLogsQuery, GetWorkspaceLogsQuery} from '@graphqlTypes';

describe('Logs resolver', () => {
  const projectId = '1';

  describe('workspace query resolver', () => {
    it('should resolve workspace logs', async () => {
      const {data, errors = []} = await executeQuery<GetWorkspaceLogsQuery>(
        GET_WORKSPACE_LOGS_QUERY,
        {
          args: {},
        },
      );
      const workspaceLogs = data?.workspaceLogs;
      expect(errors).toHaveLength(0);
      expect(workspaceLogs).toHaveLength(5);
      expect(workspaceLogs?.[0]?.message).toContain(
        'auth.API.GetPermissionsForPrincipal',
      );
      expect(workspaceLogs?.[0]?.timestamp?.seconds).toEqual(1623264952);
      expect(workspaceLogs?.[1]?.message).toContain(
        'pfs-over-HTTP - TLS disabled',
      );
      expect(workspaceLogs?.[1]?.timestamp?.seconds).toEqual(1623264953);
      expect(workspaceLogs?.[2]?.message).toContain('auth.API.WhoAmI');
      expect(workspaceLogs?.[2]?.timestamp?.seconds).toEqual(1623264954);
      expect(workspaceLogs?.[3]?.message).toContain('pps.API.GetLogs');
      expect(workspaceLogs?.[3]?.timestamp?.seconds).toEqual(1623264955);
      expect(workspaceLogs?.[4]?.message).toEqual(
        'PPS master: processing event for "edges"',
      );
      expect(workspaceLogs?.[4]?.timestamp).toBe(null);
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
      const workspaceLogs = data?.logs;
      expect(errors).toHaveLength(0);
      expect(workspaceLogs).toHaveLength(6);
      expect(workspaceLogs?.[0]?.message).toEqual('started datum task');
      expect(workspaceLogs?.[1]?.message).toEqual('beginning to run user code');
      expect(workspaceLogs?.[2]?.message).toContain(
        'UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.',
      );
      expect(workspaceLogs?.[3]?.message).toEqual('finished datum task');
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
      const workspaceLogs = data?.logs;
      expect(errors).toHaveLength(0);
      expect(workspaceLogs).toHaveLength(6);
      expect(workspaceLogs?.[0]?.message).toEqual('started datum task');
      expect(workspaceLogs?.[0]?.timestamp?.seconds).toEqual(1616533099);
      expect(workspaceLogs?.[5]?.message).toEqual('finished datum task');
      expect(workspaceLogs?.[5]?.timestamp?.seconds).toEqual(1616533220);
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
      const workspaceLogs = data?.logs;
      expect(errors).toHaveLength(0);
      expect(workspaceLogs).toHaveLength(4);
      expect(workspaceLogs?.[0]?.message).toEqual('started datum task');
      expect(workspaceLogs?.[0]?.timestamp?.seconds).toEqual(1616533099);
      expect(workspaceLogs?.[3]?.message).toEqual('finished datum task');
      expect(workspaceLogs?.[3]?.timestamp?.seconds).toEqual(1616533106);
    });

    it('should reverse logs order', async () => {
      const {data, errors = []} = await executeQuery<GetLogsQuery>(
        GET_LOGS_QUERY,
        {
          args: {
            pipelineName: 'montage',
            jobId: '33b9af7d5d4343219bc8e02ff44cd55a',
            start: 1614126189,
            projectId,
            reverse: true,
          },
        },
      );
      const workspaceLogs = data?.logs;
      expect(errors).toHaveLength(0);
      expect(workspaceLogs).toHaveLength(2);
      expect(workspaceLogs?.[0]?.message).toEqual('finished datum task');
      expect(workspaceLogs?.[1]?.message).toEqual('started datum task');
    });
  });
});
