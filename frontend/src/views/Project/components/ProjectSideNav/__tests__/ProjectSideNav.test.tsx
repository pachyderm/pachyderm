import {mockServer} from '@dash-backend/testHelpers';
import {render, waitFor, screen} from '@testing-library/react';
import React from 'react';

import {withContextProviders, click, type} from '@dash-frontend/testHelpers';

import ProjectSideNavComponent from '../../ProjectSideNav';

describe('project sidenav', () => {
  const ProjectSideNav = withContextProviders(ProjectSideNavComponent);

  it('should not display notification badge for projects with no unhealthy jobs', async () => {
    window.history.replaceState(
      '',
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab',
    );

    render(<ProjectSideNav />);

    await waitFor(() =>
      expect(
        screen.queryByTestId('ProjectHeader__projectNameLoader'),
      ).not.toBeInTheDocument(),
    );
    expect(screen.queryByLabelText('Number of failed')).not.toBeInTheDocument();
  });

  it('should allow users to create new repos', async () => {
    window.history.replaceState('', '', '/project/Empty-Project');

    render(<ProjectSideNav />);

    const createButton = screen.getByText('Create Repo');
    await click(createButton);

    const nameInput = await screen.findByLabelText('Repo Name', {exact: false});
    const descriptionInput = await screen.findByLabelText(
      'Description (optional)',
      {
        exact: false,
      },
    );
    const submitButton = screen.getByText('Create');

    await type(nameInput, 'newRepo');
    await type(descriptionInput, 'newRepo Description');

    expect(mockServer.getState().repos['Empty-Project']).toHaveLength(0);

    await click(submitButton);

    await waitFor(() =>
      expect(mockServer.getState().repos['Empty-Project']).toHaveLength(1),
    );
  });
});
