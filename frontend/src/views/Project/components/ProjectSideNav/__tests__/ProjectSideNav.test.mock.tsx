import {mockServer} from '@dash-backend/testHelpers';
import {render, waitFor, screen, within} from '@testing-library/react';
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

  describe('create repo modal', () => {
    it('should allow users to create new repos', async () => {
      window.history.replaceState('', '', '/project/Empty-Project');

      render(<ProjectSideNav />);

      const createButton = screen.getByText('Create Repo');
      await click(createButton);

      const nameInput = await screen.findByLabelText('Name', {exact: false});
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

    it('should error if the repo already exists', async () => {
      window.history.replaceState('', '', '/project/Data-Cleaning-Process');

      render(<ProjectSideNav />);

      const createButton = screen.getByText('Create Repo');
      await click(createButton);

      const nameInput = await screen.findByLabelText('Name', {
        exact: false,
      });

      await type(nameInput, 'training');

      expect(
        await screen.findByText('Repo name already in use'),
      ).toBeInTheDocument();
    });

    const validInputs = [
      ['good'],
      ['good-'],
      ['good_'],
      ['good1'],
      ['_good'],
      ['a'.repeat(63)],
    ];
    const invalidInputs = [
      [
        'bad repo',
        'Repo name can only contain alphanumeric characters, underscores, and dashes',
      ],
      [
        'bad!',
        'Repo name can only contain alphanumeric characters, underscores, and dashes',
      ],
      [
        'bad.',
        'Repo name can only contain alphanumeric characters, underscores, and dashes',
      ],
      ['a'.repeat(64), 'Repo name exceeds maximum allowed length'],
    ];
    test.each(validInputs)(
      'should not error with a valid repo name (%j)',
      async (input) => {
        window.history.replaceState('', '', '/project/Empty-Project');

        render(<ProjectSideNav />);

        const createButton = screen.getByText('Create Repo');
        await click(createButton);

        const modal = screen.getByRole('dialog');
        const nameInput = await within(modal).findByLabelText('Name', {
          exact: false,
        });

        await type(nameInput, input);

        expect(within(modal).queryByRole('alert')).not.toBeInTheDocument();
      },
    );
    test.each(invalidInputs)(
      'should error with an invalid repo name (%j)',
      async (input, assertionText) => {
        window.history.replaceState('', '', '/project/Empty-Project');

        render(<ProjectSideNav />);

        const createButton = screen.getByText('Create Repo');
        await click(createButton);

        const modal = screen.getByRole('dialog');
        const nameInput = await within(modal).findByLabelText('Name', {
          exact: false,
        });

        await type(nameInput, input);

        expect(within(modal).getByText(assertionText)).toBeInTheDocument();
      },
    );
  });
});
