import {render, screen, within} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {Permission} from '@dash-frontend/api/auth';
import {
  mockPipelinesEmpty,
  mockPipelines,
  mockRepos,
  mockTrueGetAuthorize,
  mockFalseGetAuthorize,
} from '@dash-frontend/mocks';
import {withContextProviders, click, hover} from '@dash-frontend/testHelpers';

import RepoActionsMenuComponent from '../RepoActionsMenu';

describe('Repo Actions Menu', () => {
  const server = setupServer();

  const RepoActionsMenu = withContextProviders(() => (
    <RepoActionsMenuComponent repoId="montage" />
  ));

  beforeAll(() => server.listen());

  beforeEach(() => {
    window.history.replaceState('', '', '/lineage/default/repos/montage');
    server.resetHandlers();
    server.use(mockPipelinesEmpty());
    server.use(mockRepos());
    server.use(
      mockTrueGetAuthorize([Permission.REPO_DELETE, Permission.REPO_WRITE]),
    );
  });

  afterAll(() => server.close());

  it('should allow users to view a repo in the DAG', async () => {
    window.history.replaceState('', '', '/project/default/repos');
    render(<RepoActionsMenu />);

    await click(screen.getByRole('button', {name: 'Repo Actions'}));
    await click(
      await screen.findByRole('menuitem', {
        name: /view in dag/i,
      }),
    );

    expect(window.location.pathname).toBe('/lineage/default/repos/montage');
  });

  it('should not allow users to view a repo in the DAG if already in the DAG', async () => {
    render(<RepoActionsMenu />);

    await click(screen.getByRole('button', {name: 'Repo Actions'}));
    expect(
      screen.queryByRole('menuitem', {
        name: /view in dag/i,
      }),
    ).not.toBeInTheDocument();
  });

  it('should allow users to open the file upload modal', async () => {
    render(<RepoActionsMenu />);

    await click(screen.getByRole('button', {name: 'Repo Actions'}));
    await click(
      await screen.findByRole('menuitem', {
        name: /upload files/i,
      }),
    );

    expect(window.location.pathname).toBe(
      '/lineage/default/repos/montage/upload',
    );
  });

  it('should allow users to open the delete repo modal', async () => {
    render(<RepoActionsMenu />);

    await click(screen.getByRole('button', {name: 'Repo Actions'}));
    await click(
      await screen.findByRole('menuitem', {
        name: /delete repo/i,
      }),
    );

    expect(
      within(await screen.findByRole('dialog')).getByRole('heading', {
        name: 'Are you sure you want to delete this Repo?',
      }),
    ).toBeInTheDocument();
  });

  it('should disable the delete button when there are associated pipelines', async () => {
    server.use(mockPipelines());
    render(<RepoActionsMenu />);

    await click(screen.getByRole('button', {name: 'Repo Actions'}));

    const deleteButton = await screen.findByRole('menuitem', {
      name: /delete repo/i,
    });
    expect(deleteButton).toBeDisabled();
    await hover(deleteButton);
    expect(
      screen.getByText(
        "This repo can't be deleted while it has associated pipelines.",
      ),
    ).toBeInTheDocument();
  });

  it('should disable repo actions without proper permissions', async () => {
    server.use(mockFalseGetAuthorize());
    render(<RepoActionsMenu />);

    await click(screen.getByRole('button', {name: 'Repo Actions'}));

    const deleteButton = await screen.findByRole('menuitem', {
      name: /delete repo/i,
    });
    expect(deleteButton).toBeDisabled();
    await hover(deleteButton);
    expect(
      screen.getByText('You need at least repoOwner to delete this.'),
    ).toBeInTheDocument();

    const uploadButton = await screen.findByRole('menuitem', {
      name: /upload files/i,
    });
    expect(uploadButton).toBeDisabled();
    await hover(uploadButton);
    expect(
      screen.getByText('You need at least repoWriter to upload files.'),
    ).toBeInTheDocument();
  });
});
