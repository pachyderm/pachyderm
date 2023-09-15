import {render, screen, within, waitFor} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  buildFile,
  mockDeleteFiles,
  mockRepoWithNullLinkedPipeline,
  mockRepoWithLinkedPipeline,
} from '@dash-frontend/mocks';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import FilePreviewComponent from '../';

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
    server.use(mockRepoWithLinkedPipeline());
  });

  afterAll(() => server.close());

  describe('File Preview File Types', () => {
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
      const file = buildFile({
        download: `/download/markdown.${ext}`,
        path: `/markdown.${ext}`,
        repoName: 'markdown',
      });

      server.use(
        rest.get(`/download/markdown.${ext}`, (_req, res, ctx) => {
          return res(
            ctx.text(
              '# H1\n## H2\n### H3\n---\n**labore et dolore magna aliqua**',
            ),
          );
        }),
      );

      window.history.replaceState(
        {},
        '',
        `/project/default/repos/${file.repoName}/branch/master/commit/${file.commitId}${file.path}`,
      );

      render(<FilePreview file={file} />);

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
      {name: 'xml', ext: 'xml'},
    ])('should render a $name file', async ({name, ext}) => {
      const fileName = `${name}.${ext}`;
      const file = buildFile({
        download: `/download/${fileName}`,
        path: `/${fileName}`,
        repoName: name,
      });

      window.history.replaceState(
        {},
        '',
        `/project/default/repos/${file.repoName}/branch/master/commit/${file.commitId}${file.path}`,
      );

      render(<FilePreview file={file} />);

      expect(
        await screen.findByTestId(`FilePreviewContent__${name}`),
      ).toHaveAttribute('src', file.download);
    });

    it.each(['yml', 'yaml'])('should render a %s file', async (ext) => {
      const file = buildFile({
        download: `/download/data.${ext}`,
        path: `/data.${ext}`,
        repoName: 'yaml',
      });

      server.use(
        rest.get(`/download/data.${ext}`, (_req, res, ctx) => {
          return res(ctx.text('hello: world\n'));
        }),
      );

      window.history.replaceState(
        {},
        '',
        `/project/default/repos/${file.repoName}/branch/master/commit/${file.commitId}${file.path}`,
      );

      render(<FilePreview file={file} />);

      expect(await screen.findByText('hello')).toBeInTheDocument();
      expect(await screen.findByText('world')).toBeInTheDocument();
    });

    it.each(['txt', 'jsonl', 'textpb'])(
      'should render a %s file',
      async (ext) => {
        const file = buildFile({
          download: `/download/data.${ext}`,
          path: `/data.${ext}`,
          repoName: ext,
        });

        server.use(
          rest.get(`/download/data.${ext}`, (_req, res, ctx) => {
            return res(ctx.text(`mock ${ext}\n`));
          }),
        );

        window.history.replaceState(
          {},
          '',
          `/project/default/repos/${file.repoName}/branch/master/commit/${file.commitId}${file.path}`,
        );

        render(<FilePreview file={file} />);

        expect(await screen.findByText(`mock ${ext}`)).toBeInTheDocument();
      },
    );

    it.each(['htm', 'html'])('should render an %s file', async (ext) => {
      const file = buildFile({
        download: `/download/index.${ext}`,
        path: `/index.${ext}`,
        repoName: ext,
      });

      server.use(
        rest.get(`/download/index.${ext}`, (_req, res, ctx) => {
          return res(ctx.text(`<html>helloworld</html>`));
        }),
      );

      window.history.replaceState(
        {},
        '',
        `/project/default/repos/${file.repoName}/branch/master/commit/${file.commitId}${file.path}`,
      );

      render(<FilePreview file={file} />);

      expect(await screen.findByTestId('WebPreview__iframe')).toContainHTML(
        '<html>helloworld</html>',
      );
    });

    it('should render a json file', async () => {
      const file = buildFile({
        download: '/download/data.json',
        path: 'data.json',
        repoName: 'json',
      });

      server.use(
        rest.get(`/download/data.json`, (_req, res, ctx) => {
          return res(ctx.text(JSON.stringify({hello: 'world'})));
        }),
      );

      window.history.replaceState(
        {},
        '',
        `/project/default/repos/${file.repoName}/branch/master/commit/${file.commitId}${file.path}`,
      );

      render(<FilePreview file={file} />);

      expect(await screen.findByText('"hello"')).toBeInTheDocument();
      expect(await screen.findByText(':')).toBeInTheDocument();
      expect(await screen.findByText('"world"')).toBeInTheDocument();
    });

    it('should render a csv file', async () => {
      const file = buildFile({
        download: '/download/data.csv',
        path: 'data.csv',
        repoName: 'csv',
      });

      server.use(
        rest.get(`/download/data.csv`, (_req, res, ctx) => {
          return res(ctx.text('hello,world\n1,2\n'));
        }),
      );

      window.history.replaceState(
        {},
        '',
        `/project/default/repos/${file.repoName}/branch/master/commit/${file.commitId}${file.path}`,
      );

      render(<FilePreview file={file} />);

      expect(await screen.findByText('Separator: comma')).toBeInTheDocument();
      expect(await screen.findByText('hello')).toBeInTheDocument();
      expect(await screen.findByText('world')).toBeInTheDocument();
      expect(await screen.findByText('"1"')).toBeInTheDocument();
      expect(await screen.findByText('"2"')).toBeInTheDocument();
    });

    it.each(['tsv', 'tab'])('should render a %s file', async (ext) => {
      const file = buildFile({
        download: `/download/data.${ext}`,
        path: `/data.${ext}`,
        repoName: ext,
      });

      server.use(
        rest.get(`/download/data.${ext}`, (_req, res, ctx) => {
          return res(ctx.text(`hello\tworld\n1\t2\n`));
        }),
      );

      window.history.replaceState(
        {},
        '',
        `/project/default/repos/${file.repoName}/branch/master/commit/${file.commitId}${file.path}`,
      );

      render(<FilePreview file={file} />);

      expect(await screen.findByText('Separator: tab')).toBeInTheDocument();
      expect(await screen.findByText('hello')).toBeInTheDocument();
      expect(await screen.findByText('world')).toBeInTheDocument();
      expect(await screen.findByText('"1"')).toBeInTheDocument();
      expect(await screen.findByText('"2"')).toBeInTheDocument();
    });

    it('should render a python file', async () => {
      const file = buildFile({
        download: '/download/index.py',
        path: 'index.py',
        repoName: 'py',
      });

      server.use(
        rest.get(`/download/index.py`, (_req, res, ctx) => {
          return res(ctx.text('x = 1\nif x == 1:\n\tprint(x)'));
        }),
      );

      window.history.replaceState(
        {},
        '',
        `/project/default/repos/${file.repoName}/branch/master/commit/${file.commitId}${file.path}`,
      );

      render(<FilePreview file={file} />);

      expect(
        await screen.findByText((_content, element) => {
          return element?.textContent === 'x = 1';
        }),
      ).toBeInTheDocument();
      expect(
        await screen.findByText((_content, element) => {
          return element?.textContent === 'if x == 1:';
        }),
      ).toBeInTheDocument();
      expect(
        await screen.findByText((_content, element) => {
          return element?.textContent === '\tprint(x)';
        }),
      ).toBeInTheDocument();
    });

    it('should render a message when the file type cannot be rendered', async () => {
      const file = buildFile({
        download: '/download/data.unsupported',
        path: 'data.unsupported',
        repoName: 'unsupported',
      });

      window.history.replaceState(
        {},
        '',
        `/project/default/repos/${file.repoName}/branch/master/commit/${file.commitId}${file.path}`,
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
    });

    it('should render a message when the file is too large to be rendered', async () => {
      const file = buildFile({
        download: null,
        path: 'data.txt',
        repoName: 'text',
        sizeBytes: 5000000,
        sizeDisplay: '5 MB',
      });

      window.history.replaceState(
        {},
        '',
        `/project/default/repos/${file.repoName}/branch/master/commit/${file.commitId}${file.path}`,
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
        download: '/download/image.png',
        path: '/image.png',
        repoName: 'image',
        sizeDisplay: '58.65 kB',
      });

      window.history.replaceState(
        {},
        '',
        `/project/default/repos/${file.repoName}/branch/master/commit/${file.commitId}${file.path}`,
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
      download: '/download/image.png',
      path: '/image.png',
      repoName: 'image',
      sizeDisplay: '58.65 kB',
    });

    window.history.replaceState(
      {},
      '',
      '/project/default/repos/image/branch/master/commit/default/image.png',
    );

    it('should navigate to parent path on button link', async () => {
      render(<FilePreview file={file} />);

      await click(
        await screen.findByRole('button', {
          name: 'Back to file list',
        }),
      );

      expect(window.location.pathname).toBe(
        '/project/default/repos/image/branch/master/commit/default/',
      );
    });

    it('should copy path on action click', async () => {
      render(<FilePreview file={file} />);

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      await click((await screen.findAllByText('Copy Path'))[0]);

      expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
        'image@master=default:/image.png',
      );
    });

    it('should not allow file deletion for output repos', async () => {
      server.use(mockRepoWithLinkedPipeline());
      render(<FilePreview file={file} />);

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      expect(screen.queryByText('Delete')).not.toBeInTheDocument();
    });

    it('should delete file on action click', async () => {
      server.use(mockRepoWithNullLinkedPipeline());
      server.use(mockDeleteFiles());

      render(<FilePreview file={file} />);

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      await click((await screen.findAllByText('Delete'))[0]);

      const deleteConfirm = await screen.findByTestId('ModalFooter__confirm');

      await click(deleteConfirm);

      // The delete is finished after navigating to the new commit
      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/default/repos/image/branch/master/commit/deleted/',
        ),
      );
    });
  });
});
