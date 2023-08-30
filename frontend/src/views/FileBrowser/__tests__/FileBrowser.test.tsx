import {mockGetFilesQuery} from '@graphqlTypes';
import {
  render,
  screen,
  waitForElementToBeRemoved,
  within,
} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockEmptyCommitDiff,
  mockEmptyGetAuthorize,
  mockGetCommitA4,
  mockGetCommitsA4Only,
  mockGetImageCommits,
  mockGetVersionInfo,
  mockRepoImages,
  mockRepoImagesWithLinkedPipeline,
} from '@dash-frontend/mocks';
import {mockGetBranches} from '@dash-frontend/mocks/branches';
import {
  generatePagingFiles,
  mockEmptyFiles,
  mockErrorFiles,
  mockImagesFiles,
} from '@dash-frontend/mocks/files';
import {click, withContextProviders} from '@dash-frontend/testHelpers';

import FileBrowserComponent from '../FileBrowser';
const FileBrowser = withContextProviders(() => {
  return <FileBrowserComponent />;
});

describe('File Browser', () => {
  const server = setupServer();

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    server.resetHandlers();
    window.history.replaceState(
      {},
      '',
      '/project/default/repos/images/branch/master/commit/4a83c74809664f899261baccdb47cd90',
    );
    server.use(mockGetBranches());
    server.use(mockRepoImages());
    server.use(mockRepoImagesWithLinkedPipeline());
    server.use(mockGetCommitA4());
    server.use(mockEmptyCommitDiff());
    server.use(mockImagesFiles());
    server.use(mockGetCommitsA4Only());
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
        '4a83c74809664f899261baccdb47cd90',
      );

      expect(await screen.findByTestId('LeftPanel_crumb')).toHaveTextContent(
        'Commit: 4a83c7...',
      );
    });

    it('should select the correct commit from the url', async () => {
      server.use(mockGetImageCommits());
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

  describe('Right Panel', () => {
    it('should display repo details at top level and in a folder', async () => {
      render(<FileBrowser />);
      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));
      expect(await screen.findByText('added mako')).toBeInTheDocument();
      expect(await screen.findByText('images')).toBeInTheDocument();
      expect(await screen.findByText('repo of images')).toBeInTheDocument();
      await click(screen.getByText('cats'));
      expect(await screen.findByText('added mako')).toBeInTheDocument();
      expect(await screen.findByText('images')).toBeInTheDocument();
      expect(await screen.findByText('repo of images')).toBeInTheDocument();
    });

    it('should display file history when previewing a file', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      await click(screen.getByText('AT-AT.png'));
      expect(await screen.findByText('File Versions')).toBeInTheDocument();
      expect(
        await screen.findByRole('button', {
          name: 'Load older file versions',
        }),
      ).toBeInTheDocument();
    });
  });

  describe('Middle Section', () => {
    it('should display commit id and branch, and copy path on click', async () => {
      render(<FileBrowser />);

      expect(
        screen.getByText('Commit: 4a83c74809664f899261baccdb47cd90'),
      ).toBeInTheDocument();
      expect(screen.getByText('Branch: master')).toBeInTheDocument();

      const copyAction = await screen.findByRole('button', {
        name: 'Copy commit id',
      });
      await click(copyAction);

      expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
        'images@4a83c74809664f899261baccdb47cd90',
      );
    });

    it('should navigate to file preview on action click', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      await click((await screen.findAllByText('Preview'))[0]);

      expect(window.location.pathname).toBe(
        '/project/default/repos/images/branch/master/commit/4a83c74809664f899261baccdb47cd90/AT-AT.png/',
      );
    });

    it('should navigate up a folder on back button click', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/branch/master/commit/4a83c74809664f899261baccdb47cd90/cats%2F/',
      );
      render(<FileBrowser />);
      await screen.findByText('Folder: cats');
      await click(screen.getByRole('button', {name: 'Back'}));

      expect(screen.getByText('Commit files for')).toBeInTheDocument();
      expect(window.location.pathname).toBe(
        '/project/default/repos/images/branch/master/commit/4a83c74809664f899261baccdb47cd90/',
      );
    });

    it('should display a message if the selected commit is open', async () => {
      server.use(mockGetImageCommits());

      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/branch/master/commit/4eb1aa567dab483f93a109db4641ee75/',
      );
      render(<FileBrowser />);
      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));
      expect(
        screen.getByText('This commit is currently open'),
      ).toBeInTheDocument();
    });

    it('should display a message if the selected commit empty', async () => {
      server.use(mockEmptyFiles());

      render(<FileBrowser />);
      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));
      expect(
        screen.getByText("This commit doesn't have any files"),
      ).toBeInTheDocument();
    });

    it('should display a message if the there is an error retrieving files', async () => {
      server.use(mockErrorFiles());

      render(<FileBrowser />);
      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));
      expect(
        screen.getByText("We couldn't load the file list"),
      ).toBeInTheDocument();
    });
  });

  describe('File Paging', () => {
    beforeEach(() => {
      const files = generatePagingFiles({
        n: 55,
        repoName: 'images',
        commitId: '4a83c74809664f899261baccdb47cd90',
      });
      server.use(
        mockGetFilesQuery((req, res, ctx) => {
          const {projectId, repoName, branchName, commitId, cursorPath, limit} =
            req.variables.args;
          if (
            projectId === 'default' &&
            repoName === 'images' &&
            branchName === 'master' &&
            commitId === '4a83c74809664f899261baccdb47cd90'
          ) {
            if (limit === 50 && !cursorPath) {
              return res(
                ctx.data({
                  files: {
                    files: files.slice(0, 50),
                    cursor: files[50].path,
                    hasNextPage: true,
                  },
                }),
              );
            }

            if (limit === 50 && cursorPath === files[50].path) {
              return res(
                ctx.data({
                  files: {
                    files: files.slice(50),
                    cursor: null,
                    hasNextPage: false,
                  },
                }),
              );
            }

            if (limit === 50 && !cursorPath) {
              return res(
                ctx.data({
                  files: {
                    files: files,
                    cursor: null,
                    hasNextPage: false,
                  },
                }),
              );
            }

            return res(
              ctx.data({
                files: {
                  files: files,
                  cursor: null,
                  hasNextPage: false,
                },
              }),
            );
          }
          return res();
        }),
      );
    });

    it('should allow users to navigate through paged files', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      let foundFiles = screen.getAllByTestId('FileTableRow__row');
      expect(foundFiles).toHaveLength(50);
      expect(foundFiles[0]).toHaveTextContent('0.png');
      expect(foundFiles[49]).toHaveTextContent('49.png');

      let pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();
      await click(within(pager).getByTestId('Pager__forward'));

      foundFiles = screen.getAllByTestId('FileTableRow__row');
      expect(foundFiles).toHaveLength(5);
      expect(foundFiles[0]).toHaveTextContent('50.png');
      expect(foundFiles[4]).toHaveTextContent('54.png');

      pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__forward')).toBeDisabled();
      await click(within(pager).getByTestId('Pager__backward'));

      foundFiles = screen.getAllByTestId('FileTableRow__row');
      expect(foundFiles).toHaveLength(50);
    });

    it('should allow users to update page size', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      let foundFiles = screen.getAllByTestId('FileTableRow__row');
      expect(foundFiles).toHaveLength(50);

      let pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__forward')).toBeEnabled();
      expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();

      await click(within(pager).getByTestId('DropdownButton__button'));
      await click(within(pager).getByText(100));

      pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__forward')).toBeDisabled();
      expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();

      foundFiles = screen.getAllByTestId('FileTableRow__row');
      expect(foundFiles).toHaveLength(55);
    });
  });
});
