import {
  render,
  waitForElementToBeRemoved,
  screen,
} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockGetEnterpriseInfoInactive,
  mockGetAllJobs,
  mockEmptyJob,
} from '@dash-frontend/mocks';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import JobSetListComponent from '../JobSetList';

describe('JobSet Jobs List', () => {
  const server = setupServer();

  const JobSetList = withContextProviders(() => {
    return <JobSetListComponent />;
  });

  beforeAll(() => {
    server.listen();
    server.use(mockGetAllJobs());
    server.use(mockGetEnterpriseInfoInactive());
  });

  beforeEach(() => {
    window.history.replaceState('', '', '/project/default/jobs');
  });

  afterAll(() => server.close());

  it('should display run details', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RunsTable__loadingDots'),
    );

    const jobs = screen.getAllByTestId('RunsList__row');
    expect(jobs[1]).toHaveTextContent('a4423427351e42aabc40c1031928628e');
    expect(jobs[1]).toHaveTextContent('22 s');
    expect(jobs[1]).toHaveTextContent('2 Jobs Total');
    expect(jobs[1]).toHaveTextContent('Aug 1, 2023; 14:20');
  });

  it('should allow the user to select a subset of jobs', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RunsTable__loadingDots'),
    );

    expect(
      screen.getByText('Select rows to view detailed info'),
    ).toBeInTheDocument();
    expect(await screen.findAllByTestId('RunsList__row')).toHaveLength(5);

    await click(screen.getByText('a4423427351e42aabc40c1031928628e'));
    await click(screen.getByText('1dc67e479f03498badcc6180be4ee6ce'));

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
      '?globalIdFilter=1dc67e479f03498badcc6180be4ee6ce',
    );
  });

  it('should sort jobs based on creation time', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RunsTable__loadingDots'),
    );

    let jobs = screen.getAllByTestId('RunsList__row');
    expect(jobs[0]).toHaveTextContent('1dc67e479f03498badcc6180be4ee6ce');
    expect(jobs[1]).toHaveTextContent('a4423427351e42aabc40c1031928628e');
    expect(jobs[2]).toHaveTextContent('5c1aa9bc87dd411ba5a1be0c80a3ebc2');
    expect(jobs[3]).toHaveTextContent('bc322db1b24c4d16873e1a4db198b5c9');

    await click(screen.getByLabelText('expand filters'));
    await click(
      screen.getByRole('radio', {name: 'Created: Oldest', hidden: false}),
    );

    jobs = screen.getAllByTestId('RunsList__row');
    expect(jobs[0]).toHaveTextContent('cf302e9203874015be2d453d75864721');
    expect(jobs[1]).toHaveTextContent('5c1aa9bc87dd411ba5a1be0c80a3ebc2');
    expect(jobs[2]).toHaveTextContent('bc322db1b24c4d16873e1a4db198b5c9');
    expect(jobs[3]).toHaveTextContent('a4423427351e42aabc40c1031928628e');
  });

  it('should filter jobs by job status', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RunsTable__loadingDots'),
    );

    expect(await screen.findAllByTestId('RunsList__row')).toHaveLength(5);

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByRole('checkbox', {name: 'Failed', hidden: false}));
    let jobs = screen.getAllByTestId('RunsList__row');
    expect(jobs).toHaveLength(2);
    expect(jobs[0]).toHaveTextContent('1dc67e479f03498badcc6180be4ee6ce');
    expect(jobs[1]).toHaveTextContent('a4423427351e42aabc40c1031928628e');

    await click(screen.getByRole('checkbox', {name: 'Running', hidden: false}));
    await click(screen.getByTestId('Filter__jobsetStatusFailedChip'));
    jobs = screen.getAllByTestId('RunsList__row');
    expect(jobs).toHaveLength(1);
    expect(jobs[0]).toHaveTextContent('1dc67e479f03498badcc6180be4ee6ce');

    await click(screen.getByRole('checkbox', {name: 'Failed', hidden: false}));
    await click(screen.getByRole('checkbox', {name: 'Success', hidden: false}));
    expect(await screen.findAllByTestId('RunsList__row')).toHaveLength(5);
  });

  it('should display an empty state', async () => {
    server.use(mockEmptyJob());

    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('RunsTable__loadingDots'),
    );
    expect(screen.getByText('This DAG has no jobs')).toBeInTheDocument();
    expect(
      screen.getByText(
        "This is normal for new DAGs, but we still wanted to notify you because Pachyderm didn't detect a job on our end.",
      ),
    ).toBeInTheDocument();
  });
});
