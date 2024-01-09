import {
  render,
  waitForElementToBeRemoved,
  screen,
  within,
} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {CommitInfo, ListCommitRequest} from '@dash-frontend/api/pfs';
import {
  mockGetImageCommits,
  mockGetEnterpriseInfoInactive,
  generatePagingCommits,
} from '@dash-frontend/mocks';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import RepoListComponent from '../RepoList';

describe('Repo Commits List', () => {
  const server = setupServer();

  const RepoList = withContextProviders(() => {
    return <RepoListComponent />;
  });

  beforeAll(() => {
    server.listen();
    server.use(mockGetImageCommits());
    server.use(mockGetEnterpriseInfoInactive());
  });

  beforeEach(async () => {
    window.history.replaceState(
      '',
      '',
      '/project/default/repos/commits?selectedRepos=images',
    );
  });

  afterAll(() => server.close());

  it('should display commit details', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('CommitsList__loadingDots'),
    );

    const commit = screen.getAllByTestId('CommitsList__row')[2];
    expect(commit).toHaveTextContent('@images');
    expect(commit).toHaveTextContent('master');
    expect(commit).toHaveTextContent('Jul 24, 2023; 17:58');
    expect(commit).toHaveTextContent('58.65 kB');
    expect(commit).toHaveTextContent('c43fffd650a24b40b7d9f1bf90fcfdbe');
    expect(commit).toHaveTextContent('sold materia');
  });

  it('should allow users to inspect a commit', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('CommitsList__loadingDots'),
    );

    await click((await screen.findAllByTestId('DropdownButton__button'))[2]);
    await click((await screen.findAllByText('Inspect commit'))[2]);

    expect(window.location.pathname).toBe(
      '/project/default/repos/images/branch/master/commit/c43fffd650a24b40b7d9f1bf90fcfdbe/',
    );
  });

  it('should apply a globalId and redirect', async () => {
    render(<RepoList />);

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
      '?globalIdFilter=c43fffd650a24b40b7d9f1bf90fcfdbe',
    );
  });

  describe('commit paging', () => {
    beforeAll(() => {
      const commits = generatePagingCommits({n: 30});
      server.use(
        rest.post<ListCommitRequest, Empty, CommitInfo[]>(
          '/api/pfs_v2.API/ListCommit',
          async (req, res, ctx) => {
            const body = await req.json();
            const {number, startedTime} = body;

            if (Number(number) === 16 && !startedTime) {
              return res(ctx.json(commits.slice(0, 16)));
            }

            if (
              Number(number) === 16 &&
              startedTime === '1970-01-01T00:00:14Z'
            ) {
              return res(ctx.json(commits.slice(15, 31)));
            }

            if (Number(number) === 51 && !startedTime) {
              return res(ctx.json(commits));
            }

            return res(ctx.json(commits));
          },
        ),
      );
    });

    beforeEach(() => {
      window.history.replaceState(
        '',
        '',
        '/project/Load-Project/repos/commits?selectedRepos=repo',
      );
    });

    it('should allow users to navigate through paged files', async () => {
      render(<RepoList />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('CommitsList__loadingDots'),
      );

      let commits = screen.getAllByTestId('CommitsList__row');
      expect(commits).toHaveLength(15);
      expect(commits[0]).toHaveTextContent('item 0');
      expect(commits[14]).toHaveTextContent('item 14');

      let pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();
      await click(within(pager).getByTestId('Pager__forward'));

      expect(await screen.findByText('item 15')).toBeInTheDocument();
      commits = screen.getAllByTestId('CommitsList__row');
      expect(commits).toHaveLength(15);
      expect(commits[0]).toHaveTextContent('item 15');
      commits = screen.getAllByTestId('CommitsList__row');
      expect(commits[0]).toHaveTextContent('item 15');
      expect(commits[14]).toHaveTextContent('item 29');

      pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__forward')).toBeDisabled();
      await click(within(pager).getByTestId('Pager__backward'));

      commits = screen.getAllByTestId('CommitsList__row');
      expect(commits).toHaveLength(15);
      expect(commits[0]).toHaveTextContent('item 0');
      expect(commits[14]).toHaveTextContent('item 14');
    });

    it('should allow users to update page size', async () => {
      render(<RepoList />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('CommitsList__loadingDots'),
      );

      expect(await screen.findByText('item 0')).toBeInTheDocument();

      let commits = screen.getAllByTestId('CommitsList__row');
      expect(commits).toHaveLength(15);

      const pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__forward')).toBeEnabled();
      expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();

      await click(within(pager).getByTestId('DropdownButton__button'));
      await click(within(pager).getByText(50));

      expect(await screen.findByText('item 16')).toBeInTheDocument();
      commits = screen.getAllByTestId('CommitsList__row');
      expect(commits).toHaveLength(30);
    });
  });
});
