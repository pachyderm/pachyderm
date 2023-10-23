import {
  render,
  screen,
  waitForElementToBeRemoved,
  within,
} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockReposWithCommit,
  mockEmptyGetAuthorize,
  mockTrueGetAuthorize,
  mockEmptyGetRoles,
} from '@dash-frontend/mocks';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import RepoListComponent from '../RepoList';

describe('Repo List', () => {
  const server = setupServer();

  const RepoList = withContextProviders(() => {
    return <RepoListComponent />;
  });

  beforeAll(() => {
    server.listen();
    server.use(mockEmptyGetAuthorize());
    server.use(mockReposWithCommit());
  });

  beforeEach(() => {
    window.history.replaceState('', '', '/project/default/repos');
  });

  afterAll(() => server.close());

  it('should display repo details', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('ReposTable__loadingDots'),
    );

    const repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[1]).toHaveTextContent('edges');
    expect(repos[1]).toHaveTextContent('22.79 kB');
    expect(repos[1]).toHaveTextContent('Jul 24, 2023; 17:58');
    expect(repos[1]).toHaveTextContent('544e62..');
    expect(repos[1]).toHaveTextContent(
      'Output repo for pipeline default/edges.',
    );
    expect(repos[1]).toHaveTextContent('clusterAdmin');

    expect(repos[3]).toHaveTextContent('empty');
    expect(repos[3]).toHaveTextContent('Jul 24, 2023; 17:58');
    expect(repos[3]).toHaveTextContent('0 B');
    expect(repos[3]).toHaveTextContent('N/A');
    expect(repos[3]).toHaveTextContent('clusterAdmin, repoOwner');
  });

  it('should allow the user to select a repo', async () => {
    render(<RepoList />);

    expect(
      screen.getByText('Select a row to view detailed info'),
    ).toBeInTheDocument();

    expect(await screen.findAllByTestId('RepoListRow__row')).toHaveLength(4);
    await click(screen.getByText('edges'));

    expect(screen.getByText('Detailed info for edges')).toBeInTheDocument();
  });

  it('should sort repos', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('ReposTable__loadingDots'),
    );

    const expand = screen.getByRole('button', {
      name: /expand filters/i,
    });

    let repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[0]).toHaveTextContent('montage');
    expect(repos[1]).toHaveTextContent('edges');
    expect(repos[2]).toHaveTextContent('images');
    expect(repos[3]).toHaveTextContent('empty');

    await click(expand);
    await click(screen.getByRole('radio', {name: /Created: Oldest/}));

    repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[0]).toHaveTextContent('empty');
    expect(repos[1]).toHaveTextContent('images');
    expect(repos[2]).toHaveTextContent('edges');
    expect(repos[3]).toHaveTextContent('montage');

    await click(expand);
    await click(screen.getByRole('radio', {name: /Alphabetical: A-Z/}));

    repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[0]).toHaveTextContent('edges');
    expect(repos[1]).toHaveTextContent('empty');
    expect(repos[2]).toHaveTextContent('images');
    expect(repos[3]).toHaveTextContent('montage');

    await click(expand);
    await click(screen.getByRole('radio', {name: /Alphabetical: Z-A/}));

    repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[0]).toHaveTextContent('montage');
    expect(repos[1]).toHaveTextContent('images');
    expect(repos[2]).toHaveTextContent('empty');
    expect(repos[3]).toHaveTextContent('edges');

    await click(expand);
    await click(screen.getByRole('radio', {name: /Size/}));

    repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[0]).toHaveTextContent('empty');
    expect(repos[1]).toHaveTextContent('images');
    expect(repos[2]).toHaveTextContent('edges');
    expect(repos[3]).toHaveTextContent('montage');
  });

  it('should allow users to inspect latest commit in repo', async () => {
    render(<RepoList />);

    await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
    await click(
      await screen.findByRole('menuitem', {
        name: /inspect commits/i,
      }),
    );

    expect(window.location.pathname).toBe(
      '/project/default/repos/montage/branch/master/commit/544e62c9a16f4a0aa995eb993ded1e52/',
    );
  });

  it('should allow users to view a repo in the DAG', async () => {
    render(<RepoList />);

    await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
    await click(
      await screen.findByRole('menuitem', {
        name: /view in dag/i,
      }),
    );

    expect(window.location.pathname).toBe('/lineage/default/repos/montage');
  });

  describe('with repo edit permission', () => {
    beforeAll(() => {
      server.use(mockTrueGetAuthorize());
      server.use(mockEmptyGetRoles());
    });

    it('should allow users to open the roles modal', async () => {
      render(<RepoList />);

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      await click(
        await screen.findByRole('menuitem', {
          name: /set roles/i,
        }),
      );

      const modal = await screen.findByRole('dialog');
      expect(modal).toBeInTheDocument();

      expect(
        within(modal).getByRole('heading', {
          name: 'Set Repo Level Roles: default/montage',
        }),
      ).toBeInTheDocument();
    });
  });
});
