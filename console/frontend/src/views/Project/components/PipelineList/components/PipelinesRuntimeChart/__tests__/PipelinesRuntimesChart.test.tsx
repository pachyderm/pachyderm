import {
  render,
  screen,
  within,
  waitForElementToBeRemoved,
  act,
} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {JobInfo, ListJobRequest} from '@dash-frontend/api/pps';
import {
  MOCK_MONTAGE_JOBS_INFO,
  MOCK_EDGES_JOBS_INFO,
} from '@dash-frontend/mocks';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import RuntimesChartComponent from '../PipelinesRuntimeChart';

describe('Pipelines Runtimes Chart', () => {
  const server = setupServer();

  const RuntimesChart = withContextProviders(() => {
    return (
      <RuntimesChartComponent
        filtersExpanded
        selectedPipelines={['edges', 'montage']}
      />
    );
  });

  beforeAll(() => {
    server.listen();

    server.use(
      rest.post<ListJobRequest, Empty, JobInfo[]>(
        '/api/pps_v2.API/ListJob',
        async (req, res, ctx) => {
          const body = await req.json();

          if (
            body?.['pipeline']?.['name'] === 'edges' &&
            body?.number === '3'
          ) {
            return res(ctx.json([...MOCK_EDGES_JOBS_INFO.slice(0, 3)]));
          } else if (
            body?.['pipeline']?.['name'] === 'montage' &&
            body?.number === '3'
          ) {
            return res(ctx.json([...MOCK_MONTAGE_JOBS_INFO.slice(0, 3)]));
          } else if (body?.['pipeline']?.['name'] === 'edges') {
            return res(ctx.json(MOCK_EDGES_JOBS_INFO));
          } else if (body?.['pipeline']?.['name'] === 'montage') {
            return res(ctx.json(MOCK_MONTAGE_JOBS_INFO));
          }
        },
      ),
    );
  });

  beforeEach(() => {
    window.history.replaceState('', '', '/project/default/pipelines/runtimes');
  });

  afterAll(() => server.close());

  it('should display datasets based on given jobs', async () => {
    render(<RuntimesChart />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RuntimesChart__loadingDots'),
    );

    expect(screen.getByText(/; 1dc67e.../)).toBeInTheDocument();
    expect(screen.getByText(/; a44234.../)).toBeInTheDocument();
    expect(screen.getByText(/; 5c1aa9.../)).toBeInTheDocument();
    expect(screen.getByText(/Runtimes for 2 pipelines/)).toBeInTheDocument();

    expect(screen.getByText('Failed Datums')).toBeInTheDocument();
    expect(
      screen.getByRole('img', {
        name: /runtimes chart/i,
      }),
    ).toBeInTheDocument();
  });

  it('should filter jobs by ID and pipeline', async () => {
    render(<RuntimesChart />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RuntimesChart__loadingDots'),
    );

    await click(screen.getByText('Select job IDs'));
    await click(screen.getByLabelText('1dc67e...'));

    expect(screen.getByText(/; 1dc67e.../)).toBeInTheDocument();
    expect(screen.queryByText(/; a44234.../)).not.toBeInTheDocument();
    expect(screen.queryByText(/; 5c1aa9.../)).not.toBeInTheDocument();
    expect(
      screen.getByTestId('Filter__jobIds1dc67e...Chip'),
    ).toBeInTheDocument();

    await click(screen.getByLabelText('a44234...'));

    expect(screen.getByText(/; 1dc67e.../)).toBeInTheDocument();
    expect(screen.getByText(/; a44234.../)).toBeInTheDocument();
    expect(screen.queryByText(/; 5c1aa9.../)).not.toBeInTheDocument();
    expect(
      screen.getByTestId('Filter__jobIdsa44234...Chip'),
    ).toBeInTheDocument();

    await click(screen.getByText('Select steps'));
    await click(screen.getByText('edges'));

    expect(screen.getByText(/Runtimes for 1 pipeline/)).toBeInTheDocument();
    expect(
      screen.getByTestId('Filter__pipelineStepsedgesChip'),
    ).toBeInTheDocument();
  });

  it('should query for up to x jobs', async () => {
    render(<RuntimesChart />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RuntimesChart__loadingDots'),
    );

    expect(screen.queryByText(/; bc322d.../)).not.toBeInTheDocument();
    expect(screen.queryByText(/; cf302e.../)).not.toBeInTheDocument();

    await click(screen.getAllByText('Last 3 Jobs')[0]);
    await click(screen.getByText('Last 6 Jobs'));

    expect(screen.getByText(/; bc322d.../)).toBeInTheDocument();
    expect(screen.getByText(/; cf302e.../)).toBeInTheDocument();
  });

  it('should show a failed datums job sidebar and link to datum logs', async () => {
    render(<RuntimesChart />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RuntimesChart__loadingDots'),
    );

    await click(screen.getByRole('button', {name: 'Jobs with failed datums'}));

    const sidebar = screen.getByRole('region', {name: 'project-sidebar'});
    const jobItem = within(sidebar).getByLabelText(
      'a4423427351e42aabc40c1031928628e failed datums',
    );
    expect(within(jobItem).getByText('a44234; @edges')).toBeInTheDocument();
    expect(
      within(jobItem).getByText('1 Failed Datum; 2 s Runtime'),
    ).toBeInTheDocument();
    act(() => within(jobItem).getByRole('button').click());
    expect(window.location.pathname).toBe(
      '/project/default/pipelines/edges/jobs/a4423427351e42aabc40c1031928628e/logs/datum',
    );
  });
});
