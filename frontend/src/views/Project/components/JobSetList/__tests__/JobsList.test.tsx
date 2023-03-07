import {
  render,
  waitForElementToBeRemoved,
  screen,
} from '@testing-library/react';
import React from 'react';

import {withContextProviders, click} from '@dash-frontend/testHelpers';

import JobSetListComponent from '../JobSetList';

describe('JobSet SubJobs List', () => {
  const JobSetList = withContextProviders(() => {
    return <JobSetListComponent />;
  });

  beforeEach(() => {
    window.history.replaceState(
      '',
      '',
      '/project/Solar-Panel-Data-Sorting/jobs/subjobs',
    );
  });

  it('should display jobs details', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('JobsList__loadingDots'),
    );

    const job = screen.getAllByTestId('JobsList__row')[0];
    expect(job).toHaveTextContent('@montage');
    expect(job).toHaveTextContent('4 Total');
    expect(job).toHaveTextContent('23b9af...');
    expect(job).toHaveTextContent('4 seconds');
    expect(job).toHaveTextContent('0 B');
  });

  it('should apply a globalId and redirect', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('JobsList__loadingDots'),
    );

    await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
    await click(
      (
        await screen.findAllByText('Apply Global ID and view in DAG')
      )[0],
    );

    expect(window.location.search).toBe(
      '?globalIdFilter=23b9af7d5d4343219bc8e02ff44cd55a',
    );
  });

  it('should redirect to datum logs', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('JobsList__loadingDots'),
    );

    await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
    await click((await screen.findAllByText('Inspect job'))[0]);

    expect(window.location.pathname).toBe(
      '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
    );
  });

  it('should sort jobs based on creation time', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('JobsList__loadingDots'),
    );

    let jobs = screen.getAllByTestId('JobsList__row');
    expect(jobs[0]).toHaveTextContent('23b9af...');
    expect(jobs[1]).toHaveTextContent('23b9af...');
    expect(jobs[2]).toHaveTextContent('33b9af...');
    expect(jobs[3]).toHaveTextContent('7798fh...');
    expect(jobs[4]).toHaveTextContent('o90du4...');

    await click(screen.getByLabelText('expand filters'));
    await click(
      screen.getByRole('radio', {name: 'Created: Oldest', hidden: false}),
    );

    jobs = screen.getAllByTestId('JobsList__row');
    expect(jobs[0]).toHaveTextContent('o90du4...');
    expect(jobs[1]).toHaveTextContent('7798fh...');
    expect(jobs[2]).toHaveTextContent('23b9af...');
    expect(jobs[3]).toHaveTextContent('33b9af...');
    expect(jobs[4]).toHaveTextContent('23b9af...');
  });

  it('should filter jobs by job status', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('JobsList__loadingDots'),
    );

    expect(await screen.findAllByTestId('JobsList__row')).toHaveLength(5);

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByRole('checkbox', {name: 'Failed', hidden: false}));
    let jobs = screen.getAllByTestId('JobsList__row');
    expect(jobs).toHaveLength(2);
    expect(jobs[0]).toHaveTextContent('33b9af...');
    expect(jobs[1]).toHaveTextContent('o90du4...');

    await click(screen.getByRole('checkbox', {name: 'Success', hidden: false}));
    await click(screen.getByTestId('Filter__jobStatusFailedChip'));
    jobs = screen.getAllByTestId('JobsList__row');
    expect(jobs).toHaveLength(2);
    expect(jobs[0]).toHaveTextContent('23b9af...');
    expect(jobs[1]).toHaveTextContent('23b9af...');
  });

  it('should filter jobs by ID or Pipeline step', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('JobsList__loadingDots'),
    );

    expect(await screen.findAllByTestId('JobsList__row')).toHaveLength(5);

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByText('Select job IDs'));
    await click(screen.getByLabelText('23b9af...'));

    let jobs = await screen.findAllByTestId('JobsList__row');
    expect(jobs).toHaveLength(2);
    expect(jobs[0]).toHaveTextContent('23b9af...');

    await click(screen.getByText('Select steps'));
    await click(screen.getByLabelText('montage'));

    jobs = screen.getAllByTestId('JobsList__row');
    expect(jobs).toHaveLength(1);
    expect(jobs[0]).toHaveTextContent('@montage');
  });
});
