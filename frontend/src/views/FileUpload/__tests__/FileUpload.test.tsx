import {readFileSync} from 'fs';
import path from 'path';

import {render, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

import {click, type, withContextProviders} from '@dash-frontend/testHelpers';

import FileUploadComponent from '..';

describe('File Upload', () => {
  beforeEach(() => {
    window.history.replaceState({}, '', '/lineage/3/repos/cron/upload');
  });

  const FileUpload = withContextProviders(() => {
    return <FileUploadComponent />;
  });

  it('should require a path', async () => {
    const {findByText, findByLabelText} = render(<FileUpload />);

    const pathInput = await findByLabelText('File Path');
    userEvent.clear(pathInput);

    expect(await findByText('A path is required')).toBeInTheDocument();
  });

  it('should allow users to enter paths with only alphanumeric characters', async () => {
    const {findByLabelText, queryByText, findByText} = render(<FileUpload />);

    const pathInput = await findByLabelText('File Path');
    await type(pathInput, '$');

    expect(
      queryByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).not.toBeNull();

    userEvent.clear(pathInput);
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

    userEvent.clear(pathInput);
    expect(await findByText('A path is required')).toBeInTheDocument();

    await type(pathInput, '/test123');

    expect(
      queryByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).toBeNull();
  });

  it('should require paths start with a forward slash', async () => {
    const {findByLabelText, queryByText, findByText} = render(<FileUpload />);

    const pathInput = await findByLabelText('File Path');
    userEvent.clear(pathInput);

    await type(pathInput, 't');

    expect(
      await queryByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).not.toBeNull();

    userEvent.clear(pathInput);
    expect(await findByText('A path is required')).toBeInTheDocument();

    await type(pathInput, '/test');

    expect(
      await queryByText(
        'Paths can only contain alphanumeric characters and must start with a forward slash',
      ),
    ).toBeNull();
  });

  it('should allow a user to input a file with the file input', async () => {
    const {findByLabelText, queryByText, getByLabelText} = render(
      <FileUpload />,
    );

    await waitFor(() =>
      expect(getByLabelText('Attach File')).not.toBeDisabled(),
    );

    const fileInput = (await findByLabelText(
      'Attach File',
    )) as HTMLInputElement;

    userEvent.upload(
      fileInput,
      new File(['hello'], 'hello.png', {type: 'image/png'}),
    );

    expect(queryByText('hello.png')).not.toBeNull();
    expect(fileInput.files).toHaveLength(1);
  });

  it('should not allow a user to upload a file with regex metacharacters', async () => {
    const {findByLabelText, findByText, getByLabelText} = render(
      <FileUpload />,
    );

    await waitFor(() =>
      expect(getByLabelText('Attach File')).not.toBeDisabled(),
    );

    const fileInput = (await findByLabelText(
      'Attach File',
    )) as HTMLInputElement;

    userEvent.upload(
      fileInput,
      new File(['hello'], 'hel^lo.png', {type: 'image/png'}),
    );

    expect(await findByText('hel^lo.png')).toBeInTheDocument();
    expect(
      await findByText('File names cannot contain regex metacharacters.'),
    ).not.toBeNull();
  });

  it('should allow a user to clear the file input', async () => {
    const {
      findByLabelText,
      queryByText,
      getByLabelText,
      findByRole,
      findByTestId,
    } = render(<FileUpload />);

    await waitFor(() =>
      expect(getByLabelText('Attach File')).not.toBeDisabled(),
    );

    const fileInput = (await findByLabelText(
      'Attach File',
    )) as HTMLInputElement;

    userEvent.upload(
      fileInput,
      new File(['hello'], 'hello.png', {type: 'image/png'}),
    );

    expect(queryByText('hello.png')).not.toBeNull();

    const removeButton = await findByTestId('UploadInfo__cancel');
    click(removeButton);

    expect(await findByRole('button', {name: 'Upload'})).toBeDisabled();
    expect(queryByText('hello.png')).toBeNull();
  });
});
