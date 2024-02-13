import {
  render,
  screen,
  within,
  waitFor,
  waitForElementToBeRemoved,
} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {FileType} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import {
  mockGetEnterpriseInfo,
  buildFile,
  mockStartCommit,
  mockFinishCommit,
  mockGetMontagePipeline,
  mockEmptyInspectPipeline,
  mockGetImageCommitsNoBranch,
} from '@dash-frontend/mocks';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import FilePreviewComponent from '../';

import files from './files';

describe('File Preview', () => {
  const server = setupServer();
  const FilePreview = withContextProviders<typeof FilePreviewComponent>(
    ({file}) => {
      return <FilePreviewComponent file={file} />;
    },
  );

  beforeAll(() => server.listen());

  beforeEach(() => {
    server.resetHandlers();
    server.use(mockGetEnterpriseInfo());
    server.use(mockEmptyInspectPipeline());
  });

  afterAll(() => server.close());

  describe('File Preview File Types', () => {
    const renderFilePreview = (
      ext: string,
      contents = '',
      fullFileName = '',
    ) => {
      const fileName = fullFileName || `file.${ext}`;

      const file = buildFile({
        file: {
          commit: {
            repo: {
              name: 'file',
              type: 'user',
              project: {
                name: 'default',
              },
            },
            id: '252d1850a5fa484ca7320ce1091cf483',
            branch: {
              repo: {
                name: 'file',
                type: 'user',
                project: {
                  name: 'default',
                },
              },
              name: 'master',
            },
          },
          path: `/${fileName}`,
          datum: 'default',
        },
        fileType: FileType.FILE,
        committed: '2023-11-08T18:12:19.363338Z',
        sizeBytes: '80588',
      });

      server.use(
        rest.get(
          `/proxyForward/pfs/default/file/252d1850a5fa484ca7320ce1091cf483/${fileName}`,
          (_req, res, ctx) => {
            return res(ctx.text(contents));
          },
        ),
      );

      window.history.replaceState(
        {},
        '',
        `/project/default/repos/file/commit/252d1850a5fa484ca7320ce1091cf483/${fileName}`,
      );

      render(<FilePreview file={file} />);
    };

    it.each([
      'markdn',
      'markdown',
      'md',
      'mdown',
      'mdwn',
      'mkdn',
      'mkdown',
      'mkd',
    ])('should render a %s file', async (ext) => {
      renderFilePreview(ext, files.markdown);

      // Should render the markdown contents
      expect(await screen.findByText('H1')).toBeInTheDocument();
      expect(await screen.findByText('H2')).toBeInTheDocument();
      expect(await screen.findByText('H3')).toBeInTheDocument();
      expect(
        await screen.findByText('labore et dolore magna aliqua'),
      ).toBeInTheDocument();

      // Should not render markdown source
      expect(screen.queryByText('#')).not.toBeInTheDocument();
      expect(screen.queryByText('##')).not.toBeInTheDocument();
      expect(screen.queryByText('###')).not.toBeInTheDocument();
      expect(screen.queryByText('---')).not.toBeInTheDocument();
      expect(screen.queryByText('**')).not.toBeInTheDocument();

      const viewSourceButton = screen.getByTestId('Switch__buttonThumb');

      await click(viewSourceButton);

      // Should render the markdown source
      expect(await screen.findByText('#')).toBeInTheDocument();
      expect(await screen.findByText('##')).toBeInTheDocument();
      expect(await screen.findByText('###')).toBeInTheDocument();
      expect(await screen.findByText('---')).toBeInTheDocument();
      expect(await screen.findAllByText('**')).toHaveLength(2);
    });

    it.each([
      {name: 'image', ext: 'png'},
      {name: 'video', ext: 'mp4'},
      {name: 'audio', ext: 'mp3'},
    ])('should render a $name file', async ({name, ext}) => {
      renderFilePreview(ext, '');

      expect(
        await screen.findByTestId(`FilePreviewContent__${name}`),
      ).toHaveAttribute(
        'src',
        `/proxyForward/pfs/default/file/252d1850a5fa484ca7320ce1091cf483/file.${ext}`,
      );
    });

    it.each(['yml', 'yaml'])('should render a %s file', async (ext) => {
      renderFilePreview(ext, files.yaml);

      expect(await screen.findByText('hello')).toBeInTheDocument();
      expect(await screen.findByText('world')).toBeInTheDocument();
    });

    it.each(['txt', 'jsonl', 'textpb'])(
      'should render a %s file',
      async (ext) => {
        renderFilePreview(ext, `mock ${ext}\n`);

        expect(await screen.findByText(`mock ${ext}`)).toBeInTheDocument();
      },
    );

    it('should not render an error response from pfs', async () => {
      renderFilePreview('txt', `problem inspecting file: rpc error: code `);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      expect(
        screen.queryByText('problem inspecting file: rpc error: code'),
      ).not.toBeInTheDocument();
    });

    it.each(['htm', 'html'])('should render an %s file', async (ext) => {
      renderFilePreview(ext, files.html);

      expect(await screen.findByTestId('WebPreview__iframe')).toContainHTML(
        '<html><body>hello world</body></html>',
      );

      const viewSourceButton = screen.getByTestId('Switch__buttonThumb');

      await click(viewSourceButton);

      // Should render the html source
      expect(await screen.findByText('<!DOCTYPE html>')).toBeInTheDocument();
    });

    it('should render an xml file', async () => {
      renderFilePreview('xml', files.xml);

      expect(
        await screen.findByTestId(`FilePreviewContent__xml`),
      ).toHaveAttribute(
        'src',
        '/proxyForward/pfs/default/file/252d1850a5fa484ca7320ce1091cf483/file.xml',
      );

      const viewSourceButton = screen.getByTestId('Switch__buttonThumb');

      await click(viewSourceButton);

      // Should render the xml source
      expect(await screen.findByText('volume')).toBeInTheDocument();
    });

    it('should render an svg file', async () => {
      renderFilePreview('svg', files.svg);

      expect(
        await screen.findByTestId(`FilePreviewContent__image`),
      ).toHaveAttribute(
        'src',
        '/proxyForward/pfs/default/file/252d1850a5fa484ca7320ce1091cf483/file.svg',
      );

      const viewSourceButton = screen.getByTestId('Switch__buttonThumb');

      await click(viewSourceButton);

      // Should render the xml source
      expect(await screen.findByText('fill')).toBeInTheDocument();
    });

    it('should render a css file', async () => {
      renderFilePreview('css', files.css);

      expect(await screen.findByText('codePreview')).toBeInTheDocument();
    });

    it('should render a c file', async () => {
      renderFilePreview('c', files.c);

      expect(await screen.findByText('<stdio.h>')).toBeInTheDocument();
    });

    it('should render a cpp file', async () => {
      renderFilePreview('cpp', files.cpp);

      expect(await screen.findByText('<iostream>')).toBeInTheDocument();
    });

    it('should render a java file', async () => {
      renderFilePreview('java', files.java);

      expect(await screen.findByText('System')).toBeInTheDocument();
    });

    it('should render a php file', async () => {
      renderFilePreview('php', files.php);

      expect(await screen.findByText('echo')).toBeInTheDocument();
    });

    it('should render a rust file', async () => {
      renderFilePreview('rs', files.rs);

      expect(await screen.findByText('main')).toBeInTheDocument();
    });

    it('should render an sql file', async () => {
      renderFilePreview('sql', files.sql);

      expect(await screen.findByText('CREATE')).toBeInTheDocument();
    });

    it('should render a json file', async () => {
      renderFilePreview('json', files.json);

      expect(await screen.findByText('"hello"')).toBeInTheDocument();
      expect(await screen.findByText(':')).toBeInTheDocument();
      expect(await screen.findByText('"world"')).toBeInTheDocument();
    });

    it('should render a csv file', async () => {
      renderFilePreview('csv', files.csv);

      expect(await screen.findByText('Separator: comma')).toBeInTheDocument();
      expect(await screen.findByText('hello')).toBeInTheDocument();
      expect(await screen.findByText('world')).toBeInTheDocument();
      expect(await screen.findByText('"1"')).toBeInTheDocument();
      expect(await screen.findByText('"2"')).toBeInTheDocument();
    });

    it.each(['tsv', 'tab'])('should render a %s file', async (ext) => {
      renderFilePreview(ext, files.tsv);

      expect(await screen.findByText('Separator: tab')).toBeInTheDocument();
      expect(await screen.findByText('hello')).toBeInTheDocument();
      expect(await screen.findByText('world')).toBeInTheDocument();
      expect(await screen.findByText('"1"')).toBeInTheDocument();
      expect(await screen.findByText('"2"')).toBeInTheDocument();
    });

    it('should render a python file', async () => {
      renderFilePreview('py', files.python);

      expect(await screen.findByText('print')).toBeInTheDocument();
    });

    it('should render a javascript file', async () => {
      renderFilePreview('tsx', files.javascript);

      expect(await screen.findByText('React')).toBeInTheDocument();
    });

    it('should render a go file', async () => {
      renderFilePreview('go', files.go);

      expect(await screen.findByText('Println')).toBeInTheDocument();
    });

    it('should render a shell file', async () => {
      renderFilePreview('sh', files.shell);

      expect(await screen.findByText('#!/bin/bash')).toBeInTheDocument();
    });

    it('should render an r file', async () => {
      renderFilePreview('r', files.r);

      expect(await screen.findByText('print')).toBeInTheDocument();
    });

    it('should render a julia file', async () => {
      renderFilePreview('jl', files.julia);

      expect(await screen.findByText('println')).toBeInTheDocument();
    });

    it('should render a ruby file', async () => {
      renderFilePreview('rb', files.ruby);

      expect(await screen.findByText('puts')).toBeInTheDocument();
    });

    it('should render a Dockerfile', async () => {
      renderFilePreview('', files.docker, 'Dockerfile');

      expect(await screen.findByText('EXPOSE')).toBeInTheDocument();
    });

    it('should render a protobuf file', async () => {
      renderFilePreview('proto', files.protobuf);

      expect(await screen.findByText('syntax')).toBeInTheDocument();
    });

    it('should render a message when the file type cannot be rendered', async () => {
      const file = buildFile({
        file: {
          commit: {
            repo: {
              name: 'lots-of-commits',
              type: 'user',
              project: {
                name: 'default',
              },
            },
            id: '252d1850a5fa484ca7320ce1091cf483',
            branch: {
              repo: {
                name: 'lots-of-commits',
                type: 'user',
                project: {
                  name: 'default',
                },
              },
              name: 'master',
            },
          },
          path: '/data.unsupported',
          datum: 'default',
        },
        fileType: FileType.FILE,
        committed: '2023-11-08T18:12:19.363338Z',
        sizeBytes: '80588',
      });

      window.history.replaceState(
        {},
        '',
        `/project/${file.file?.commit?.repo?.project?.name}/repos/${file.file?.commit?.repo?.name}/commit/${file.file?.commit?.id}${file.file?.path}`,
      );

      render(<FilePreview file={file} />);

      expect(
        await screen.findByText('Unable to preview this file'),
      ).toBeInTheDocument();
      expect(
        await screen.findByText(
          'This file format is not supported for file previews',
        ),
      ).toBeInTheDocument();

      const viewRawLink = await screen.findByText('View Raw');

      expect(viewRawLink).toHaveAttribute(
        'href',
        '/proxyForward/pfs/default/lots-of-commits/252d1850a5fa484ca7320ce1091cf483/data.unsupported',
      );
      expect(viewRawLink).toHaveAttribute('target', '_blank');
    });

    it('should render a message when the file is too large to be rendered', async () => {
      const file = buildFile({
        file: {
          commit: {
            repo: {
              name: 'text',
              type: 'user',
              project: {
                name: 'default',
              },
            },
            id: '252d1850a5fa484ca7320ce1091cf483',
            branch: {
              repo: {
                name: 'text',
                type: 'user',
                project: {
                  name: 'default',
                },
              },
              name: 'master',
            },
          },
          path: '/data.txt',
          datum: 'default',
        },
        fileType: FileType.FILE,
        committed: '2023-11-08T18:12:19.363338Z',
        sizeBytes: '500000000',
      });

      window.history.replaceState(
        {},
        '',
        `/project/${file.file?.commit?.repo?.project?.name}/repos/${file.file?.commit?.repo?.name}/commit/${file.file?.commit?.id}${file.file?.path}`,
      );

      render(<FilePreview file={file} />);

      expect(
        await screen.findByText('Unable to preview this file'),
      ).toBeInTheDocument();
      expect(
        await screen.findByText('This file is too large to preview'),
      ).toBeInTheDocument();
    });

    it('should render file metadata', async () => {
      const file = buildFile({
        file: {
          commit: {
            repo: {
              name: 'image',
              type: 'user',
              project: {
                name: 'default',
              },
            },
            id: '252d1850a5fa484ca7320ce1091cf483',
            branch: {
              repo: {
                name: 'image',
                type: 'user',
                project: {
                  name: 'default',
                },
              },
              name: 'master',
            },
          },
          path: '/image.png',
          datum: 'default',
        },
        fileType: FileType.FILE,
        committed: '2023-11-08T18:12:19.363338Z',
        sizeBytes: '58650',
      });

      window.history.replaceState(
        {},
        '',
        `/project/${file.file?.commit?.repo?.project?.name}/repos/${file.file?.commit?.repo?.name}/commit/${file.file?.commit?.id}${file.file?.path}`,
      );

      render(<FilePreview file={file} />);

      const metadata = await screen.findByLabelText('file metadata');

      expect(within(metadata).getByText('master')).toBeInTheDocument();
      expect(within(metadata).getByText('png')).toBeInTheDocument();
      expect(within(metadata).getByText('58.65 kB')).toBeInTheDocument();
      expect(within(metadata).getByText('/image.png')).toBeInTheDocument();
    });
  });

  describe('File Preview Actions', () => {
    const file = buildFile({
      file: {
        commit: {
          repo: {
            name: 'image',
            type: 'user',
            project: {
              name: 'default',
            },
          },
          id: '252d1850a5fa484ca7320ce1091cf483',
          branch: {
            repo: {
              name: 'image',
              type: 'user',
              project: {
                name: 'default',
              },
            },
            name: 'master',
          },
        },
        path: '/image.png',
        datum: 'default',
      },
      fileType: FileType.FILE,
      committed: '2023-11-08T18:12:19.363338Z',
      sizeBytes: '58650',
    });

    window.history.replaceState(
      {},
      '',
      '/project/default/repos/image/commit/default/image.png',
    );

    it('should navigate to parent path on button link', async () => {
      render(<FilePreview file={file} />);

      await click(
        await screen.findByRole('button', {
          name: 'Back to file list',
        }),
      );

      expect(window.location.pathname).toBe(
        '/project/default/repos/image/commit/252d1850a5fa484ca7320ce1091cf483/',
      );
    });

    it('should copy path on action click', async () => {
      render(<FilePreview file={file} />);

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      await click((await screen.findAllByText('Copy Path'))[0]);

      expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
        'image@252d1850a5fa484ca7320ce1091cf483:/image.png',
      );
    });

    it('should not allow file deletion for output repos', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/montage/commit/default/image.png',
      );
      server.use(mockGetMontagePipeline());
      render(<FilePreview file={file} />);

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      expect(screen.queryByText('Delete')).not.toBeInTheDocument();
    });

    it('should not allow file deletion for commits with no branch', async () => {
      const file = buildFile({
        file: {
          commit: {
            repo: {
              name: 'image',
              type: 'user',
              project: {
                name: 'default',
              },
            },
            id: '252d1850a5fa484ca7320ce1091cf483',
          },
          path: '/image.png',
          datum: 'default',
        },
        fileType: FileType.FILE,
        committed: '2023-11-08T18:12:19.363338Z',
        sizeBytes: '58650',
      });

      window.history.replaceState(
        {},
        '',
        '/project/default/repos/montage/commit/default/image.png',
      );
      server.use(mockGetImageCommitsNoBranch());
      render(<FilePreview file={file} />);

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      expect(screen.queryByText('Delete')).not.toBeInTheDocument();
    });

    it('should delete file on action click', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/repos/image/commit/default/image.png',
      );
      server.use(mockEmptyInspectPipeline());
      server.use(mockStartCommit('720d471659dc4682a53576fdb637a482'));
      server.use(
        rest.post<string, Empty>(
          '/api/pfs_v2.API/ModifyFile',
          async (req, res, ctx) => {
            const body = await req.text();
            const expected =
              '{"setCommit":{"repo":{"name":"images","type":"user","project":{"name":"default"},"__typename":"Repo"},"id":"720d471659dc4682a53576fdb637a482","branch":{"repo":{"name":"images","type":"user","project":{"name":"default"}},"name":"master"},"__typename":"Commit"}}\n' +
              '{"deleteFile":{"path":"/image.png"}}';
            if (body === expected) {
              return res(ctx.json({}));
            }
          },
        ),
      );
      server.use(mockFinishCommit());

      render(<FilePreview file={file} />);

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      await click((await screen.findAllByText('Delete'))[0]);

      const deleteConfirm = await screen.findByTestId('ModalFooter__confirm');

      await click(deleteConfirm);

      // The delete is finished after navigating to the new commit
      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/default/repos/image/commit/720d471659dc4682a53576fdb637a482/',
        ),
      );
    });
  });
});
