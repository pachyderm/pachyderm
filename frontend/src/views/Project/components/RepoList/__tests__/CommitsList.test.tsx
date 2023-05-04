import {
  render,
  waitForElementToBeRemoved,
  screen,
  within,
} from '@testing-library/react';
import React from 'react';

import {withContextProviders, click} from '@dash-frontend/testHelpers';

import RepoListComponent from '../RepoList';

describe('Repo Commits List', () => {
  const RepoList = withContextProviders(() => {
    return <RepoListComponent />;
  });

  beforeEach(async () => {
    window.history.replaceState(
      '',
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos',
    );
  });

  it('should display commit details', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('ReposTable__loadingDots'),
    );
    await click(screen.getAllByTestId('RepositoriesList__row')[1]);
    await click(screen.getByText('Commits'));
    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('CommitsList__loadingDots'),
    );

    const repos = screen.getAllByTestId('CommitsList__row')[2];
    expect(repos).toHaveTextContent('@cron');
    expect(repos).toHaveTextContent('master');
    expect(repos).toHaveTextContent('44.28 kB');
    expect(repos).toHaveTextContent('added mako');
  });

  it('should allow users to inspect a commit', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('ReposTable__loadingDots'),
    );
    await click(screen.getAllByTestId('RepositoriesList__row')[1]);
    await click(screen.getByText('Commits'));
    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('CommitsList__loadingDots'),
    );

    await click((await screen.findAllByTestId('DropdownButton__button'))[2]);
    await click((await screen.findAllByText('Inspect commit'))[2]);

    expect(window.location.pathname).toBe(
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/',
    );
  });

  it('should apply a globalId and redirect', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('ReposTable__loadingDots'),
    );
    await click(screen.getAllByTestId('RepositoriesList__row')[1]);
    await click(screen.getByText('Commits'));
    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('CommitsList__loadingDots'),
    );

    await click((await screen.findAllByTestId('DropdownButton__button'))[2]);
    await click(
      (
        await screen.findAllByText('Apply Global ID and view in DAG')
      )[2],
    );

    expect(window.location.search).toBe(
      '?globalIdFilter=0918ac9d5daa76b86e3bb5e88e4c43a4',
    );
  });

  it('should allow commit order to be reversed', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('ReposTable__loadingDots'),
    );
    await click(screen.getAllByTestId('RepositoriesList__row')[1]);
    await click(screen.getByText('Commits'));
    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('CommitsList__loadingDots'),
    );

    let commits = screen.getAllByTestId('CommitsList__row');
    expect(commits[0]).toHaveTextContent('9d5daa0918ac4c43a476b86e3bb5e88e');
    expect(commits[1]).toHaveTextContent('0918ac4c43a476b86e3bb5e88e9d5daa');
    expect(commits[2]).toHaveTextContent('0918ac9d5daa76b86e3bb5e88e4c43a4');

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByText('Created: Oldest'));

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('CommitsList__loadingDots'),
    );

    commits = screen.getAllByTestId('CommitsList__row');
    expect(commits[0]).toHaveTextContent('0518ac9d5daa76b86e3bb5e88e4c43a5');
    expect(commits[1]).toHaveTextContent('0218ac9d5daa76b86e3bb5e88e4c43a5');
    expect(commits[2]).toHaveTextContent('0918ac9d5daa76b86e3bb5e88e4c43a5');
  });

  it('should allow users to navigate through paged files', async () => {
    window.history.replaceState(
      '',
      '',
      '/project/Load-Project/repos/commits?selectedRepos=load-repo-0',
    );
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('CommitsList__loadingDots'),
    );

    let commits = screen.getAllByTestId('CommitsList__row');
    expect(commits).toHaveLength(15);
    expect(commits[0]).toHaveTextContent('0-0');
    expect(commits[14]).toHaveTextContent('0-14');

    let pager = screen.getByTestId('Pager__pager');
    expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();
    await click(within(pager).getByTestId('Pager__forward'));

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('CommitsList__loadingDots'),
    );

    commits = screen.getAllByTestId('CommitsList__row');
    expect(commits).toHaveLength(15);
    expect(commits[0]).toHaveTextContent('0-15');
    expect(commits[14]).toHaveTextContent('0-29');

    pager = screen.getByTestId('Pager__pager');
    expect(within(pager).getByTestId('Pager__forward')).toBeEnabled();
    await click(within(pager).getByTestId('Pager__backward'));

    commits = screen.getAllByTestId('CommitsList__row');
    expect(commits).toHaveLength(15);
  });

  it('should allow users to update page size', async () => {
    window.history.replaceState(
      '',
      '',
      '/project/Load-Project/repos/commits?selectedRepos=load-repo-0',
    );
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('CommitsList__loadingDots'),
    );

    let commits = screen.getAllByTestId('CommitsList__row');
    expect(commits).toHaveLength(15);

    const pager = screen.getByTestId('Pager__pager');
    expect(within(pager).getByTestId('Pager__forward')).toBeEnabled();
    expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();

    await click(within(pager).getByTestId('DropdownButton__button'));
    await click(within(pager).getByText(50));

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('CommitsList__loadingDots'),
    );

    commits = screen.getAllByTestId('CommitsList__row');
    expect(commits).toHaveLength(50);
  });
});
