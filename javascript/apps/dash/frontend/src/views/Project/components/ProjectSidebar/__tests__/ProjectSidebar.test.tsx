import {render} from '@testing-library/react';
import React from 'react';
import {Route} from 'react-router';

import {withContextProviders} from '@dash-frontend/testHelpers';

import ProjectSidebar from '../ProjectSidebar';

describe('ProjectSidebar', () => {
  const Project = withContextProviders(() => {
    return <Route path="/project/:projectId/jobs" component={ProjectSidebar} />;
  });

  beforeEach(() => {
    window.history.replaceState('', '', '/');
  });

  it('should display job list', async () => {
    window.history.replaceState('', '', '/project/1/jobs');

    const {queryByTestId, findByTestId} = render(<Project />);

    expect(queryByTestId('JobListSkeleton__list')).toBeInTheDocument();
    expect(await findByTestId('JobList__project1')).toBeInTheDocument();
  });

  it('should not display the job list if not on jobs route', async () => {
    window.history.replaceState('', '', '/project/1');

    const {queryByTestId} = render(<Project />);

    expect(queryByTestId('JobListSkeleton__list')).not.toBeInTheDocument();
  });
});
