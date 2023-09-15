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
    server.use(mockRepoWithLinkedPipeline());
  });

  afterAll(() => server.close());

  describe('File Preview File Types', () => {
    const renderFilePreview = (ext: string, contents = '') => {
      const fileName = `file.${ext}`;
      const file = buildFile({
        download: `/download/${fileName}`,
        path: `/${fileName}`,
        repoName: 'file',
      });

      server.use(
        rest.get(`/download/${fileName}`, (_req, res, ctx) => {
          return res(ctx.text(contents));
        }),
      );

      window.history.replaceState(
        {},
        '',
        `/project/default/repos/${file.repoName}/branch/master/commit/${file.commitId}${file.path}`,
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
      ).toHaveAttribute('src', `/download/file.${ext}`);
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
      ).toHaveAttribute('src', '/download/file.xml');

      const viewSourceButton = screen.getByTestId('Switch__buttonThumb');

      await click(viewSourceButton);

      // Should render the xml source
      expect(await screen.findByText('volume')).toBeInTheDocument();
    });

    it('should render an svg file', async () => {
      renderFilePreview('svg', files.svg);

      expect(
        await screen.findByTestId(`FilePreviewContent__image`),
      ).toHaveAttribute('src', '/download/file.svg');

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
