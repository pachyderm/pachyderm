import {fireEvent, render, screen, waitFor} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  FileInfo,
  FileType,
  ListFileRequest,
} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import {MOCK_IMAGES_FILES} from '@dash-frontend/mocks';
import {withContextProviders} from '@dash-frontend/testHelpers';

import {NAVIGATION_KEYS} from '../../constants/FileBrowser';
import useFilePreviewNavigation from '../useFilePreviewNavigation';

const FilePreviewNavigationComponent = withContextProviders(
  ({repoName = 'images'}: {repoName?: string}) => {
    const file = {
      file: {
        commit: {
          repo: {
            name: repoName,
            type: 'user',
            project: {
              name: 'default',
            },
          },
          id: '4a83c74809664f899261baccdb47cd90',
          branch: {
            repo: {
              name: repoName,
              type: 'user',
              project: {
                name: 'default',
              },
            },
            name: 'master',
          },
        },
        path: `/cool.png`,
      },
      fileType: FileType.FILE,
      committed: '2023-11-08T18:12:19.363338Z',
      sizeBytes: '80590',
    };
    const {loading} = useFilePreviewNavigation({
      file,
    });

    return loading ? <span>Loading...</span> : <span>Ready</span>;
  },
);

describe('File Preview Navigation', () => {
  const server = setupServer();

  beforeAll(() => server.listen());

  beforeEach(() => {
    window.history.replaceState(
      {},
      '',
      '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/cool.png',
    );
    server.resetHandlers();
    server.use(
      rest.post<ListFileRequest, Empty, FileInfo[]>(
        '/api/pfs_v2.API/ListFile',
        async (req, res, ctx) => {
          const body = await req.json();

          if (
            body.file.commit.repo.project.name === 'default' &&
            body.file.commit.repo.name === 'images' &&
            body.file.commit.id === '4a83c74809664f899261baccdb47cd90' &&
            body.file.path === '/' &&
            body.paginationMarker.commit.id ===
              '4a83c74809664f899261baccdb47cd90' &&
            body.number === '1'
          ) {
            if (body.reverse) {
              return res(ctx.json([MOCK_IMAGES_FILES[0]]));
            } else {
              return res(ctx.json([MOCK_IMAGES_FILES[1]]));
            }
          }
          return res(ctx.json([]));
        },
      ),
    );
  });

  afterAll(() => server.close());

  it('should navigate to the next file', async () => {
    render(<FilePreviewNavigationComponent />);

    await screen.findByText('Ready');
    fireEvent.keyDown(document, {key: NAVIGATION_KEYS.NEXT});
    await waitFor(() => {
      expect(window.location.pathname).toBe(
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/liberty.png',
      );
    });
  });

  it('should navigate to the previous file', async () => {
    render(<FilePreviewNavigationComponent />);

    await screen.findByText('Ready');
    fireEvent.keyDown(document, {key: NAVIGATION_KEYS.PREVIOUS});
    await waitFor(() => {
      expect(window.location.pathname).toBe(
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/AT-AT.png',
      );
    });
  });

  it('should navigate to the parent directory', async () => {
    render(<FilePreviewNavigationComponent />);

    await screen.findByText('Ready');
    fireEvent.keyDown(document, {key: NAVIGATION_KEYS.PARENT});
    await waitFor(() => {
      expect(window.location.pathname).toBe(
        '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90',
      );
    });
  });

  it('should not navigate when there is not a next or previous file', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/default/repos/other/commit/4a83c74809664f899261baccdb47cd90/cool.png',
    );
    render(<FilePreviewNavigationComponent repoName="other" />);

    await screen.findByText('Ready');
    fireEvent.keyDown(document, {key: NAVIGATION_KEYS.NEXT});
    await waitFor(() => {
      expect(window.location.pathname).toBe(
        '/project/default/repos/other/commit/4a83c74809664f899261baccdb47cd90/cool.png',
      );
    });

    await screen.findByText('Ready');
    fireEvent.keyDown(document, {key: NAVIGATION_KEYS.PREVIOUS});
    await waitFor(() => {
      expect(window.location.pathname).toBe(
        '/project/default/repos/other/commit/4a83c74809664f899261baccdb47cd90/cool.png',
      );
    });
  });
});
