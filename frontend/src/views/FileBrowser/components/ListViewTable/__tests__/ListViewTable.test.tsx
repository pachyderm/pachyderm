import {
  mockDeleteFilesMutation,
  mockRepoWithLinkedPipelineQuery,
} from '@graphqlTypes';
import {render, screen, waitFor, within} from '@testing-library/react';
import {setupServer} from 'msw/lib/node';
import React from 'react';

import {
  mockEmptyGetAuthorize,
  mockGetVersionInfo,
  mockRepoImagesWithLinkedPipeline,
} from '@dash-frontend/mocks';
import {MOCK_IMAGES_FILES} from '@dash-frontend/mocks/files';
import {click, withContextProviders} from '@dash-frontend/testHelpers';

import {default as ListViewTableComponent} from '../ListViewTable';

const ListViewTable = withContextProviders(({files}) => {
  return <ListViewTableComponent files={files} />;
});

describe('List View Table', () => {
  const server = setupServer();

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    window.history.replaceState(
      {},
      '',
      '/project/default/repos/images/branch/master/commit/4a83c74809664f899261baccdb47cd90',
    );
    server.use(mockEmptyGetAuthorize());
    server.use(mockGetVersionInfo());
    server.use(mockRepoImagesWithLinkedPipeline());
  });

  afterEach(() => server.resetHandlers());

  afterAll(() => server.close());

  it('should display file info per table row', async () => {
    render(<ListViewTable files={MOCK_IMAGES_FILES} />);

    const files = screen.getAllByTestId('FileTableRow__row');
    expect(files[0]).toHaveTextContent('AT-AT.png');
    expect(files[0]).toHaveTextContent('-');
    expect(files[0]).toHaveTextContent('80.59 kB');
    expect(files[1]).toHaveTextContent('liberty.png');
    expect(files[1]).toHaveTextContent('Added');
    expect(files[1]).toHaveTextContent('58.65 kB');
  });

  it('should navigate to preview on file link', async () => {
    render(<ListViewTable files={MOCK_IMAGES_FILES} />);

    await click(screen.getByText('AT-AT.png'));
    expect(window.location.pathname).toBe(
      '/project/default/repos/images/branch/master/commit/4a83c74809664f899261baccdb47cd90/AT-AT.png/',
    );
  });

  it('should navigate to dir on file link', async () => {
    render(<ListViewTable files={MOCK_IMAGES_FILES} />);

    await click(screen.getByText('cats'));
    expect(window.location.pathname).toBe(
      '/project/default/repos/images/branch/master/commit/4a83c74809664f899261baccdb47cd90/cats%2F/',
    );
  });

  it('should copy path on action click', async () => {
    render(<ListViewTable files={MOCK_IMAGES_FILES} />);

    await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
    await click((await screen.findAllByText('Copy Path'))[0]);

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      'images@master=4a83c74809664f899261baccdb47cd90:/AT-AT.png',
    );
  });

  it('should delete file on action click', async () => {
    server.use(
      mockDeleteFilesMutation((req, res, ctx) => {
        const {projectId, repo, branch, filePaths} = req.variables.args;
        if (
          projectId === 'default' &&
          repo === 'images' &&
          branch === 'master' &&
          JSON.stringify(filePaths) === JSON.stringify(['/AT-AT.png'])
        ) {
          return res(
            ctx.data({
              deleteFiles: '720d471659dc4682a53576fdb637a482',
            }),
          );
        }
        return res(ctx.errors(['file does not exist']));
      }),
    );

    render(<ListViewTable files={MOCK_IMAGES_FILES} />);

    await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
    await click((await screen.findAllByText('Delete'))[0]);

    const deleteConfirm = await screen.findByTestId('ModalFooter__confirm');
    await click(deleteConfirm);

    expect(window.location.pathname).toBe(
      '/project/default/repos/images/branch/master/commit/720d471659dc4682a53576fdb637a482/',
    );
  });

  it('should delete multiple files on action click', async () => {
    server.use(
      mockDeleteFilesMutation((req, res, ctx) => {
        const {projectId, repo, branch, filePaths} = req.variables.args;
        if (
          projectId === 'default' &&
          repo === 'images' &&
          branch === 'master' &&
          JSON.stringify(filePaths) ===
            JSON.stringify(['/AT-AT.png', '/cats/', '/liberty.png'])
        ) {
          return res(
            ctx.data({
              deleteFiles: '720d471659dc4682a53576fdb637a482',
            }),
          );
        }
        return res(ctx.errors(['file does not exist']));
      }),
    );
    render(<ListViewTable files={MOCK_IMAGES_FILES} />);

    const deleteButton = screen.getByRole('button', {
      name: /delete selected items/i,
    });
    expect(deleteButton).toBeDisabled();

    await click(
      screen.getByRole('cell', {
        name: /at-at\.png/i,
      }),
    );
    await click(
      screen.getByRole('cell', {
        name: /cats/i,
      }),
    );
    await click(
      screen.getByRole('cell', {
        name: /liberty\.png/i,
      }),
    );

    await waitFor(() => {
      expect(deleteButton).toBeEnabled();
    });

    await click(deleteButton);

    const modal = await screen.findByRole('dialog');
    expect(await within(modal).findByRole('list')).toHaveTextContent(
      '/AT-AT.png/cats//liberty.png',
    );
    const deleteConfirm = await within(modal).findByRole('button', {
      name: /delete/i,
    });
    await click(deleteConfirm);

    expect(window.location.pathname).toBe(
      '/project/default/repos/images/branch/master/commit/720d471659dc4682a53576fdb637a482/',
    );
  });

  it('should not allow file deletion for outputRepos', async () => {
    server.use(
      mockRepoWithLinkedPipelineQuery((req, res, ctx) => {
        const {projectId, id} = req.variables.args;
        if (projectId === 'default' && id === 'images') {
          return res(
            ctx.data({
              repo: {
                branches: [
                  {
                    name: 'master',
                    __typename: 'Branch',
                  },
                ],
                createdAt: 1690221504,
                description: '',
                id: 'images',
                name: 'images',
                sizeDisplay: '0 B',
                sizeBytes: 14783,
                access: true,
                projectId: 'default',
                linkedPipeline: {
                  id: 'default_pipeline',
                  name: 'pipeline',
                  __typename: 'Pipeline',
                },
                authInfo: null,
                __typename: 'Repo',
              },
            }),
          );
        }
        return res();
      }),
    );

    render(<ListViewTable files={MOCK_IMAGES_FILES} />);

    await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
    expect(screen.queryByText('Delete')).not.toBeInTheDocument();
  });
});
