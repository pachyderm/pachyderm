import {executeQuery} from '@dash-backend/testHelpers';
import {
  GET_JOB_LOGS_QUERY,
  GET_PIPELINE_LOGS_QUERY,
  GET_WORKSPACE_LOGS_QUERY,
} from '@dash-frontend/queries/GetLogsQuerys';
import {
  GetJobLogsQuery,
  GetPipelineLogsQuery,
  GetWorkspaceLogsQuery,
} from '@graphqlTypes';

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
      expect(errors.length).toBe(0);
      expect(workspaceLogs?.length).toEqual(5);
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

  describe('pipeline query resolver', () => {
    it('should resolve pipeline logs', async () => {
      const {data, errors = []} = await executeQuery<GetPipelineLogsQuery>(
        GET_PIPELINE_LOGS_QUERY,
        {
          args: {pipelineName: 'montage', start: 1616533099, projectId},
        },
      );
      const workspaceLogs = data?.pipelineLogs;
      expect(errors.length).toBe(0);
      expect(workspaceLogs?.length).toEqual(4);
      expect(workspaceLogs?.[0]?.message).toEqual('started datum task');
      expect(workspaceLogs?.[1]?.message).toEqual('beginning to run user code');
      expect(workspaceLogs?.[2]?.message).toContain(
        'UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.',
      );
      expect(workspaceLogs?.[3]?.message).toEqual('finished datum task');
    });
  });

  describe('job query resolver', () => {
    it('should resolve job logs', async () => {
      const {data, errors = []} = await executeQuery<GetJobLogsQuery>(
        GET_JOB_LOGS_QUERY,
        {
          args: {
            pipelineName: 'edges',
            jobId: '23b9af7d5d4343219bc8e02ff4acd33a',
            start: 1614126189,
            projectId,
          },
        },
      );
      const workspaceLogs = data?.jobLogs;
      expect(errors.length).toBe(0);
      expect(workspaceLogs?.length).toEqual(2);
      expect(workspaceLogs?.[0]?.message).toEqual('started datum task');
      expect(workspaceLogs?.[1]?.message).toEqual('finished datum task');
    });
  });
});
