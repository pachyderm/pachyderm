import {render, waitFor, within} from '@testing-library/react';
import React from 'react';

import jobs from '@dash-backend/mock/fixtures/jobs';
import {click, withContextProviders} from '@dash-frontend/testHelpers';

import JobListComponent from '../JobList';

describe('JobList', () => {
  const JobList = withContextProviders(JobListComponent);

  it('should display the list of jobs for a project', async () => {
    const {getByRole, queryByTestId} = render(<JobList projectId="2" />);

    await waitFor(() =>
      expect(queryByTestId('JobListSkeleton__list')).not.toBeInTheDocument(),
    );

    const {queryAllByRole, queryByText, queryAllByText} = within(
      getByRole('list'),
    );

    expect(queryAllByRole('listitem').length).toBe(jobs['2'].length);
    expect(queryByText('Failure')).toBeInTheDocument();
    expect(queryByText('Egressing')).toBeInTheDocument();
    expect(queryByText('Killed')).toBeInTheDocument();
    expect(queryByText('Running')).toBeInTheDocument();
    expect(queryByText('Starting')).toBeInTheDocument();
    expect(queryAllByText('Read logs').length).toBe(0);
    expect(queryAllByText('See Details').length).toBe(0);
  });

  it('should display a list of actions', async () => {
    const {findAllByText} = render(<JobList projectId="2" expandActions />);

    const readLogsButtons = await findAllByText('Read logs');
    const seeDetailsButtons = await findAllByText('See Details');

    expect(readLogsButtons.length).toBe(jobs['2'].length);
    expect(seeDetailsButtons.length).toBe(jobs['2'].length);
  });

  it('should allow user to filter on job state', async () => {
    const {getByLabelText, findByText, getByRole} = render(
      <JobList projectId="2" showStatusFilter />,
    );

    expect(
      await findByText(`Last ${jobs['2'].length} Jobs`),
    ).toBeInTheDocument();

    const startingButton = getByLabelText(/Starting/);
    const runningButton = getByLabelText(/Running/);
    const failureButton = getByLabelText(/Failure/);
    const killedButton = getByLabelText(/Killed/);
    const egressingButton = getByLabelText(/Egressing/);

    const {queryByText: queryByTextWithinList} = within(getByRole('list'));

    click(startingButton);
    expect(queryByTextWithinList('Starting')).not.toBeInTheDocument();

    click(runningButton);
    expect(queryByTextWithinList('Starting')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Running')).not.toBeInTheDocument();

    click(failureButton);
    expect(queryByTextWithinList('Starting')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Running')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Failure')).not.toBeInTheDocument();

    click(killedButton);
    expect(queryByTextWithinList('Starting')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Running')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Failure')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Killed')).not.toBeInTheDocument();

    click(egressingButton);
    expect(queryByTextWithinList('Starting')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Running')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Failure')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Killed')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Egressing')).not.toBeInTheDocument();
  });
});
