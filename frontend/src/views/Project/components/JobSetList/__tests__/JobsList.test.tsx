import {
  render,
  waitForElementToBeRemoved,
  screen,
} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {mockGetAllJobs, mockEmptyJobs} from '@dash-frontend/mocks';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import JobSetListComponent from '../JobSetList';

describe('JobSet SubJobs List', () => {
  const server = setupServer();

  const JobSetList = withContextProviders(() => {
    return <JobSetListComponent />;
  });

  beforeAll(() => {
    server.listen();
    server.use(mockGetAllJobs());
  });

  beforeEach(() => {
    window.history.replaceState('', '', '/project/default/jobs/subjobs');
  });

  afterAll(() => server.close());

  it('should display jobs details', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('JobsList__loadingDots'),
    );

    const job1 = screen.getAllByTestId('JobsList__row')[1];
    expect(job1).toHaveTextContent('@edges');
    expect(job1).toHaveTextContent('v:1');
    expect(job1).toHaveTextContent('0 Total');
    expect(job1).toHaveTextContent('In Progress');
    expect(job1).toHaveTextContent('1dc67e...');
    expect(job1).toHaveTextContent('0 B');

    const job2 = screen.getAllByTestId('JobsList__row')[4];
    expect(job2).toHaveTextContent('@montage');
    expect(job1).toHaveTextContent('v:1');
    expect(job2).toHaveTextContent('1 Total');
    expect(job2).toHaveTextContent('5c1aa9...');
    expect(job2).toHaveTextContent('3 s');
    expect(job2).toHaveTextContent('380.91 kB');
    expect(job2).toHaveTextContent('1.42 MB');
    expect(job2).toHaveTextContent('Aug 1, 2023; 14:20');
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
      '?globalIdFilter=1dc67e479f03498badcc6180be4ee6ce',
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
      '/project/default/jobs/1dc67e479f03498badcc6180be4ee6ce/pipeline/montage/logs',
    );
  });

  it('should sort jobs based on creation time', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('JobsList__loadingDots'),
    );

    let jobs = screen.getAllByTestId('JobsList__row');
    expect(jobs[0]).toHaveTextContent('1dc67e...');
    expect(jobs[1]).toHaveTextContent('1dc67e...');
    expect(jobs[2]).toHaveTextContent('a44234...');
    expect(jobs[3]).toHaveTextContent('a44234...');
    expect(jobs[4]).toHaveTextContent('5c1aa9...');

    await click(screen.getByLabelText('expand filters'));
    await click(
      screen.getByRole('radio', {name: 'Created: Oldest', hidden: false}),
    );

    jobs = screen.getAllByTestId('JobsList__row');
    expect(jobs[0]).toHaveTextContent('cf302e...');
    expect(jobs[1]).toHaveTextContent('bc322d...');
    expect(jobs[2]).toHaveTextContent('5c1aa9...');
    expect(jobs[3]).toHaveTextContent('5c1aa9...');
    expect(jobs[4]).toHaveTextContent('a44234...');
  });

  it('should filter jobs by job status', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('JobsList__loadingDots'),
    );

    expect(await screen.findAllByTestId('JobsList__row')).toHaveLength(8);

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByRole('checkbox', {name: 'Failed', hidden: false}));
    let jobs = screen.getAllByTestId('JobsList__row');
    expect(jobs).toHaveLength(3);
    expect(jobs[0]).toHaveTextContent('1dc67e...');
    expect(jobs[1]).toHaveTextContent('a44234...');
    expect(jobs[2]).toHaveTextContent('a44234...');

    await click(screen.getByRole('checkbox', {name: 'Success', hidden: false}));
    await click(screen.getByTestId('Filter__jobStatusFailedChip'));
    jobs = screen.getAllByTestId('JobsList__row');
    expect(jobs).toHaveLength(4);
    expect(jobs[0]).toHaveTextContent('5c1aa9...');
    expect(jobs[1]).toHaveTextContent('bc322d...');
    expect(jobs[2]).toHaveTextContent('5c1aa9...');
    expect(jobs[3]).toHaveTextContent('cf302e...');
  });

  it('should filter jobs by ID, Pipeline, or Pipeline Version', async () => {
    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('JobsList__loadingDots'),
    );

    expect(await screen.findAllByTestId('JobsList__row')).toHaveLength(8);

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByText('Select job IDs'));
    await click(screen.getByLabelText('a44234...'));

    let jobs = await screen.findAllByTestId('JobsList__row');
    expect(jobs).toHaveLength(2);
    expect(jobs[0]).toHaveTextContent('a44234...');
    expect(jobs[1]).toHaveTextContent('a44234...');

    await click(screen.getByText('Select steps'));
    await click(screen.getByLabelText('montage'));

    jobs = screen.getAllByTestId('JobsList__row');
    expect(jobs).toHaveLength(1);
    expect(jobs[0]).toHaveTextContent('@montage');

    await click(screen.getByTestId('Filter__jobIdsa44234...Chip'));
    await click(screen.getByTestId('Filter__pipelineStepsmontageChip'));

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByText('Select versions'));
    await click(screen.getByLabelText('2'));

    jobs = screen.getAllByTestId('JobsList__row');
    expect(jobs).toHaveLength(2);
    expect(jobs[0]).toHaveTextContent('v:2');
  });

  it('should display an empty state', async () => {
    server.use(mockEmptyJobs());

    render(<JobSetList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('JobsList__loadingDots'),
    );

    expect(
      screen.getByText(
        "We couldn't find any results matching your filters. Try using a different set of filters.",
      ),
    ).toBeInTheDocument();
  });
});
