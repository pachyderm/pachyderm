import {render, waitFor, screen} from '@testing-library/react';
import React from 'react';

import {
  type,
  withContextProviders,
  clear,
  upload,
} from '@dash-frontend/testHelpers';

import FileUploadComponent from '..';
import {GLOB_CHARACTERS} from '../lib/constants';

describe('File Upload', () => {
  beforeEach(() => {
    window.history.replaceState({}, '', '/lineage/3/repos/cron/upload');
  });

  const FileUpload = withContextProviders(() => {
    return <FileUploadComponent />;
  });

  it('should require a path', async () => {
    render(<FileUpload />);

    await waitFor(() =>
      expect(screen.getByLabelText('Attach Files')).not.toBeDisabled(),
    );

    const pathInput = await screen.findByLabelText('File Path');
    await waitFor(() => expect(pathInput).not.toBeDisabled());
    await clear(pathInput);

    expect(await screen.findByText('A path is required')).toBeInTheDocument();
  });

  it('should allow users to enter paths with only alphanumeric characters', async () => {
    render(<FileUpload />);

    await waitFor(() =>
      expect(screen.getByLabelText('Attach Files')).not.toBeDisabled(),
    );

    const pathInput = await screen.findByLabelText('File Path');
    await waitFor(() => expect(pathInput).not.toBeDisabled());
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
      expect(screen.getByLabelText('Attach Files')).not.toBeDisabled(),
    );

    const pathInput = await screen.findByLabelText('File Path');
    await waitFor(() => expect(pathInput).not.toBeDisabled());

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
      expect(screen.getByLabelText('Attach Files')).not.toBeDisabled(),
    );

    const fileInput = (await screen.findByLabelText(
      'Attach Files',
    )) as HTMLInputElement;

    await waitFor(() => expect(fileInput).not.toBeDisabled());

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
      expect(screen.getByLabelText('Attach Files')).not.toBeDisabled(),
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
});
