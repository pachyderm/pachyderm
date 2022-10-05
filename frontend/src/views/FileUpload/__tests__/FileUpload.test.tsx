import {render, waitFor} from '@testing-library/react';
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
    const {findByText, findByLabelText, getByLabelText} = render(
      <FileUpload />,
    );

    await waitFor(() =>
      expect(getByLabelText('Attach Files')).not.toBeDisabled(),
    );

    const pathInput = await findByLabelText('File Path');
    await waitFor(() => expect(pathInput).not.toBeDisabled());
    await clear(pathInput);

    expect(await findByText('A path is required')).toBeInTheDocument();
  });

  it('should allow users to enter paths with only alphanumeric characters', async () => {
    const {findByLabelText, queryByText, findByText, getByLabelText} = render(
      <FileUpload />,
    );

    await waitFor(() =>
      expect(getByLabelText('Attach Files')).not.toBeDisabled(),
    );

    const pathInput = await findByLabelText('File Path');
    await waitFor(() => expect(pathInput).not.toBeDisabled());
    await type(pathInput, '$');

    expect(
      queryByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).not.toBeNull();

    await clear(pathInput);
    expect(await findByText('A path is required')).toBeInTheDocument();

    expect(
      queryByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).toBeNull();
    expect(queryByText('A path is required')).not.toBeNull();

    await type(pathInput, '^');

    expect(
      queryByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).not.toBeNull();

    await clear(pathInput);
    expect(await findByText('A path is required')).toBeInTheDocument();

    await type(pathInput, '/test123');

    expect(
      queryByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).toBeNull();
  });

  it('should require paths start with a forward slash', async () => {
    const {findByLabelText, queryByText, findByText, getByLabelText} = render(
      <FileUpload />,
    );

    await waitFor(() =>
      expect(getByLabelText('Attach Files')).not.toBeDisabled(),
    );

    const pathInput = await findByLabelText('File Path');
    await waitFor(() => expect(pathInput).not.toBeDisabled());

    await clear(pathInput);

    await type(pathInput, 't');

    expect(
      await queryByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).not.toBeNull();

    await clear(pathInput);
    expect(await findByText('A path is required')).toBeInTheDocument();

    await type(pathInput, '/test');

    expect(
      await queryByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).toBeNull();
  });

  it('should allow a user to input a file with the file input', async () => {
    const {findByLabelText, findByText, getByLabelText} = render(
      <FileUpload />,
    );

    await waitFor(() =>
      expect(getByLabelText('Attach Files')).not.toBeDisabled(),
    );

    const fileInput = (await findByLabelText(
      'Attach Files',
    )) as HTMLInputElement;

    await waitFor(() => expect(fileInput).not.toBeDisabled());

    await upload(fileInput, [
      new File(['hello'], 'hello.png', {type: 'image/png'}),
    ]);

    expect(await findByText('hello.png', {}, {timeout: 10000})).not.toBeNull();
    expect(fileInput.files?.length).toEqual(1);
  });

  it('should not allow a user to upload a file with regex metacharacters', async () => {
    const {findByLabelText, findByText, getByLabelText} = render(
      <FileUpload />,
    );

    await waitFor(() =>
      expect(getByLabelText('Attach Files')).not.toBeDisabled(),
    );

    const fileInput = (await findByLabelText(
      'Attach Files',
    )) as HTMLInputElement;

    await upload(fileInput, [
      new File(['hello'], 'hel^lo.png', {type: 'image/png'}),
    ]);

    expect(await findByText('hel^lo.png')).toBeInTheDocument();
    expect(
      await findByText(
        `Below file names cannot contain ${GLOB_CHARACTERS}. Please rename or re-upload.`,
      ),
    ).not.toBeNull();
  });
});
