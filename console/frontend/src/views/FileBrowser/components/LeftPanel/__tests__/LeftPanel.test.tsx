import {render, screen, waitFor, within} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  FileInfo,
  FileType,
  GlobFileRequest,
  ListFileRequest,
} from '@dash-frontend/api/pfs';
import {
  mockGetEnterpriseInfoInactive,
  mockGetImageCommits,
  mockGetBranchesMasterOnly,
  mockGetVersionInfo,
  inspectCommit,
  mockImagesFiles,
  buildFile,
  generatePagingFiles,
  mockEmptyFiles,
} from '@dash-frontend/mocks';
import {
  click,
  keyboard,
  type,
  withContextProviders,
} from '@dash-frontend/testHelpers';

import {Empty} from '../../../../../../@types/google/protobuf/empty.pb';
import {default as LeftPanelComponent} from '../LeftPanel';

const LeftPanel = withContextProviders(
  ({selectedCommitId, isCommitOpen = false}) => {
    return (
      <LeftPanelComponent
        selectedCommitId={selectedCommitId}
        isCommitOpen={isCommitOpen}
      />
    );
  },
);

const buildFilePathAndType = (path: string, type: FileType) =>
  buildFile({
    file: {
      commit: {
        repo: {
          name: 'images',
          type: 'user',
          project: {
            name: 'default',
          },
        },
        id: '4a83c74809664f899261baccdb47cd90',
        branch: {
          repo: {
            name: 'images',
            type: 'user',
            project: {
              name: 'default',
            },
          },
          name: 'master',
        },
      },
      path,
      datum: 'default',
    },
    fileType: type,
    committed: '2023-11-08T18:12:19.363338Z',
    sizeBytes: '80590',
  });

/*  
File tree in mock
/
  /one
    one.txt
    one.png
  /two
    two.txt
    two.png
    /three
      three.txt
      three.png
  root.txt
  root.png
*/

const mockDeepFileTree = () =>
  rest.post<ListFileRequest, Empty, FileInfo[]>(
    '/api/pfs_v2.API/ListFile',
    async (req, res, ctx) => {
      const body = await req.json();
      if (
        body.file.commit.repo.project.name === 'default' &&
        body.file.commit.repo.name === 'images' &&
        body.file.commit.id === '4a83c74809664f899261baccdb47cd90'
      ) {
        if (body.file.path === '/' || body.file.path === '') {
          return res(
            ctx.json([
              buildFilePathAndType('/one/', FileType.DIR),
              buildFilePathAndType('/two/', FileType.DIR),
              buildFilePathAndType('root.txt', FileType.FILE),
              buildFilePathAndType('root.png', FileType.FILE),
            ]),
          );
        } else if (body.file.path === 'one/' || body.file.path === '/one/') {
          return res(
            ctx.json([
              buildFilePathAndType('/one/one.txt', FileType.FILE),
              buildFilePathAndType('/one/one.png', FileType.FILE),
            ]),
          );
        } else if (body.file.path === 'two/' || body.file.path === '/two/') {
          return res(
            ctx.json([
              buildFilePathAndType('/two/two.txt', FileType.FILE),
              buildFilePathAndType('/two/two.png', FileType.FILE),
              buildFilePathAndType('/two/three/', FileType.DIR),
            ]),
          );
        } else if (
          body.file.path === 'two/three/' ||
          body.file.path === '/two/three/'
        ) {
          return res(
            ctx.json([
              buildFilePathAndType('/two/three/three.txt', FileType.FILE),
              buildFilePathAndType('/two/three/three.png', FileType.FILE),
            ]),
          );
        }
      }
      return res(ctx.json([]));
    },
  );

// Jsdom is being annoying here. We need to click on the link not
// the list item that surrounds it.
const clickListItem = async (item: HTMLElement) => {
  await click(await within(item).findByRole('link'));
};

describe('Left Panel', () => {
  const server = setupServer();

  beforeAll(() => server.listen());

  beforeEach(() => {
    server.resetHandlers();
    server.use(mockGetVersionInfo());
    server.use(mockGetImageCommits());
    server.use(mockGetBranchesMasterOnly());
    server.use(mockGetEnterpriseInfoInactive());
    server.use(inspectCommit());
    server.use(mockImagesFiles());
  });

  afterAll(() => server.close());

  describe('Commit Select', () => {
    it('should dislay commits in dropdown correctly', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90',
      );
      render(
        <LeftPanel selectedCommitId={'4a83c74809664f899261baccdb47cd90'} />,
      );

      await waitFor(async () => {
        const commitsButton = await screen.findByTestId('CommitSelect__button');
        expect(commitsButton).toHaveTextContent(/jul 24, 2023; 17:58 4a83c7/i);
        await click(commitsButton);
      });

      const menuItems = await screen.findAllByRole('menuitem');
      expect(menuItems).toHaveLength(3);
      expect(menuItems[0]).toHaveTextContent(/jul 24, 2023; 17:58 4a83c7/i);
      expect(menuItems[1]).toHaveTextContent(/jul 24, 2023; 17:58 4eb1aa/i);
      expect(menuItems[2]).toHaveTextContent(/jul 24, 2023; 17:58 c43fff/i);
    });

    it('should display a link to the commits table at the bottom of the commits list dropdown', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90',
      );
      render(
        <LeftPanel selectedCommitId={'4a83c74809664f899261baccdb47cd90'} />,
      );

      await click(await screen.findByTestId('CommitSelect__button'));

      const link = screen.getByRole('link', {
        name: /view all commits/i,
      });
      expect(link).toHaveAttribute(
        'href',
        '/project/default/repos/commits?selectedRepos=images',
      );
    });

    it('should allow users to search for a commit in the dropdown', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/commit/4eb1aa567dab483f93a109db4641ee75',
      );

      render(
        <LeftPanel selectedCommitId={'4eb1aa567dab483f93a109db4641ee75'} />,
      );

      await waitFor(async () => {
        const commitsButton = await screen.findByTestId('CommitSelect__button');
        expect(commitsButton).toHaveTextContent(/jul 24, 2023; 17:58 4eb1aa/i);
        await click(commitsButton);
      });

      const menu = screen.getByRole('menu');
      const search = within(menu).getByRole('textbox');

      expect(screen.queryByText('No matching commits')).not.toBeInTheDocument();

      await type(search, '9d5daa0918ac4c22a476b86e3bb5e88e');
      await screen.findByText('No matching commits');

      await click(
        screen.getByRole('button', {
          name: /clear/i,
        }),
      );
      expect(search).toHaveTextContent('');
      await type(search, '4a83c74809664f899261baccdb47cd90');

      screen.getByRole('menuitem', {
        name: /jul 24, 2023; 17:58 4a83c7/i,
      });

      expect(screen.queryByText('No matching commits')).not.toBeInTheDocument();
    });
  });

  describe('File Tree', () => {
    beforeEach(() => {
      server.use(mockDeepFileTree());
    });

    it('should display file tree from root when there is no path in url', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/',
      );
      render(
        <LeftPanel selectedCommitId={'4a83c74809664f899261baccdb47cd90'} />,
      );

      const fileTree = await screen.findByRole('region', {
        name: /file tree/i,
      });
      await within(fileTree).findByRole('listitem', {
        name: /file root\.txt/i,
      });
      await within(fileTree).findByRole('listitem', {
        name: /file root\.png/i,
      });

      const folderOne = await within(fileTree).findByRole('listitem', {
        name: /directory one\/$/i,
      });
      expect(
        within(folderOne).queryByRole('listitem', {
          name: /file one\.png/i,
        }),
      ).not.toBeInTheDocument();

      const folderTwo = await within(fileTree).findByRole('listitem', {
        name: /directory two\/$/i,
      });
      expect(
        within(folderTwo).queryByRole('listitem', {
          name: /file two\.png/i,
        }),
      ).not.toBeInTheDocument();
    });

    it('should display file tree from path and select file', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/two%2Fthree%2F/',
      );
      render(
        <LeftPanel selectedCommitId={'4a83c74809664f899261baccdb47cd90'} />,
      );

      const fileTree = await screen.findByRole('region', {
        name: /file tree/i,
      });
      await within(fileTree).findByRole('listitem', {
        name: /file root\.txt/i,
      });
      await within(fileTree).findByRole('listitem', {
        name: /file root\.png/i,
      });

      const folderOne = await within(fileTree).findByRole('listitem', {
        name: /directory one\/$/i,
      });
      expect(
        within(folderOne).queryByRole('listitem', {
          name: /file one\.png/i,
        }),
      ).not.toBeInTheDocument();

      const folderTwo = await within(fileTree).findByRole('listitem', {
        name: /directory two\/$/i,
      });
      await within(folderTwo).findByRole('listitem', {
        name: /file two\.txt/i,
      });
      await within(folderTwo).findByRole('listitem', {
        name: /file two\.png/i,
      });
      const folderThree = await within(folderTwo).findByRole('listitem', {
        name: /directory two\/three\/$/i,
      });
      await within(folderThree).findByRole('listitem', {
        name: /file three\.txt/i,
      });
      await within(folderThree).findByRole('listitem', {
        name: /file three\.png/i,
      });
    });

    it('should set path on click', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/',
      );
      render(
        <LeftPanel selectedCommitId={'4a83c74809664f899261baccdb47cd90'} />,
      );

      const fileTree = await screen.findByRole('region', {
        name: /file tree/i,
      });

      const folderOne = await within(fileTree).findByRole('listitem', {
        name: /directory one\/$/i,
      });
      await clickListItem(folderOne);

      expect(window.location.pathname).toBe(
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/one%2F/',
      );

      await clickListItem(
        await within(folderOne).findByRole('listitem', {
          name: /file one\.png/i,
        }),
      );
      expect(window.location.pathname).toBe(
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/one%2Fone.png/',
      );
    });

    it('should show the max file message if folder has too many files', async () => {
      const files = generatePagingFiles({
        n: 1001,
        repoName: 'images',
        commitId: '4a83c74809664f899261baccdb47cd90',
      });
      server.use(
        rest.post<ListFileRequest, Empty, FileInfo[]>(
          '/api/pfs_v2.API/ListFile',
          async (req, res, ctx) => {
            const body = await req.json();
            if (
              body.file.commit.repo.project.name === 'default' &&
              body.file.commit.repo.name === 'images' &&
              body.file.commit.id === '4a83c74809664f899261baccdb47cd90'
            ) {
              return res(ctx.json(files));
            }
            return res(ctx.json([]));
          },
        ),
      );

      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/',
      );

      render(
        <LeftPanel selectedCommitId={'4a83c74809664f899261baccdb47cd90'} />,
      );

      const fileTree = await screen.findByRole('region', {
        name: /file tree/i,
      });
      await within(fileTree).findByRole('listitem', {
        name: /file 999\.png/i,
      });
      await screen.findByText('Limited to 1000 entries');
    });

    it('should search for files and clear filter', async () => {
      server.use(
        rest.post<GlobFileRequest, Empty, FileInfo[]>(
          '/api/pfs_v2.API/GlobFile',
          async (req, res, ctx) => {
            const body = await req.json();
            if (
              body.commit.repo.project.name === 'default' &&
              body.commit.repo.name === 'images' &&
              body.commit.id === '4a83c74809664f899261baccdb47cd90'
            ) {
              if (body.pattern === '*one*') {
                return res(
                  ctx.json([buildFilePathAndType('/one/', FileType.DIR)]),
                );
              } else if (body.pattern === '**/*one*') {
                return res(
                  ctx.json([
                    buildFilePathAndType('/one/one.txt', FileType.FILE),
                  ]),
                );
              }
            }
            return res(ctx.json([]));
          },
        ),
      );

      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/',
      );
      render(
        <LeftPanel selectedCommitId={'4a83c74809664f899261baccdb47cd90'} />,
      );
      const searchBox = await screen.findByRole('textbox');
      await type(searchBox, 'two');
      await screen.findByText(/no files found/i);
      await click(
        screen.getByRole('button', {
          name: /clear/i,
        }),
      );

      await type(searchBox, 'one');
      expect(await screen.findByRole('link', {name: '/one/'})).toHaveAttribute(
        'href',
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/one%2F/',
      );
      expect(
        await screen.findByRole('link', {name: '/one/one.txt'}),
      ).toHaveAttribute(
        'href',
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/one%2Fone.txt/',
      );
    });

    it('should show correct file tree for empty commit', async () => {
      server.use(mockEmptyFiles());
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/',
      );
      render(
        <LeftPanel selectedCommitId={'4a83c74809664f899261baccdb47cd90'} />,
      );

      const fileTree = await screen.findByRole('region', {
        name: /file tree/i,
      });
      await within(fileTree).findByText('No files found');
    });

    it('should not load files and disable search if commit is open', async () => {
      render(
        <LeftPanel
          selectedCommitId={'4a83c74809664f899261baccdb47cd90'}
          isCommitOpen={true}
        />,
      );

      await screen.findByText('This commit is currently open');
      expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    });

    it('should allow users to select files with up and down arrows', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/',
      );
      render(
        <LeftPanel selectedCommitId={'4a83c74809664f899261baccdb47cd90'} />,
      );

      const fileTree = await screen.findByRole('region', {
        name: /file tree/i,
      });
      await within(fileTree).findByRole('listitem', {
        name: /file root\.txt/i,
      });
      await within(fileTree).findByRole('listitem', {
        name: /file root\.png/i,
      });

      await keyboard('[ArrowDown]');
      expect(window.location.pathname).toBe(
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/one%2F/',
      );

      await keyboard('[ArrowDown]');
      expect(window.location.pathname).toBe(
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/two%2F/',
      );

      await keyboard('[ArrowDown]');
      expect(window.location.pathname).toBe(
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/root.txt/',
      );

      await keyboard('[ArrowUp]');
      expect(window.location.pathname).toBe(
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/two%2F/',
      );

      await keyboard('[ArrowUp]');
      expect(window.location.pathname).toBe(
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/one%2F/',
      );

      const folderOne = await within(fileTree).findByRole('listitem', {
        name: /directory one\/$/i,
      });
      expect(
        within(folderOne).queryByRole('listitem', {
          name: /file one\.png/i,
        }),
      ).not.toBeInTheDocument();

      const folderTwo = await within(fileTree).findByRole('listitem', {
        name: /directory two\/$/i,
      });
      expect(
        within(folderTwo).queryByRole('listitem', {
          name: /file two\.png/i,
        }),
      ).not.toBeInTheDocument();
    });

    it('should allow users to open a folder with the right arrow key', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/',
      );
      render(
        <LeftPanel selectedCommitId={'4a83c74809664f899261baccdb47cd90'} />,
      );

      const fileTree = await screen.findByRole('region', {
        name: /file tree/i,
      });
      await within(fileTree).findByRole('listitem', {
        name: /file root\.txt/i,
      });
      await within(fileTree).findByRole('listitem', {
        name: /file root\.png/i,
      });

      const folderOne = await within(fileTree).findByRole('listitem', {
        name: /directory one\/$/i,
      });
      expect(
        within(folderOne).queryByRole('listitem', {
          name: /file one\.png/i,
        }),
      ).not.toBeInTheDocument();

      const folderTwo = await within(fileTree).findByRole('listitem', {
        name: /directory two\/$/i,
      });
      expect(
        within(folderTwo).queryByRole('listitem', {
          name: /file two\.png/i,
        }),
      ).not.toBeInTheDocument();

      await keyboard('[ArrowDown]');
      await keyboard('[ArrowRight]');
      expect(window.location.pathname).toBe(
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/one%2F/',
      );

      await within(fileTree).findByRole('listitem', {
        name: /file root\.txt/i,
      });
      await within(fileTree).findByRole('listitem', {
        name: /file root\.png/i,
      });

      within(folderOne).getByRole('listitem', {
        name: /file one\.txt/i,
      });
      within(folderOne).getByRole('listitem', {
        name: /file one\.png/i,
      });

      expect(
        within(folderTwo).queryByRole('listitem', {
          name: /file two\.png/i,
        }),
      ).not.toBeInTheDocument();
    });

    it('should allow users to close a folder by pressing left arrow twice', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/two%2Fthree%2Fthree.png/',
      );
      render(
        <LeftPanel selectedCommitId={'4a83c74809664f899261baccdb47cd90'} />,
      );

      const fileTree = await screen.findByRole('region', {
        name: /file tree/i,
      });
      await within(fileTree).findByRole('listitem', {
        name: /file root\.txt/i,
      });
      await within(fileTree).findByRole('listitem', {
        name: /file root\.png/i,
      });
      const folderOne = await within(fileTree).findByRole('listitem', {
        name: /directory one\/$/i,
      });
      expect(
        within(folderOne).queryByRole('listitem', {
          name: /file one\.png/i,
        }),
      ).not.toBeInTheDocument();

      const folderTwo = await within(fileTree).findByRole('listitem', {
        name: /directory two\/$/i,
      });
      await within(folderTwo).findByRole('listitem', {
        name: /file two\.txt/i,
      });
      await within(folderTwo).findByRole('listitem', {
        name: /file two\.png/i,
      });
      const folderThree = await within(folderTwo).findByRole('listitem', {
        name: /directory two\/three\/$/i,
      });
      await within(folderThree).findByRole('listitem', {
        name: /file three\.txt/i,
      });
      await within(folderThree).findByRole('listitem', {
        name: /file three\.png/i,
      });

      await keyboard('[ArrowLeft]');
      expect(window.location.pathname).toBe(
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/two%2Fthree%2F/',
      );

      await keyboard('[ArrowLeft]');
      expect(window.location.pathname).toBe(
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/two%2F/',
      );

      await within(fileTree).findByRole('listitem', {
        name: /file root\.txt/i,
      });
      await within(fileTree).findByRole('listitem', {
        name: /file root\.png/i,
      });

      expect(
        within(folderOne).queryByRole('listitem', {
          name: /file one\.png/i,
        }),
      ).not.toBeInTheDocument();

      await within(folderTwo).findByRole('listitem', {
        name: /file two\.txt/i,
      });
      await within(folderTwo).findByRole('listitem', {
        name: /file two\.png/i,
      });
      expect(
        within(folderThree).queryByRole('listitem', {
          name: /file three\.txt/i,
        }),
      ).not.toBeInTheDocument();
    });
  });
});
