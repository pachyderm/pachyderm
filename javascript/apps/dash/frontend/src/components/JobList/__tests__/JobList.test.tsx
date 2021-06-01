import {render, waitFor, within} from '@testing-library/react';
import React from 'react';

import pipelineJobs from '@dash-backend/mock/fixtures/pipelineJobs';
import {click, withContextProviders} from '@dash-frontend/testHelpers';

import JobListComponent from '../JobList';

describe('JobList', () => {
  const JobList = withContextProviders(JobListComponent);

  afterEach(() => {
    window.history.pushState({}, document.title, '/');
  });

  it('should display the list of jobs for a project', async () => {
    const {getByRole, queryByTestId} = render(
      <JobList
        projectId="2"
        emptyStateTitle="Empty State Title"
        emptyStatemessage="Empty State Message"
      />,
    );

    await waitFor(() =>
      expect(queryByTestId('JobListSkeleton__list')).not.toBeInTheDocument(),
    );

    const {queryAllByRole, queryByText, queryAllByText} = within(
      getByRole('list'),
    );

    expect(queryAllByRole('listitem').length).toBe(pipelineJobs['2'].length);
    expect(queryByText('Failure')).toBeInTheDocument();
    expect(queryByText('Egressing')).toBeInTheDocument();
    expect(queryByText('Killed')).toBeInTheDocument();
    expect(queryByText('Running')).toBeInTheDocument();
    expect(queryByText('Starting')).toBeInTheDocument();
    expect(queryAllByText('Read logs').length).toBe(0);
    expect(queryAllByText('See Details').length).toBe(0);
  });

  it('should display a list of jobs for a pipeline', async () => {
    const {getByRole, queryByTestId} = render(
      <JobList
        projectId="1"
        pipelineId="montage"
        emptyStateTitle="Empty State Title"
        emptyStatemessage="Empty State Message"
      />,
    );

    await waitFor(() =>
      expect(queryByTestId('JobListSkeleton__list')).not.toBeInTheDocument(),
    );

    const {queryAllByRole, queryByText} = within(getByRole('list'));

    expect(queryAllByRole('listitem').length).toBe(
      pipelineJobs['1'].filter(
        (job) => job.getPipeline()?.getName() === 'montage',
      ).length,
    );
    expect(queryByText('Success')).toBeInTheDocument();
  });

  it('should display a list of actions', async () => {
    const {findAllByText} = render(
      <JobList
        projectId="2"
        expandActions
        emptyStateTitle="Empty State Title"
        emptyStatemessage="Empty State Message"
      />,
    );

    const readLogsButtons = await findAllByText('Read logs');
    const seeDetailsButtons = await findAllByText('See Details');

    expect(readLogsButtons.length).toBe(pipelineJobs['2'].length);
    expect(seeDetailsButtons.length).toBe(pipelineJobs['2'].length);
  });

  it('should allow user to filter on job state', async () => {
    const {getByText, findByText, getByRole} = render(
      <JobList
        projectId="2"
        showStatusFilter
        emptyStateTitle="Empty State Title"
        emptyStatemessage="Empty State Message"
      />,
    );

    expect(
      await findByText(`Last ${pipelineJobs['2'].length} Jobs`),
    ).toBeInTheDocument();

    const startingButton = getByText(/Starting \(\d\)/);
    const runningButton = getByText(/Running \(\d\)/);
    const failureButton = getByText(/Failure \(\d\)/);
    const killedButton = getByText(/Killed \(\d\)/);
    const egressingButton = getByText(/Egressing \(\d\)/);

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

  it('should display the filter empty state message if no filters are selected', async () => {
    const {getByText, findByText} = render(
      <JobList
        projectId="3"
        showStatusFilter
        emptyStateTitle="Empty State Title"
        emptyStatemessage="Empty State Message"
      />,
    );

    expect(
      await findByText(`Last ${pipelineJobs['3'].length} Jobs`),
    ).toBeInTheDocument();

    const successButton = getByText(/Success \(\d\)/);
    click(successButton);

    expect(await findByText('Select Job Filters Above :)')).toBeInTheDocument();
  });

  it('should display an empty state if there are no jobs', async () => {
    const {findByText, queryByRole} = render(
      <JobList
        projectId="5"
        emptyStateTitle="Empty State Title"
        emptyStatemessage="Empty State Message"
      />,
    );

    expect(await findByText('Empty State Title')).toBeInTheDocument();
    expect(queryByRole('list')).not.toBeInTheDocument();
  });
});
