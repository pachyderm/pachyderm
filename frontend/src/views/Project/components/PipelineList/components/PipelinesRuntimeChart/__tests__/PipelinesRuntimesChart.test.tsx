import {
  render,
  screen,
  within,
  waitForElementToBeRemoved,
} from '@testing-library/react';
import React from 'react';

import {withContextProviders, click} from '@dash-frontend/testHelpers';

import RuntimesChartComponent from '../PipelinesRuntimeChart';

describe('Pipelines Runtimes Chart', () => {
  const RuntimesChart = withContextProviders(() => {
    return (
      <RuntimesChartComponent
        filtersExpanded
        selectedPipelines={['edges', 'montage']}
      />
    );
  });

  beforeEach(() => {
    window.history.replaceState(
      '',
      '',
      '/project/Solar-Panel-Data-Sorting/pipelines/runtimes',
    );
  });

  it('should display datasets based on given jobs', async () => {
    render(<RuntimesChart />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RuntimesChart__loadingDots'),
    );

    expect(screen.getByText(/; 23b9af.../)).toBeInTheDocument();
    expect(screen.getByText(/; 33b9af.../)).toBeInTheDocument();
    expect(screen.getByText(/; 7798fh.../)).toBeInTheDocument();

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
    await click(screen.getByLabelText('23b9af...'));

    expect(screen.getByText(/; 23b9af.../)).toBeInTheDocument();
    expect(screen.queryByText(/; 33b9af.../)).not.toBeInTheDocument();
    expect(screen.queryByText(/; 7798fh.../)).not.toBeInTheDocument();

    await click(screen.getByLabelText('23b9af...'));
    await click(screen.getByText('Select steps'));
    await click(screen.getByText('edges'));

    expect(screen.getByText(/; 23b9af.../)).toBeInTheDocument();
    expect(screen.queryByText(/; 33b9af.../)).not.toBeInTheDocument();
    expect(screen.queryByText(/; 7798fh.../)).not.toBeInTheDocument();
  });

  it('should query for up to x jobs', async () => {
    render(<RuntimesChart />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RuntimesChart__loadingDots'),
    );

    expect(screen.queryByText(/; o90du4.../)).not.toBeInTheDocument();

    await click(screen.getAllByText('Last 3 Jobs')[0]);
    await click(screen.getByText('Last 6 Jobs'));

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RuntimesChart__loadingDots'),
    );

    expect(screen.getByText(/; o90du4.../)).toBeInTheDocument();
  });

  it('should show a failed datums job sidebar and link to datum logs', async () => {
    render(<RuntimesChart />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RuntimesChart__loadingDots'),
    );

    await click(screen.getByRole('button', {name: 'Jobs with failed datums'}));

    const jobItem = screen.getByLabelText(
      '23b9af7d5d4343219bc8e02ff44cd55a failed datums',
    );

    expect(within(jobItem).getByText('23b9af; @montage')).toBeInTheDocument();
    expect(
      within(jobItem).getByText('1 Failed Datum; 3 s Runtime'),
    ).toBeInTheDocument();
    within(jobItem).getByRole('button').click();
    expect(window.location.pathname).toBe(
      '/project/Solar-Panel-Data-Sorting/pipelines/montage/jobs/23b9af7d5d4343219bc8e02ff44cd55a/logs/datum',
    );
  });
});
