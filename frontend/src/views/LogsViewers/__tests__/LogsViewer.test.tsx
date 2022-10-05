import {pipelineAndJobLogs} from '@dash-backend/mock/fixtures/logs';
import {render, waitFor} from '@testing-library/react';
import {format, fromUnixTime} from 'date-fns';
import React from 'react';

import {withContextProviders, click} from '@dash-frontend/testHelpers';

import {LOGS_DATE_FORMAT} from '../constants/logsViewersConstants';
import JobLogsViewerComponent from '../JobLogsViewer';
import PipelineLogsViewerComponent from '../PipelineLogsViewer';

const PipelineLogsViewer = withContextProviders(() => {
  return <PipelineLogsViewerComponent />;
});
const JobLogsViewer = withContextProviders(() => {
  return <JobLogsViewerComponent />;
});

jest.mock(
  'react-virtualized-auto-sizer',
  () =>
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    ({children}: {children: (size: {height: number; width: number}) => any}) =>
      children({height: 50, width: 50}),
);

describe('Logs Viewer', () => {
  beforeEach(() => {
    window.history.replaceState({}, '', '/project/1/pipelines/edges/logs');
  });

  afterEach(() => {
    window.localStorage.removeItem('pachyderm-console-1');
  });

  it('should display empty state when there are no logs', async () => {
    const {findByRole, getByText, findAllByTestId, findByText} = render(
      <PipelineLogsViewer />,
    );

    expect(await findAllByTestId('LogRow__checkbox')).toHaveLength(2);

    const defaultOption = await findByRole('button', {
      name: 'Last Pipeline Job',
    });
    await click(defaultOption);
    const otherOption = getByText('Last 30 Minutes');
    await click(otherOption);
    expect(
      await findByText('No logs found for this time range.'),
    ).toBeInTheDocument();
  });

  it('should enable and disable export options', async () => {
    const {findByTestId, findAllByTestId, findByRole} = render(
      <PipelineLogsViewer />,
    );

    const selectAll = await findByTestId('LogsListHeader__select_all');
    const downloadButton = await findByRole('button', {
      name: 'Download',
    });

    const copyButton = await findByRole('button', {
      name: 'Copy selected rows',
    });

    await findAllByTestId('LogRow__checkbox');

    expect(copyButton).toBeDisabled();
    expect(downloadButton).toBeDisabled();

    await click(selectAll);

    expect(copyButton).not.toBeDisabled();
    expect(downloadButton).not.toBeDisabled();
  });

  it('should display logs', async () => {
    const {findAllByTestId} = render(<PipelineLogsViewer />);

    const rows = await findAllByTestId('LogRow__base');
    expect(rows).toHaveLength(2);
    expect(rows[0].textContent).toEqual(
      `${format(
        fromUnixTime(pipelineAndJobLogs['1'][0].getTs()?.getSeconds() || 0),
        LOGS_DATE_FORMAT,
      )} started datum task`,
    );
    expect(rows[1].textContent).toEqual(
      `${format(
        fromUnixTime(pipelineAndJobLogs['1'][1].getTs()?.getSeconds() || 0),
        LOGS_DATE_FORMAT,
      )} finished datum task`,
    );
  });

  it('should highlight user logs', async () => {
    const {queryAllByTestId, findAllByTestId, findByRole} = render(
      <PipelineLogsViewer />,
    );
    expect(await queryAllByTestId('LogRow__user_log')).toHaveLength(0);

    await click(await findByRole('switch', {name: 'Highlight User Logs'}));

    expect(await findAllByTestId('LogRow__user_log')).toHaveLength(1);
  });

  it('should display raw logs', async () => {
    const {findAllByTestId, findByRole} = render(<PipelineLogsViewer />);

    await click(await findByRole('switch', {name: 'Raw Logs'}));

    const rows = await findAllByTestId('RawLogRow__base');
    expect(rows).toHaveLength(2);
    expect(rows[0].textContent).toEqual('started datum task');
    expect(rows[1].textContent).toEqual('finished datum task');
  });

  it('should highlight raw user logs', async () => {
    const {queryAllByTestId, findAllByTestId, findByRole} = render(
      <PipelineLogsViewer />,
    );

    await click(await findByRole('switch', {name: 'Raw Logs'}));

    expect(await queryAllByTestId('RawLogRow__user_log')).toHaveLength(0);

    await click(await findByRole('switch', {name: 'Highlight User Logs'}));

    expect(await findAllByTestId('RawLogRow__user_log')).toHaveLength(1);
  });

  it('should store logs view preference on refresh', async () => {
    const {findAllByTestId, queryAllByTestId, findByRole} = render(
      <PipelineLogsViewer />,
    );

    expect(await findAllByTestId('LogRow__base')).toHaveLength(2);
    expect(await queryAllByTestId('RawLogRow__base')).toHaveLength(0);
    await click(await findByRole('switch', {name: 'Raw Logs'}));
    expect(await queryAllByTestId('LogRow__base')).toHaveLength(0);
    expect(await findAllByTestId('RawLogRow__base')).toHaveLength(2);

    window.history.replaceState({}, '', '/project/1/pipelines/edges/logs');
    expect(await queryAllByTestId('LogRow__base')).toHaveLength(0);
    expect(await findAllByTestId('RawLogRow__base')).toHaveLength(2);
  });

  it('should scroll to the latest log on first load', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/2/pipelines/likelihoods/logs',
    );
    const {queryByText, findAllByTestId} = render(<PipelineLogsViewer />);
    expect((await findAllByTestId('LogRow__checkbox')).length).toBeGreaterThan(
      2,
    );

    await waitFor(() =>
      expect(queryByText(/last message/)).toBeInTheDocument(),
    );
  });

  describe('Pipeline Logs Viewer', () => {
    beforeEach(() => {
      window.history.replaceState({}, '', '/project/1/pipelines/edges/logs');
    });

    it('should display pipeline name from url', async () => {
      const {findByText} = render(<PipelineLogsViewer />);

      expect(await findByText('edges')).toBeInTheDocument();
    });

    it('should route to pipeline view when modal is closed', async () => {
      const {findByTestId} = render(<PipelineLogsViewer />);
      await click(await findByTestId('FullPageModal__close'));

      await waitFor(() =>
        expect(window.location.pathname).toBe('/project/1/pipelines/edges'),
      );
    });

    it('should add the correct dropdown value', async () => {
      const {queryByRole} = render(<PipelineLogsViewer />);

      await waitFor(() =>
        expect(
          queryByRole('button', {
            name: 'Last Pipeline Job',
          }),
        ).toBeInTheDocument(),
      );
    });
  });

  describe('Job Logs Viewer', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/montage/logs',
      );
    });

    it('should display pipeline name from url', async () => {
      const {findByText} = render(<JobLogsViewer />);

      expect(
        await findByText('montage:23b9af7d5d4343219bc8e02ff44cd55a'),
      ).toBeInTheDocument();
    });

    it('should route to job view when modal is closed', async () => {
      const {findByTestId} = render(<JobLogsViewer />);

      await click(await findByTestId('FullPageModal__close'));
      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/montage',
        ),
      );
    });

    it('should add the correct dropdown value', async () => {
      const {queryByRole} = render(<JobLogsViewer />);

      await waitFor(() =>
        expect(
          queryByRole('button', {
            name: 'Job Start Time',
          }),
        ).toBeInTheDocument(),
      );
    });
  });
});
