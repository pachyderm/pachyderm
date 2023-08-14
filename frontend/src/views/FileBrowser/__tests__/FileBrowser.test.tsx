import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/lib/node';
import React from 'react';

import {
  mockEmptyCommitDiff,
  mockEmptyGetAuthorize,
  mockGetCommitA4,
  mockGetImageCommits,
  mockGetVersionInfo,
  mockRepoImages,
} from '@dash-frontend/mocks';
import {mockGetBranches} from '@dash-frontend/mocks/branches';
import {withContextProviders} from '@dash-frontend/testHelpers';

import FileBrowserComponent from '../FileBrowser';
const FileBrowser = withContextProviders(() => {
  return <FileBrowserComponent />;
});

describe('File Browser', () => {
  const server = setupServer();

  beforeAll(() => server.listen());

  beforeEach(() => {
    server.resetHandlers();
    server.use(mockGetImageCommits());
    server.use(mockGetBranches());
    server.use(mockRepoImages());
    server.use(mockGetCommitA4());
    server.use(mockEmptyCommitDiff());
    server.use(mockGetVersionInfo());
    server.use(mockEmptyGetAuthorize());
  });

  afterAll(() => server.close());

  describe('Left Panel', () => {
    it('should select the latest commit', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/branch/master/latest',
      );
      render(<FileBrowser />);
      const selectedCommit = (
        await screen.findAllByTestId('CommitList__listItem')
      )[0];

      expect(selectedCommit).toHaveClass('selected');
      expect(selectedCommit).toHaveTextContent(
        '4eb1aa567dab483f93a109db4641ee75',
      );

      expect(await screen.findByTestId('LeftPanel_crumb')).toHaveTextContent(
        'Commit: 4eb1aa...',
      );
    });
    it('should select the correct commit from the url', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/branch/master/commit/c43fffd650a24b40b7d9f1bf90fcfdbe',
      );
      render(<FileBrowser />);

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
  });
});
