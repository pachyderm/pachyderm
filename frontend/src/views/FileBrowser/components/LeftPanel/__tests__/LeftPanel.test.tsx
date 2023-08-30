import {mockCommitSearchQuery, mockGetCommitsQuery} from '@graphqlTypes';
import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {
  COMMIT_4A,
  COMMIT_C4,
  generatePagingCommits,
  mockEmptyGetAuthorize,
  mockGetImageCommits,
  mockGetVersionInfo,
} from '@dash-frontend/mocks';
import {
  mockGetBranches,
  mockGetBranchesMasterOnly,
} from '@dash-frontend/mocks/branches';
import {click, type, withContextProviders} from '@dash-frontend/testHelpers';

import {default as LeftPanelComponent} from '../LeftPanel';

const LeftPanel = withContextProviders(({selectedCommitId}) => {
  return <LeftPanelComponent selectedCommitId={selectedCommitId} />;
});

describe('Left Panel', () => {
  const server = setupServer();

  beforeAll(() => server.listen());

  beforeEach(() => {
    server.resetHandlers();
    server.use(mockEmptyGetAuthorize());
    server.use(mockGetVersionInfo());
    server.use(mockGetImageCommits());
    server.use(mockGetBranchesMasterOnly());
  });

  afterAll(() => server.close());

  it('should format commit list item correctly', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/default/repos/images/branch/master/commit/4eb1aa567dab483f93a109db4641ee75',
    );
    render(<LeftPanel />);
    expect(
      (await screen.findAllByTestId('CommitList__listItem'))[0],
    ).toHaveTextContent(
      `${getStandardDate(1690221505)}4eb1aa567dab483f93a109db4641ee75`,
    );
  });

  it('should load commits and select commit from prop', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/default/repos/images/branch/master/commit/XYZ',
    );
    render(<LeftPanel selectedCommitId={'c43fffd650a24b40b7d9f1bf90fcfdbe'} />);
    const selectedCommit = (
      await screen.findAllByTestId('CommitList__listItem')
    )[2];

    expect(selectedCommit).toHaveClass('selected');
    expect(selectedCommit).toHaveTextContent(
      'c43fffd650a24b40b7d9f1bf90fcfdbe',
    );

    expect(await screen.findByTestId('LeftPanel_crumb')).toHaveTextContent(
      'Commit: c43fff...',
    );
  });

  it('should allow users to search for a commit', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/default/repos/images/branch/master/commit/4eb1aa567dab483f93a109db4641ee75',
    );

    server.use(
      mockCommitSearchQuery((req, res, ctx) => {
        if (req.variables.args.id === '4a83c74809664f899261baccdb47cd90') {
          return res(ctx.data({commitSearch: COMMIT_4A}));
        }
        return res(ctx.data({commitSearch: null}));
      }),
    );

    render(<LeftPanel />);

    const search = await screen.findByTestId('CommitList__search');

    expect(
      screen.queryByText('No matching commits found'),
    ).not.toBeInTheDocument();

    await type(search, '9d5daa0918ac4c22a476b86e3bb5e88e');
    await screen.findByText('No matching commits found');
    await click(await screen.findByTestId('CommitList__searchClear'));
    expect(search).toHaveTextContent('');
    await type(search, 'werweriuowiejrklwkejrwiepriojw');
    expect(screen.getByText('Enter exact commit ID')).toBeInTheDocument();
    await click(await screen.findByTestId('CommitList__searchClear'));
    expect(search).toHaveTextContent('');
    await type(search, '4a83c74809664f899261baccdb47cd90');

    const selectedCommit = await screen.findByTestId('CommitList__listItem');
    expect(selectedCommit).toHaveTextContent(
      '4a83c74809664f899261baccdb47cd90',
    );
    expect(
      screen.queryByText('No matching commits found'),
    ).not.toBeInTheDocument();
  });

  it('should allow users to switch branches', async () => {
    server.use(mockGetBranches());
    server.use(
      mockGetCommitsQuery((req, res, ctx) => {
        if (req.variables.args.branchName === 'master') {
          return res(
            ctx.data({
              commits: {items: [COMMIT_4A], cursor: null, parentCommit: null},
            }),
          );
        }
        return res(
          ctx.data({
            commits: {items: [COMMIT_C4], cursor: null, parentCommit: null},
          }),
        );
      }),
    );

    window.history.replaceState(
      {},
      '',
      '/lineage/default/repos/images/branch/master/latest',
    );
    render(<LeftPanel />);

    const dropdown = await screen.findByTestId('DropdownButton__button');

    expect(dropdown).toHaveTextContent('master');
    let selectedCommit = (
      await screen.findAllByTestId('CommitList__listItem')
    )[0];
    expect(selectedCommit).toHaveTextContent(
      '4a83c74809664f899261baccdb47cd90',
    );

    await click(dropdown);
    await click(await screen.findByText('test'));
    expect(dropdown).toHaveTextContent('test');

    selectedCommit = (await screen.findAllByTestId('CommitList__listItem'))[0];
    expect(selectedCommit).toHaveTextContent(
      'c43fffd650a24b40b7d9f1bf90fcfdbe',
    );
  });

  it('should allow a user to page through the commit list', async () => {
    const commits = generatePagingCommits({n: 100});
    server.use(
      mockGetCommitsQuery((req, res, ctx) => {
        const {commitIdCursor, number} = req.variables.args;
        if (number === 50 && !commitIdCursor) {
          return res(
            ctx.data({
              commits: {
                items: commits.slice(0, 50),
                cursor: null,
                parentCommit: commits[50].id,
              },
            }),
          );
        }

        if (number === 50 && commitIdCursor === commits[50].id) {
          return res(
            ctx.data({
              commits: {
                items: commits.slice(50, 100),
                cursor: null,
                parentCommit: null,
              },
            }),
          );
        }

        if (number === 50 && !commitIdCursor) {
          return res(
            ctx.data({
              commits: {
                items: commits,
                cursor: null,
                parentCommit: null,
              },
            }),
          );
        }

        return res(
          ctx.data({
            commits: {
              items: commits,
              cursor: null,
              parentCommit: null,
            },
          }),
        );
      }),
    );

    window.history.replaceState(
      {},
      '',
      '/lineage/default/repos/images/branch/master/commit/4eb1aa567dab483f93a109db4641ee75',
    );
    render(<LeftPanel />);
    const forwards = await screen.findByTestId('Pager__forward');
    const backwards = await screen.findByTestId('Pager__backward');
    expect(await screen.findByText('Commits 1 - 50')).toBeInTheDocument();
    expect(backwards).toBeDisabled();
    expect(forwards).toBeEnabled();
    let foundCommits = await screen.findAllByTestId('CommitList__listItem');
    expect(foundCommits[0]).toHaveTextContent(commits[0].id);
    expect(foundCommits[49]).toHaveTextContent(commits[49].id);
    await click(forwards);
    expect(await screen.findByText('Commits 51 - 100')).toBeInTheDocument();
    foundCommits = await screen.findAllByTestId('CommitList__listItem');
    expect(foundCommits[0]).toHaveTextContent(commits[50].id);
    expect(foundCommits[49]).toHaveTextContent(commits[99].id);
    expect(backwards).toBeEnabled();
    expect(forwards).toBeDisabled();
  });
});
