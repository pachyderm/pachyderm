import {DatumFilter} from '@graphqlTypes';
import {render, screen} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';
import {logsViewerJobRoute} from '@dash-frontend/views/Project/utils/routes';

import useLogsNavigation from '../useLogsNavigation';

const NavigationToComponent = withContextProviders(
  ({
    args,
    datumFilters,
  }: {
    args?: Parameters<typeof logsViewerJobRoute>[0];
    datumFilters: DatumFilter[];
  }) => {
    const {getPathToJobLogs, getPathToDatumLogs, getPathToLatestJobLogs} =
      useLogsNavigation();
    const pathToJobs = getPathToJobLogs(args);
    const pathToDatum = getPathToDatumLogs(args, datumFilters);
    const pathToLatest = getPathToLatestJobLogs(args);

    return (
      <div>
        <span>path to job logs: {pathToJobs}</span>
        <span>path to datum logs: {pathToDatum}</span>
        <span>path to latest logs: {pathToLatest}</span>
      </div>
    );
  },
);

const NavigationFromComponent = withContextProviders(
  ({backupPath}: {backupPath: string}) => {
    const {getPathFromLogs} = useLogsNavigation();
    const pathFrom = getPathFromLogs(backupPath);

    return (
      <div>
        <span>path from logs: {pathFrom}</span>
      </div>
    );
  },
);

describe('useLogsNavigation', () => {
  it('should get the correct path to the job logs', async () => {
    const backPath = '/lineage/default/pipelines/montage';
    const args = {
      projectId: 'projectA',
      jobId: '2345',
      pipelineId: 'montage',
    };
    window.history.pushState('', '', backPath);

    render(<NavigationToComponent args={args} />);
    expect(
      screen.getByText(
        `path to job logs: /lineage/${args.projectId}/pipelines/${
          args.pipelineId
        }/jobs/${args.jobId}/logs?prevPath=${encodeURIComponent(backPath)}`,
      ),
    ).toBeInTheDocument();
  });

  it('should get the correct path to datum logs', async () => {
    const backPath = '/lineage/default/pipelines/montage';
    const args = {
      projectId: 'projectA',
      jobId: '2345',
      pipelineId: 'montage',
    };
    const datumFilter = [DatumFilter.SUCCESS];
    window.history.pushState('', '', backPath);

    render(<NavigationToComponent args={args} datumFilter={datumFilter} />);
    expect(
      screen.getByText(
        `path to datum logs: /lineage/${args.projectId}/pipelines/${
          args.pipelineId
        }/jobs/${args.jobId}/logs/datum?prevPath=${encodeURIComponent(
          backPath,
        )}`,
      ),
    ).toBeInTheDocument();
  });

  it('should get the correct path to the latest logs', async () => {
    const backPath = '/lineage/default/pipelines/montage';
    const args = {
      projectId: 'projectA',
      jobId: '2345',
      pipelineId: 'montage',
    };
    window.history.pushState('', '', backPath);

    render(<NavigationToComponent args={args} />);
    expect(
      screen.getByText(
        `path to latest logs: /lineage/${args.projectId}/pipelines/${
          args.pipelineId
        }/logs?prevPath=${encodeURIComponent(backPath)}`,
      ),
    ).toBeInTheDocument();
  });

  it('should get the correct backup path from the logs if prevPath is not available', async () => {
    const backPath = '/backupPath';
    const path =
      '/lineage/projectA/pipelines/montage/jobs/2345/logs/datum?prevPath=%2FbackupPath';
    window.history.pushState('', '', path);

    render(<NavigationFromComponent backupPath={backPath} />);
    expect(screen.getByText('path from logs: /backupPath')).toBeInTheDocument();
  });
});
