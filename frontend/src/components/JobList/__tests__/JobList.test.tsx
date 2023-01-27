import jobs from '@dash-backend/mock/fixtures/jobs';
import jobSets from '@dash-backend/mock/fixtures/jobSets';
import {render, waitFor, within, screen} from '@testing-library/react';
import React from 'react';

import {useJobs} from '@dash-frontend/hooks/useJobs';
import {useJobSets} from '@dash-frontend/hooks/useJobSets';
import {click, withContextProviders} from '@dash-frontend/testHelpers';

import JobListComponent, {JobListProps} from '../JobList';

describe('JobList', () => {
  const JobList = withContextProviders(
    ({
      projectId,
      pipelineId,
      ...rest
    }: Omit<JobListProps, 'jobs' | 'loading'> & {pipelineId?: string}) => {
      const {jobs, loading} = useJobs({projectId, pipelineId});

      return (
        <JobListComponent
          jobs={jobs}
          loading={loading}
          projectId={projectId}
          {...rest}
        />
      );
    },
  );

  const JobSetList = withContextProviders(
    ({projectId, ...rest}: Omit<JobListProps, 'jobs' | 'loading'>) => {
      const {jobSets, loading} = useJobSets({projectId});
      return (
        <JobListComponent
          jobs={jobSets}
          loading={loading}
          projectId={projectId}
          {...rest}
        />
      );
    },
  );

  afterEach(() => {
    window.history.pushState({}, document.title, '/');
  });

  it('should display the list of jobs for a project', async () => {
    render(
      <JobList
        projectId="Data-Cleaning-Process"
        emptyStateTitle="Empty State Title"
        emptyStatemessage="Empty State Message"
      />,
    );

    await waitFor(() =>
      expect(
        screen.queryByTestId('JobListStatic__loadingdots'),
      ).not.toBeInTheDocument(),
    );

    const {queryAllByRole, getByText, queryByText} = within(
      screen.getByRole('list'),
    );

    expect(queryAllByRole('listitem')).toHaveLength(
      jobs['Data-Cleaning-Process'].length,
    );
    expect(getByText('Failure')).toBeInTheDocument();
    expect(getByText('Egressing')).toBeInTheDocument();
    expect(getByText('Killed')).toBeInTheDocument();
    expect(getByText('Running')).toBeInTheDocument();
    expect(getByText('Starting')).toBeInTheDocument();
    expect(queryByText('See Details')).not.toBeInTheDocument();
  });

  it('should display a list of jobs for a pipeline', async () => {
    render(
      <JobList
        projectId="Solar-Panel-Data-Sorting"
        pipelineId="montage"
        emptyStateTitle="Empty State Title"
        emptyStatemessage="Empty State Message"
      />,
    );

    await waitFor(() =>
      expect(
        screen.queryByTestId('JobListStatic__loadingdots'),
      ).not.toBeInTheDocument(),
    );

    const {queryAllByRole, getByText} = within(screen.getByRole('list'));

    expect(queryAllByRole('listitem')).toHaveLength(
      jobs['Solar-Panel-Data-Sorting'].filter(
        (job) => job.getJob()?.getPipeline()?.getName() === 'montage',
      ).length,
    );
    expect(getByText('Success')).toBeInTheDocument();
  });

  it('should display a list of actions', async () => {
    render(
      <JobList
        projectId="Data-Cleaning-Process"
        expandActions
        emptyStateTitle="Empty State Title"
        emptyStatemessage="Empty State Message"
      />,
    );

    const seeDetailsButtons = await screen.findAllByText('See Details');

    expect(seeDetailsButtons).toHaveLength(
      jobs['Data-Cleaning-Process'].length,
    );
  });

  it('should allow user to filter on job state', async () => {
    render(
      <JobList
        projectId="Data-Cleaning-Process"
        showStatusFilter
        emptyStateTitle="Empty State Title"
        emptyStatemessage="Empty State Message"
      />,
    );

    expect(
      await screen.findByText(
        `Last ${jobs['Data-Cleaning-Process'].length} Jobs`,
      ),
    ).toBeInTheDocument();

    const startingButton = screen.getByText(/Starting \(\d\)/);
    const runningButton = screen.getByText(/Running \(\d\)/);
    const failureButton = screen.getByText(/Failure \(\d\)/);
    const killedButton = screen.getByText(/Killed \(\d\)/);
    const egressingButton = screen.getByText(/Egressing \(\d\)/);

    const {queryByText: queryByTextWithinList} = within(
      screen.getByRole('list'),
    );

    await click(startingButton);
    expect(queryByTextWithinList('Starting')).not.toBeInTheDocument();

    await click(runningButton);
    expect(queryByTextWithinList('Starting')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Running')).not.toBeInTheDocument();

    await click(failureButton);
    expect(queryByTextWithinList('Starting')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Running')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Failure')).not.toBeInTheDocument();

    await click(killedButton);
    expect(queryByTextWithinList('Starting')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Running')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Failure')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Killed')).not.toBeInTheDocument();

    await click(egressingButton);
    expect(queryByTextWithinList('Starting')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Running')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Failure')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Killed')).not.toBeInTheDocument();
    expect(queryByTextWithinList('Egressing')).not.toBeInTheDocument();
  });

  it('should display the filter empty state message if no filters are selected', async () => {
    render(
      <JobList
        projectId="Solar-Power-Data-Logger-Team-Collab"
        showStatusFilter
        emptyStateTitle="Empty State Title"
        emptyStatemessage="Empty State Message"
      />,
    );

    expect(
      await screen.findByText(
        `Last ${jobs['Solar-Power-Data-Logger-Team-Collab'].length} Jobs`,
      ),
    ).toBeInTheDocument();

    const successButton = screen.getByText(/Success \(\d\)/);
    await click(successButton);

    const failureButton = screen.getByText(/Failure \(\d\)/);
    await click(failureButton);

    const killedButton = screen.getByText(/Killed \(\d\)/);
    await click(killedButton);

    expect(
      await screen.findByText('Select Job Filters Above :)'),
    ).toBeInTheDocument();
  });

  it('should display an empty state if there are no jobs', async () => {
    render(
      <JobList
        projectId="Egress-Examples"
        emptyStateTitle="Empty State Title"
        emptyStatemessage="Empty State Message"
      />,
    );

    expect(await screen.findByText('Empty State Title')).toBeInTheDocument();
    expect(screen.queryByRole('list')).not.toBeInTheDocument();
  });

  describe('JobSets', () => {
    it('should display the list of jobSets for a given project', async () => {
      render(
        <JobSetList
          projectId="Data-Cleaning-Process"
          emptyStateTitle="Empty State Title"
          emptyStateMessage="Empty State Message"
        />,
      );

      await waitFor(() =>
        expect(
          screen.queryByTestId('JobListStatic__loadingdots'),
        ).not.toBeInTheDocument(),
      );

      const {queryAllByRole} = within(screen.getByRole('list'));

      expect(queryAllByRole('listitem')).toHaveLength(
        Object.keys(jobSets['Data-Cleaning-Process']).length,
      );
    });
  });
});
