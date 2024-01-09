import {render, screen, waitFor} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {Permission} from '@dash-frontend/api/auth';
import {
  mockRepos,
  mockEmptyGetAuthorize,
  mockFalseGetAuthorize,
  mockTrueGetAuthorize,
  mockGetVersionInfo,
} from '@dash-frontend/mocks';
import {withContextProviders, click, type} from '@dash-frontend/testHelpers';

import ProjectSideNavComponent from '../../ProjectSideNav';

describe('project sidenav', () => {
  const server = setupServer();

  const ProjectSideNav = withContextProviders(ProjectSideNavComponent);

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    server.resetHandlers();
    server.use(mockEmptyGetAuthorize());
    server.use(mockGetVersionInfo());
    server.use(mockRepos());
    window.history.replaceState('', '', '/project/default');
  });

  afterAll(() => server.close());

  describe('Project sidenav', () => {
    it('create repo modal should error if the repo already exists', async () => {
      render(<ProjectSideNav />);

      await click(
        screen.getByRole('button', {
          name: /create/i,
        }),
      );
      await click(
        screen.getByRole('menuitem', {
          name: /input repository/i,
        }),
      );

      const nameInput = await screen.findByLabelText('Name', {
        exact: false,
      });

      await type(nameInput, 'montage');

      expect(
        await screen.findByText('Repo name already in use'),
      ).toBeInTheDocument();
    });

    it('create pipeline should route users to pipeline creation', async () => {
      render(<ProjectSideNav />);

      await click(
        screen.getByRole('button', {
          name: /create/i,
        }),
      );
      await click(
        screen.getByRole('menuitem', {
          name: /pipeline/i,
        }),
      );

      expect(window.location.pathname).toBe('/lineage/default/create/pipeline');
    });

    it('project defaults should route users to pipeline creation', async () => {
      render(<ProjectSideNav />);

      await click(
        screen.getByRole('link', {
          name: /project defaults/i,
        }),
      );

      expect(window.location.pathname).toBe('/project/default/defaults');
    });
  });

  describe('Create Permissions', () => {
    it('appears in CE', async () => {
      render(<ProjectSideNav />);

      expect(
        await screen.findByRole('button', {name: /create/i}),
      ).toBeEnabled();

      expect(
        screen.getByRole('link', {name: /project defaults/i}),
      ).toBeEnabled();
    });

    it('appears as a project writer', async () => {
      server.use(
        mockTrueGetAuthorize([
          Permission.PROJECT_MODIFY_BINDINGS,
          Permission.PROJECT_CREATE_REPO,
          Permission.PROJECT_SET_DEFAULTS,
        ]),
      );
      render(<ProjectSideNav />);

      expect(
        await screen.findByRole('button', {name: /create/i}),
      ).toBeEnabled();

      expect(
        screen.getByRole('link', {name: /project defaults/i}),
      ).toBeEnabled();
    });

    it('hides when not project writer', async () => {
      server.use(mockFalseGetAuthorize());
      render(<ProjectSideNav />);

      await waitFor(() =>
        expect(screen.getByRole('navigation')).toHaveAttribute(
          'aria-busy',
          'false',
        ),
      );

      expect(
        screen.queryByRole('button', {name: /create/i}),
      ).not.toBeInTheDocument();

      expect(
        screen.queryByRole('link', {name: /project defaults/i}),
      ).not.toBeInTheDocument();
    });
  });
});
