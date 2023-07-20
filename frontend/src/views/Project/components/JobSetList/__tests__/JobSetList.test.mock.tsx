import {
  render,
  waitForElementToBeRemoved,
  screen,
} from '@testing-library/react';
import React from 'react';

import {withContextProviders, click} from '@dash-frontend/testHelpers';

import JobSetListComponent from '../JobSetList';

// jobSet fixture for Trait-Discovery does not match jobs fixture. Jobs fixture is empty. Rewrite these with MSW. TODO
describe.skip('JobSet Jobs List', () => {
  const JobSetList = withContextProviders(() => {
    return <JobSetListComponent />;
  });

  beforeEach(() => {
    window.history.replaceState('', '', '/project/Trait-Discovery/jobs');
  });

  it('should display run details', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RunsTable__loadingDots'),
    );

    const jobs = screen.getAllByTestId('RunsList__row');
    expect(jobs[1]).toHaveTextContent('23b9af7d5d4343219bc8e02ff44cd55a');
    expect(jobs[1]).toHaveTextContent('27 d 20 h 35 mins');
    expect(jobs[1]).toHaveTextContent('2 Jobs Total');
  });

  it('should allow the user to select a subset of jobs', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RunsTable__loadingDots'),
    );

    expect(
      screen.getByText('Select rows to view detailed info'),
    ).toBeInTheDocument();
    expect(await screen.findAllByTestId('RunsList__row')).toHaveLength(4);

    await click(screen.getByText('23b9af7d5d4343219bc8e02ff44cd55a'));
    await click(screen.getByText('7798fhje5d4343219bc8e02ff4acd33a'));

    expect(screen.getByText('Detailed info for 2 jobs')).toBeInTheDocument();
  });

  it('should apply a globalId and redirect', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RunsTable__loadingDots'),
    );

    await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
    await click(
      (
        await screen.findAllByText('Apply Global ID and view in DAG')
      )[0],
    );

    expect(window.location.search).toBe(
      '?globalIdFilter=33b9af7d5d4343219bc8e02ff44cd55a',
    );
  });

  it('should sort jobs based on creation time', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RunsTable__loadingDots'),
    );

    let jobs = screen.getAllByTestId('RunsList__row');
    expect(jobs[0]).toHaveTextContent('33b9af7d5d4343219bc8e02ff44cd55a');
    expect(jobs[1]).toHaveTextContent('23b9af7d5d4343219bc8e02ff44cd55a');
    expect(jobs[2]).toHaveTextContent('7798fhje5d4343219bc8e02ff4acd33a');
    expect(jobs[3]).toHaveTextContent('o90du4js5d4343219bc8e02ff4acd33a');

    await click(screen.getByLabelText('expand filters'));
    await click(
      screen.getByRole('radio', {name: 'Created: Oldest', hidden: false}),
    );

    jobs = screen.getAllByTestId('RunsList__row');
    expect(jobs[0]).toHaveTextContent('o90du4js5d4343219bc8e02ff4acd33a');
    expect(jobs[1]).toHaveTextContent('7798fhje5d4343219bc8e02ff4acd33a');
    expect(jobs[2]).toHaveTextContent('23b9af7d5d4343219bc8e02ff44cd55a');
    expect(jobs[3]).toHaveTextContent('33b9af7d5d4343219bc8e02ff44cd55a');
  });

  it('should filter jobs by job status', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RunsTable__loadingDots'),
    );

    expect(await screen.findAllByTestId('RunsList__row')).toHaveLength(4);

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByRole('checkbox', {name: 'Failed', hidden: false}));
    let jobs = screen.getAllByTestId('RunsList__row');
    expect(jobs).toHaveLength(2);
    expect(jobs[0]).toHaveTextContent('33b9af7d5d4343219bc8e02ff44cd55a');
    expect(jobs[1]).toHaveTextContent('o90du4js5d4343219bc8e02ff4acd33a');

    await click(screen.getByRole('checkbox', {name: 'Running', hidden: false}));
    await click(screen.getByTestId('Filter__jobsetStatusFailedChip'));
    jobs = screen.getAllByTestId('RunsList__row');
    expect(jobs).toHaveLength(1);
    expect(jobs[0]).toHaveTextContent('7798fhje5d4343219bc8e02ff4acd33a');

    await click(screen.getByRole('checkbox', {name: 'Failed', hidden: false}));
    await click(screen.getByRole('checkbox', {name: 'Success', hidden: false}));
    expect(await screen.findAllByTestId('RunsList__row')).toHaveLength(4);
  });
});
