import jobSets from '@dash-backend/mock/fixtures/jobSets';
import {render, waitFor, within, screen} from '@testing-library/react';
import React from 'react';

import {useProjectDetails} from '@dash-frontend/hooks/useProjectDetails';
import {withContextProviders} from '@dash-frontend/testHelpers';

import ProjectJobSetListComponent from '../ProjectJobSetList';

describe('ProjectJobSetList', () => {
  const ProjectJobSetList = withContextProviders(
    ({projectId}: {projectId: string}) => {
      const {projectDetails, error, loading} = useProjectDetails(projectId);
      return (
        <ProjectJobSetListComponent
          jobs={projectDetails?.jobSets}
          loading={loading}
          error={error}
          projectId={projectId}
          emptyStateTitle="Empty State Title"
          emptyStateMessage="Empty State Message"
        />
      );
    },
  );

  afterEach(() => {
    window.history.pushState({}, document.title, '/');
  });

  it('should display the list of jobs for a project', async () => {
    render(<ProjectJobSetList projectId="Solar-Panel-Data-Sorting" />);

    await waitFor(() =>
      expect(
        screen.queryByTestId('ProjectJobSetList__loadingdots'),
      ).not.toBeInTheDocument(),
    );

    const {queryAllByRole, getByText} = within(screen.getByRole('list'));

    expect(queryAllByRole('listitem')).toHaveLength(
      Object.keys(jobSets['Solar-Panel-Data-Sorting']).length,
    );
    expect(getByText('Success')).toBeInTheDocument();
    expect(getByText('Failure')).toBeInTheDocument();
    expect(getByText('Finishing')).toBeInTheDocument();
    expect(getByText('Killed')).toBeInTheDocument();
  });

  it('should display an empty state if there are no jobs', async () => {
    render(<ProjectJobSetList projectId="Egress-Examples" />);

    expect(await screen.findByText('Empty State Title')).toBeInTheDocument();
    expect(screen.queryByRole('list')).not.toBeInTheDocument();
  });
});
