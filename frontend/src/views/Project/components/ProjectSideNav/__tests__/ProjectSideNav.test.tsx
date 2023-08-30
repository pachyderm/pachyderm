import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockEmptyGetAuthorize,
  mockGetVersionInfo,
  mockRepos,
} from '@dash-frontend/mocks';
import {withContextProviders, click, type} from '@dash-frontend/testHelpers';

import ProjectSideNavComponent from '../../ProjectSideNav';

describe('project sidenav', () => {
  const server = setupServer();

  const ProjectSideNav = withContextProviders(ProjectSideNavComponent);

  beforeAll(() => {
    server.listen();
    server.use(mockEmptyGetAuthorize());
    server.use(mockGetVersionInfo());
    server.use(mockRepos());
  });

  afterAll(() => server.close());

  describe('create repo modal', () => {
    it('should error if the repo already exists', async () => {
      window.history.replaceState('', '', '/project/default');

      render(<ProjectSideNav />);

      const createButton = screen.getByText('Create Repo');
      await click(createButton);

      const nameInput = await screen.findByLabelText('Name', {
        exact: false,
      });

      await type(nameInput, 'montage');

      expect(
        await screen.findByText('Repo name already in use'),
      ).toBeInTheDocument();
    });
  });
});
