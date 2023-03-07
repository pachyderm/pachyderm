import {
  render,
  waitForElementToBeRemoved,
  screen,
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

    const repos = screen.getAllByTestId('CommitsList__row')[0];
    expect(repos).toHaveTextContent('@cron');
    expect(repos).toHaveTextContent('master');
    expect(repos).toHaveTextContent('44.28 kB');
    expect(repos).toHaveTextContent('AUTO');
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

    await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
    await click((await screen.findAllByText('Inspect commit'))[0]);

    expect(window.location.pathname).toBe(
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4',
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

    await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
    await click(
      (
        await screen.findAllByText('Apply Global ID and view in DAG')
      )[0],
    );

    expect(window.location.search).toBe(
      '?globalIdFilter=0918ac9d5daa76b86e3bb5e88e4c43a4',
    );
  });

  it('should sort commits by start time', async () => {
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
    expect(commits[0]).toHaveTextContent('0918ac...');
    expect(commits[1]).toHaveTextContent('0918ac...');
    expect(commits[2]).toHaveTextContent('9d5daa...');

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByText('Created: Oldest'));

    commits = screen.getAllByTestId('CommitsList__row');
    expect(commits[0]).toHaveTextContent('0918ac...');
    expect(commits[1]).toHaveTextContent('0218ac...');
    expect(commits[2]).toHaveTextContent('0518ac...');
  });

  it('should filter commits by originKind', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('ReposTable__loadingDots'),
    );
    await click(screen.getAllByTestId('RepositoriesList__row')[1]);
    await click(screen.getByText('Commits'));
    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('CommitsList__loadingDots'),
    );

    expect(await screen.findAllByTestId('CommitsList__row')).toHaveLength(6);

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByText('User'));
    let commits = screen.getAllByTestId('CommitsList__row');
    expect(commits).toHaveLength(2);
    expect(commits[0]).toHaveTextContent('0918ac...');
    expect(commits[1]).toHaveTextContent('9d5daa...');

    await click(screen.getByText('Auto'));
    await click(screen.getByTestId('Filter__commitTypeUserChip'));
    commits = screen.getAllByTestId('CommitsList__row');
    expect(commits).toHaveLength(4);
  });

  it('should filter commits by ID', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('ReposTable__loadingDots'),
    );
    await click(screen.getAllByTestId('RepositoriesList__row')[1]);
    await click(screen.getByText('Commits'));
    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('CommitsList__loadingDots'),
    );

    expect(await screen.findAllByTestId('CommitsList__row')).toHaveLength(6);

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByText('Select commit IDs'));
    await click(screen.getByLabelText('9d5daa...'));

    const commits = await screen.findAllByTestId('CommitsList__row');
    expect(commits).toHaveLength(1);
    expect(commits[0]).toHaveTextContent('9d5daa...');
  });
});
