import {
  render,
  waitForElementToBeRemoved,
  screen,
} from '@testing-library/react';
import React from 'react';

import {withContextProviders, click} from '@dash-frontend/testHelpers';

import RepoListComponent from '../RepoList';

describe('Repo List', () => {
  const RepoList = withContextProviders(() => {
    return <RepoListComponent />;
  });

  beforeEach(() => {
    window.history.replaceState(
      '',
      '',
      '/project/Solar-Panel-Data-Sorting/repos',
    );
  });

  it('should display repo details', async () => {
    window.history.replaceState(
      '',
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos',
    );

    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('ReposTable__loadingDots'),
    );

    const repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[1]).toHaveTextContent('cron');
    expect(repos[1]).toHaveTextContent('621.86 kB');
    expect(repos[1]).toHaveTextContent('9d5daa..');
    expect(repos[1]).toHaveTextContent('cron job');
  });

  it('should allow the user to select a repo', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('ReposTable__loadingDots'),
    );

    expect(
      screen.getByText('Select a row to view detailed info'),
    ).toBeInTheDocument();

    expect(await screen.findAllByTestId('RepoListRow__row')).toHaveLength(3);
    await click(screen.getByText('edges'));

    expect(screen.getByText('Detailed info for edges')).toBeInTheDocument();
  });

  it('should allow users to view a repo in the DAG', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('ReposTable__loadingDots'),
    );

    await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
    await click((await screen.findAllByText('View in DAG'))[0]);

    expect(window.location.pathname).toBe(
      '/lineage/Solar-Panel-Data-Sorting/repos/montage',
    );
  });

  it('should sort repos', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('ReposTable__loadingDots'),
    );

    let repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[0]).toHaveTextContent('montage');
    expect(repos[1]).toHaveTextContent('edges');
    expect(repos[2]).toHaveTextContent('images');

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByText('Created: Oldest'));

    repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[0]).toHaveTextContent('images');
    expect(repos[1]).toHaveTextContent('edges');
    expect(repos[2]).toHaveTextContent('montage');

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByText('Alphabetical: A-Z'));

    repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[0]).toHaveTextContent('edges');
    expect(repos[1]).toHaveTextContent('images');
    expect(repos[2]).toHaveTextContent('montage');
  });

  it('should allow users to inspect latest commit in repo', async () => {
    window.history.replaceState(
      '',
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos',
    );

    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('ReposTable__loadingDots'),
    );

    await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
    await click((await screen.findAllByText('Inspect commits'))[0]);

    expect(window.location.pathname).toBe(
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/processor/branch/master/commit/f4e23cf347c342d98bd9015e4c3ad52a/',
    );
  });
});
