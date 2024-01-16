import {
  render,
  screen,
  waitForElementToBeRemoved,
  within,
} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  DiffFileRequest,
  DiffFileResponse,
  FileType,
} from '@dash-frontend/api/pfs';
import {
  mockEmptyCommitDiff,
  mockEmptyInspectPipeline,
  mockFinishCommit,
  mockGetMontagePipeline,
  mockGetVersionInfo,
  mockStartCommit,
} from '@dash-frontend/mocks';
import {
  MOCK_IMAGES_FILES,
  buildFile,
  mockEncode,
} from '@dash-frontend/mocks/files';
import {click, hover, withContextProviders} from '@dash-frontend/testHelpers';

import {default as ListViewTableComponent} from '../ListViewTable';

const ListViewTable = withContextProviders(({files}) => {
  return <ListViewTableComponent files={files} />;
});

window.open = jest.fn();

describe('List View Table', () => {
  const server = setupServer();

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    window.history.replaceState(
      {},
      '',
      '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90',
    );
    server.use(mockGetVersionInfo());
    server.use(mockEmptyCommitDiff());
    server.use(mockEmptyInspectPipeline());
  });

  afterEach(() => server.resetHandlers());

  afterAll(() => server.close());

  it('should display file info per table row', async () => {
    server.use(
      rest.post<DiffFileRequest, Empty, DiffFileResponse[]>(
        '/api/pfs_v2.API/DiffFile',
        async (req, res, ctx) => {
          const body = await req.json();

          if (body.newFile.path === '/liberty.png') {
            return res(
              ctx.json([
                {
                  newFile: {
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
                      path: '/liberty.png',
                      datum: '',
                    },
                    fileType: FileType.FILE,
                  },
                },
              ]),
            );
          }
          return res(ctx.json([]));
        },
      ),
    );

    render(<ListViewTable files={MOCK_IMAGES_FILES} />);
    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    const files = await screen.findAllByTestId('FileTableRow__row');

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
      '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/AT-AT.png/',
    );
  });

  it('should navigate to dir on file link', async () => {
    render(<ListViewTable files={MOCK_IMAGES_FILES} />);

    await click(screen.getByText('cats'));
    expect(window.location.pathname).toBe(
      '/project/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/cats%2F/',
    );
  });

  it('should copy path on action click', async () => {
    render(<ListViewTable files={MOCK_IMAGES_FILES} />);

    await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
    await click((await screen.findAllByText('Copy Path'))[0]);

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      'images@4a83c74809664f899261baccdb47cd90:/AT-AT.png',
    );
  });

  describe('Delete', () => {
    it('should disable the delete button until selections are made', async () => {
      render(<ListViewTable files={MOCK_IMAGES_FILES} />);

      const deleteButton = screen.getByRole('button', {
        name: /delete selected items/i,
      });
      expect(deleteButton).toBeDisabled();
      await hover(deleteButton);
      expect(
        screen.getByText('Select one or more files to multi-delete files'),
      ).toBeInTheDocument();

      await click(
        screen.getByRole('cell', {
          name: /at-at\.png/i,
        }),
      );

      expect(deleteButton).toBeEnabled();
      await hover(deleteButton);
      expect(
        screen.queryByText('Select one or more files to multi-delete files'),
      ).not.toBeInTheDocument();
    });

    it('should delete file on action click', async () => {
      server.use(mockStartCommit('720d471659dc4682a53576fdb637a482'));
      server.use(
        rest.post<string, Empty>(
          '/api/pfs_v2.API/ModifyFile',
          async (req, res, ctx) => {
            const body = await req.text();
            const expected =
              '{"setCommit":{"repo":{"name":"images","type":"user","project":{"name":"default"},"__typename":"Repo"},"id":"720d471659dc4682a53576fdb637a482","branch":{"repo":{"name":"images","type":"user","project":{"name":"default"}},"name":"master"},"__typename":"Commit"}}\n' +
              '{"deleteFile":{"path":"/AT-AT.png"}}';
            if (body === expected) {
              return res(ctx.json({}));
            }
          },
        ),
      );
      server.use(mockFinishCommit());

      render(<ListViewTable files={MOCK_IMAGES_FILES} />);

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      await click((await screen.findAllByText('Delete'))[0]);

      const deleteConfirm = await screen.findByTestId('ModalFooter__confirm');
      await click(deleteConfirm);

      expect(window.location.pathname).toBe(
        '/project/default/repos/images/commit/720d471659dc4682a53576fdb637a482/',
      );
    });

    it('should delete multiple files on action click', async () => {
      server.use(mockStartCommit('720d471659dc4682a53576fdb637a482'));
      server.use(
        rest.post<string, Empty>(
          '/api/pfs_v2.API/ModifyFile',
          async (req, res, ctx) => {
            const body = await req.text();
            const expected =
              '{"setCommit":{"repo":{"name":"images","type":"user","project":{"name":"default"},"__typename":"Repo"},"id":"720d471659dc4682a53576fdb637a482","branch":{"repo":{"name":"images","type":"user","project":{"name":"default"}},"name":"master"},"__typename":"Commit"}}\n' +
              '{"deleteFile":{"path":"/AT-AT.png"}}\n' +
              '{"deleteFile":{"path":"/cats/"}}\n' +
              '{"deleteFile":{"path":"/liberty.png"}}';
            if (body === expected) {
              return res(ctx.json({}));
            }
          },
        ),
      );
      server.use(mockFinishCommit());

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

      expect(deleteButton).toBeEnabled();

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
        '/project/default/repos/images/commit/720d471659dc4682a53576fdb637a482/',
      );
    });

    it('should not allow file deletion for outputRepos', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/montage/commit/4a83c74809664f899261baccdb47cd90',
      );
      server.use(mockGetMontagePipeline());
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

      expect(deleteButton).toBeDisabled();
      await hover(deleteButton);
      expect(
        screen.getByText('You cannot delete files in an output repo'),
      ).toBeInTheDocument();

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      expect(screen.queryByText('Delete')).not.toBeInTheDocument();
    });

    it('should not allow file deletion for commits without branches', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/montage/commit/4a83c74809664f899261baccdb47cd90',
      );

      const noBranchFile = buildFile({
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
          },
          path: '/AT-AT.png',
          datum: 'default',
        },
        fileType: FileType.FILE,
        committed: '2023-11-08T18:12:19.363338Z',
        sizeBytes: '80590',
      });

      render(<ListViewTable files={[noBranchFile]} />);

      const deleteButton = screen.getByRole('button', {
        name: /delete selected items/i,
      });
      expect(deleteButton).toBeDisabled();

      await click(
        screen.getByRole('cell', {
          name: /at-at\.png/i,
        }),
      );

      expect(deleteButton).toBeDisabled();
      await hover(deleteButton);
      expect(
        screen.getByText(
          'You cannot delete files from a commit that is not associated with a branch',
        ),
      ).toBeInTheDocument();

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      expect(screen.queryByText('Delete')).not.toBeInTheDocument();
    });
  });

  describe('Download', () => {
    it('should disable the download button until selections are made', async () => {
      render(<ListViewTable files={MOCK_IMAGES_FILES} />);

      const downloadButton = screen.getByRole('button', {
        name: /download selected items/i,
      });

      expect(downloadButton).toBeDisabled();
      await hover(downloadButton);
      expect(
        screen.getByText('Select one or more files to multi-download files'),
      ).toBeInTheDocument();

      await click(
        screen.getByRole('cell', {
          name: /at-at\.png/i,
        }),
      );

      expect(downloadButton).toBeEnabled();
      await hover(downloadButton);
      expect(
        screen.queryByText('Select one or more files to multi-download files'),
      ).not.toBeInTheDocument();
    });

    it('should download a small file directly from pachd', async () => {
      const spy = jest.spyOn(window, 'open');

      render(<ListViewTable files={MOCK_IMAGES_FILES} />);

      // using .* regex for the "Change" column of the table
      const hamburger = screen.getByRole('button', {
        name: /at-at\.png.*80\.59 kb/i,
      });
      await click(hamburger);
      const downloadButton = within(hamburger).getByText(/Download/i);
      await click(downloadButton);

      expect(spy).toHaveBeenLastCalledWith(
        '/proxyForward/pfs/default/images/4a83c74809664f899261baccdb47cd90/AT-AT.png',
      );
    });

    it('should download multiple files on action click', async () => {
      const spy = jest.spyOn(window, 'open');

      server.use(
        mockEncode(['/AT-AT.png', '/cats/', '/json_nested_arrays.json']),
      );

      render(<ListViewTable files={MOCK_IMAGES_FILES} />);

      const downloadButton = screen.getByRole('button', {
        name: /download selected items/i,
      });

      expect(downloadButton).toBeDisabled();

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
          name: /json_nested_arrays\.json/i,
        }),
      );
      expect(downloadButton).toBeEnabled();
      await click(downloadButton);
      expect(spy).toHaveBeenLastCalledWith(
        '/proxyForward/archive/gCACFAwASxxgccGkVzh2qIdHkwutaaoJDgD.zip',
      );
    });

    it('should allow download for large files using zip', async () => {
      const spy = jest.spyOn(window, 'open');

      server.use(mockEncode(['/json_nested_arrays.json']));

      render(<ListViewTable files={MOCK_IMAGES_FILES} />);

      // using .* regex for the "Change" column of the table
      const hamburger = screen.getByRole('button', {
        name: /json_nested_arrays\.json.*200\.01 mb/i,
      });
      await click(hamburger);

      const downloadButton = within(hamburger).getByText(/Download Zip/i);

      await click(downloadButton);
      expect(spy).toHaveBeenLastCalledWith(
        '/proxyForward/archive/gCACFAwASxxgccGkVzh2qIdHkwutaaoJDgD.zip',
      );
    });

    it('should generate a download link', async () => {
      const spy = jest.spyOn(window, 'open');

      server.use(mockEncode(['/json_nested_arrays.json']));

      render(<ListViewTable files={MOCK_IMAGES_FILES} />);

      // using .* regex for the "Change" column of the table
      const hamburger = screen.getByRole('button', {
        name: /json_nested_arrays\.json.*200\.01 mb/i,
      });

      await click(hamburger);

      const downloadButton = within(hamburger).getByText(/Download Zip/i);

      await click(downloadButton);
      expect(spy).toHaveBeenLastCalledWith(
        '/proxyForward/archive/gCACFAwASxxgccGkVzh2qIdHkwutaaoJDgD.zip',
      );
    });
  });
});
