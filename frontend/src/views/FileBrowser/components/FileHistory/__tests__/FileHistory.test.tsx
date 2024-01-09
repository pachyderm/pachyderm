import {
  render,
  screen,
  waitForElementToBeRemoved,
} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  FindCommitsRequest,
  FindCommitsResponse,
  InspectCommitRequest,
  CommitInfo,
} from '@dash-frontend/api/pfs';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import {
  mockGetEnterpriseInfoInactive,
  mockDiffFile,
} from '@dash-frontend/mocks';
import {
  COMMIT_INFO_4A,
  COMMIT_INFO_4E,
  COMMIT_INFO_C4,
} from '@dash-frontend/mocks/commits';
import {click, withContextProviders} from '@dash-frontend/testHelpers';

import FileHistoryComponent from '../FileHistory';

const FileHistory = withContextProviders(() => {
  return <FileHistoryComponent />;
});

describe('File History', () => {
  const server = setupServer();

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    server.resetHandlers();
    window.history.replaceState(
      {},
      '',
      '/project/default/repos/images/branch/master/commit/4a83c74809664f899261baccdb47cd90/AT-AT.png',
    );
    server.use(mockGetEnterpriseInfoInactive());
    server.use(
      rest.post<FindCommitsRequest, Empty, FindCommitsResponse[]>(
        '/api/pfs_v2.API/FindCommits',
        async (req, res, ctx) => {
          const body = await req.json();
          if (body.start.id === '4a83c74809664f899261baccdb47cd90') {
            return res(
              ctx.json([
                {
                  commitsSearched: 0,
                  foundCommit: COMMIT_INFO_4A.commit,
                  lastSearchedCommit: undefined,
                },
                {
                  commitsSearched: 1,
                  foundCommit: COMMIT_INFO_4E.commit,
                  lastSearchedCommit: undefined,
                },
                {
                  commitsSearched: 2,
                  foundCommit: undefined,
                  lastSearchedCommit: COMMIT_INFO_C4.commit,
                },
              ]),
            );
          }
          if (body.start.id === 'c43fffd650a24b40b7d9f1bf90fcfdbe') {
            return res(
              ctx.json([
                {
                  commitsSearched: 0,
                  foundCommit: COMMIT_INFO_C4.commit,
                  lastSearchedCommit: undefined,
                },
              ]),
            );
          }
        },
      ),
    );
    server.use(
      rest.post<InspectCommitRequest, Empty, CommitInfo>(
        '/api/pfs_v2.API/InspectCommit',
        async (req, res, ctx) => {
          const body = await req.json();
          if (body.commit.id === COMMIT_INFO_4A.commit?.id) {
            return res(ctx.json(COMMIT_INFO_4A));
          }
          if (body.commit.id === COMMIT_INFO_4E.commit?.id) {
            return res(ctx.json(COMMIT_INFO_4E));
          }
          if (body.commit.id === COMMIT_INFO_C4.commit?.id) {
            return res(ctx.json(COMMIT_INFO_C4));
          }
        },
      ),
    );
    server.use(mockDiffFile());
  });

  afterAll(() => server.close());

  it('should display file history when previewing a file', async () => {
    render(<FileHistory />);

    expect(await screen.findByText('File Versions')).toBeInTheDocument();
    await click(
      await screen.findByRole('button', {
        name: 'Load older file versions',
      }),
    );

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(
      screen.getByText(
        `${getStandardDateFromISOString(
          COMMIT_INFO_4A.started,
        )} - ${getStandardDateFromISOString(COMMIT_INFO_C4.started)}`,
      ),
    ).toBeInTheDocument();

    let commits = screen.getByTestId('FileHistory__commitList').children;

    expect(commits).toHaveLength(2);
    expect(commits[0]).toHaveTextContent(
      getStandardDateFromISOString(COMMIT_INFO_4A.started),
    );
    expect(commits[0]).toHaveTextContent('4a83c74809664f899261baccdb47cd90');
    expect(commits[0]).toHaveTextContent('added mako');
    expect(commits[1]).toHaveTextContent('Added');
    expect(commits[1]).toHaveTextContent(
      getStandardDateFromISOString(COMMIT_INFO_4E.started),
    );
    expect(commits[1]).toHaveTextContent('4eb1aa567dab483f93a109db4641ee75');
    expect(commits[1]).toHaveTextContent('commit not finished');
    expect(commits[1]).toHaveTextContent('Added');

    await click(
      screen.getByRole('button', {
        name: 'Load older file versions',
      }),
    );

    commits = screen.getByTestId('FileHistory__commitList').children;

    expect(commits).toHaveLength(3);
    expect(
      screen.getByRole('button', {
        name: 'Load older file versions',
      }),
    ).toBeDisabled();
  });
});
