import {render, waitFor, screen} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockGetEnterpriseInfoInactive,
  mockGetVersionInfo,
  mockRepoImages,
} from '@dash-frontend/mocks';
import {
  type,
  withContextProviders,
  clear,
  click,
  upload,
} from '@dash-frontend/testHelpers';

import FileUploadComponent from '..';
import {GLOB_CHARACTERS} from '../lib/constants';

describe('views/FileUpload', () => {
  const server = setupServer();
  const FileUpload = withContextProviders(() => {
    return <FileUploadComponent />;
  });

  beforeAll(() => server.listen());

  beforeEach(() => {
    server.resetHandlers();
    server.use(mockGetVersionInfo());
    server.use(mockRepoImages());
    server.use(mockGetEnterpriseInfoInactive());
  });

  afterAll(() => server.close());

  it('should require a path', async () => {
    render(<FileUpload />);

    await waitFor(() =>
      expect(screen.getByLabelText('Attach Files')).toBeEnabled(),
    );

    const pathInput = await screen.findByLabelText('File Path');
    await waitFor(() => expect(pathInput).toBeEnabled());
    await clear(pathInput);

    expect(await screen.findByText('A path is required')).toBeInTheDocument();
  });

  it('should allow users to enter paths with only alphanumeric characters', async () => {
    render(<FileUpload />);

    await waitFor(() =>
      expect(screen.getByLabelText('Attach Files')).toBeEnabled(),
    );

    const pathInput = await screen.findByLabelText('File Path');
    await waitFor(() => expect(pathInput).toBeEnabled());
    await type(pathInput, '$');

    expect(
      screen.getByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).toBeInTheDocument();

    await clear(pathInput);
    expect(await screen.findByText('A path is required')).toBeInTheDocument();

    expect(
      screen.queryByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).not.toBeInTheDocument();
    expect(screen.getByText('A path is required')).toBeInTheDocument();

    await type(pathInput, '^');

    expect(
      screen.getByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).toBeInTheDocument();

    await clear(pathInput);
    expect(await screen.findByText('A path is required')).toBeInTheDocument();

    await type(pathInput, '/test123');

    expect(
      screen.queryByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).not.toBeInTheDocument();
  });

  it('should require paths start with a forward slash', async () => {
    render(<FileUpload />);

    await waitFor(() =>
      expect(screen.getByLabelText('Attach Files')).toBeEnabled(),
    );

    const pathInput = await screen.findByLabelText('File Path');
    await waitFor(() => expect(pathInput).toBeEnabled());

    await clear(pathInput);

    await type(pathInput, 't');

    expect(
      screen.getByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).toBeInTheDocument();

    await clear(pathInput);
    expect(await screen.findByText('A path is required')).toBeInTheDocument();

    await type(pathInput, '/test');

    expect(
      screen.queryByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).not.toBeInTheDocument();
  });

  it('should allow a user to input a file with the file input', async () => {
    render(<FileUpload />);

    await waitFor(() =>
      expect(screen.getByLabelText('Attach Files')).toBeEnabled(),
    );

    const fileInput = (await screen.findByLabelText(
      'Attach Files',
    )) as HTMLInputElement;

    await waitFor(() => expect(fileInput).toBeEnabled());

    await upload(fileInput, [
      new File(['hello'], 'hello.png', {type: 'image/png'}),
    ]);

    expect(
      await screen.findByText('hello.png', {}, {timeout: 10000}),
    ).toBeInTheDocument();
    expect(fileInput.files).toHaveLength(1);
  });

  it('should not allow a user to upload a file with regex metacharacters', async () => {
    render(<FileUpload />);

    await waitFor(() =>
      expect(screen.getByLabelText('Attach Files')).toBeEnabled(),
    );

    const fileInput = (await screen.findByLabelText(
      'Attach Files',
    )) as HTMLInputElement;

    await upload(fileInput, [
      new File(['hello'], 'hel^lo.png', {type: 'image/png'}),
    ]);

    expect(await screen.findByText('hel^lo.png')).toBeInTheDocument();
    expect(
      await screen.findByText(
        `Below file names cannot contain ${GLOB_CHARACTERS}. Please rename or re-upload.`,
      ),
    ).toBeInTheDocument();
  });

  it('should allow a user to submit a file', async () => {
    server.use(
      rest.post('/upload/start', (req, res, ctx) => {
        return res(ctx.json({uploadId: 1}));
      }),
    );
    server.use(
      rest.post('/upload', (req, res, ctx) => {
        return res(ctx.json({fileName: 'hello.txt'}));
      }),
    );
    server.use(
      rest.post('/upload/finish', (req, res, ctx) => {
        return res(ctx.json({commitId: '8ee8210b85174c9d8033c181228d0857'}));
      }),
    );

    render(<FileUpload />);

    const browseButton = screen.getByText('Browse Files');
    const submitButton = screen.getByLabelText('Upload Selected Files');
    const fileInput = screen.getByLabelText('Attach Files');
    const file = new File(['hello'], 'hello.txt', {type: 'text/plain'});

    await click(browseButton);
    await waitFor(() => expect(fileInput).toBeEnabled());
    await upload(fileInput, [file]);
    await screen.findByText('hello.txt');

    await waitFor(() =>
      expect(screen.getByLabelText('Upload Selected Files')).toBeEnabled(),
    );
    await click(submitButton);

    await screen.findByText(
      'All files have been successfully uploaded, click Done to commit the files.',
    );

    const commitButton = await screen.findByLabelText('Commit Selected Files');

    await click(commitButton);
    await waitFor(() => {
      // The upload modal should automatically close
      expect(
        screen.queryByLabelText('Commit Selected Files'),
      ).not.toBeInTheDocument();
    });
  });

  it('should show an error when there is an upload error', async () => {
    server.use(
      rest.post('/upload/start', (req, res, ctx) => {
        return res(ctx.status(500), ctx.json({error: 'Bad things happened!'}));
      }),
    );

    render(<FileUpload />);

    const submitButton = screen.getByLabelText('Upload Selected Files');
    const fileInput = screen.getByLabelText('Attach Files');
    const file = new File(['hello'], 'hello.txt', {type: 'text/plain'});

    await waitFor(() => expect(fileInput).toBeEnabled());
    await upload(fileInput, [file]);
    await screen.findByText('hello.txt');

    await waitFor(() =>
      expect(screen.getByLabelText('Upload Selected Files')).toBeEnabled(),
    );
    await click(submitButton);

    expect(
      await screen.findByText(
        'Sorry that we experienced some technical difficulties, Please try uploading again.',
      ),
    ).toBeInTheDocument();
  });
});
