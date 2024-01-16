import {render, screen} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  CommitInfo,
  InspectCommitRequest,
  ListCommitRequest,
} from '@dash-frontend/api/pfs';
import {RequestError} from '@dash-frontend/api/utils/error';
import {getStandardDateFromUnixSeconds} from '@dash-frontend/lib/dateTime';
import {
  COMMIT_INFO_4A,
  COMMIT_INFO_C4,
  generatePagingCommits,
  mockGetEnterpriseInfoInactive,
  mockGetImageCommits,
  mockGetBranches,
  mockGetBranchesMasterOnly,
  mockGetVersionInfo,
} from '@dash-frontend/mocks';
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
    server.use(mockGetVersionInfo());
    server.use(mockGetImageCommits());
    server.use(mockGetBranchesMasterOnly());
    server.use(mockGetEnterpriseInfoInactive());
  });

  afterAll(() => server.close());

  it('should format commit list item correctly', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90',
    );
    render(<LeftPanel />);
    expect(
      (await screen.findAllByTestId('CommitList__listItem'))[0],
    ).toHaveTextContent(
      `${getStandardDateFromUnixSeconds(
        1690221505,
      )}4a83c74809664f899261baccdb47cd90`,
    );
  });

  it('should load commits and select commit from prop', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/default/repos/images/commit/XYZ',
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
      '/project/default/repos/images/commit/4eb1aa567dab483f93a109db4641ee75',
    );

    server.use(
      rest.post<InspectCommitRequest, Empty, CommitInfo | RequestError>(
        '/api/pfs_v2.API/InspectCommit',
        async (req, res, ctx) => {
          const body = await req.json();
          if (body.commit.id === '4a83c74809664f899261baccdb47cd90') {
            return res(ctx.json(COMMIT_INFO_4A));
          }

          return res(
            ctx.status(404),
            ctx.json({
              code: 5,
              message: 'not found',
              details: [],
            }),
          );
        },
      ),
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

  it('should allow users to switch branches and default to all commits for the /latest path', async () => {
    server.use(mockGetBranches());
    server.use(
      rest.post<ListCommitRequest, Empty, CommitInfo[]>(
        '/api/pfs_v2.API/ListCommit',
        async (req, res, ctx) => {
          const body = await req.json();
          if (body.to?.branch?.name === 'test') {
            return res(ctx.json([COMMIT_INFO_4A]));
          }
          return res(ctx.json([COMMIT_INFO_C4]));
        },
      ),
    );

    window.history.replaceState({}, '', '/lineage/default/repos/images/latest');
    render(<LeftPanel />);

    const dropdown = await screen.findByTestId('DropdownButton__button');

    expect(dropdown).toHaveTextContent('Viewing all commits');
    let selectedCommit = (
      await screen.findAllByTestId('CommitList__listItem')
    )[0];
    expect(selectedCommit).toHaveTextContent(
      'c43fffd650a24b40b7d9f1bf90fcfdbe',
    );

    await click(dropdown);
    await click(await screen.findByText('test'));
    expect(dropdown).toHaveTextContent('test');

    expect(window.location.search).toBe('?branchId=test');
    selectedCommit = (await screen.findAllByTestId('CommitList__listItem'))[0];
    expect(selectedCommit).toHaveTextContent(
      '4a83c74809664f899261baccdb47cd90',
    );
  });

  it('should allow a user to page through the commit list', async () => {
    const commits = generatePagingCommits({n: 100});
    server.use(
      rest.post<ListCommitRequest, Empty, CommitInfo[]>(
        '/api/pfs_v2.API/ListCommit',
        async (req, res, ctx) => {
          const body = await req.json();
          const {number, to} = body;

          if (Number(number) === 51 && !to?.id) {
            return res(ctx.json(commits.slice(0, 51)));
          }

          if (Number(number) === 51 && to?.id === commits[50].commit?.id) {
            return res(ctx.json(commits.slice(50, 100)));
          }

          if (Number(number) === 51 && !to?.id) {
            return res(ctx.json(commits));
          }

          return res(ctx.json(commits));
        },
      ),
    );

    window.history.replaceState(
      {},
      '',
      '/lineage/default/repos/images/commit/4eb1aa567dab483f93a109db4641ee75',
    );
    render(<LeftPanel selectedCommitId={'4eb1aa567dab483f93a109db4641ee75'} />);
    const forwards = await screen.findByTestId('Pager__forward');
    const backwards = await screen.findByTestId('Pager__backward');
    expect(await screen.findByText('Commits 1 - 50')).toBeInTheDocument();
    expect(backwards).toBeDisabled();
    expect(forwards).toBeEnabled();
    let foundCommits = await screen.findAllByTestId('CommitList__listItem');
    expect(foundCommits[0]).toHaveTextContent(commits[0].commit?.id || 'FAIL');
    expect(foundCommits[49]).toHaveTextContent(
      commits[49].commit?.id || 'FAIL',
    );
    await click(forwards);
    expect(await screen.findByText('Commits 51 - 100')).toBeInTheDocument();
    foundCommits = await screen.findAllByTestId('CommitList__listItem');
    expect(foundCommits[0]).toHaveTextContent(commits[50].commit?.id || 'FAIL');
    expect(foundCommits[49]).toHaveTextContent(
      commits[99].commit?.id || 'FAIL',
    );
    expect(backwards).toBeEnabled();
    expect(forwards).toBeDisabled();
  });
});
