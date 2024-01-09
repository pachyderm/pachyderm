import {render, screen, within} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {CreateRepoRequest} from '@dash-frontend/api/pfs';
import {RequestError} from '@dash-frontend/api/utils/error';
import {mockRepos} from '@dash-frontend/mocks';
import {withContextProviders, type, click} from '@dash-frontend/testHelpers';

import CreateRepoModalComponent from '../CreateRepoModal';

const server = setupServer();

describe('CreateRepoModal', () => {
  const CreateRepoModal = withContextProviders(() => {
    return <CreateRepoModalComponent show={true} />;
  });

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    server.resetHandlers();
    server.use(mockRepos());
    window.history.replaceState('', '', '/lineage/default');
  });

  afterAll(() => server.close());

  it('should error if the repo already exists', async () => {
    render(<CreateRepoModal />);

    const nameInput = await screen.findByRole('textbox', {name: /name/i});

    await type(nameInput, 'images');

    expect(
      await screen.findByText('Repo name already in use'),
    ).toBeInTheDocument();
  });

  it('should display an error message if mutation fails', async () => {
    server.use(
      rest.post<CreateRepoRequest, Empty, RequestError>(
        '/api/pfs_v2.API/CreateRepo',
        (_req, res, ctx) => {
          return res(
            ctx.status(400),
            ctx.json({
              code: 11,
              message: 'unable to create repo',
              details: [],
            }),
          );
        },
      ),
    );
    render(<CreateRepoModal />);

    const nameInput = await screen.findByRole('textbox', {name: /name/i});

    await type(nameInput, 'RepoError');
    await click(
      screen.getByRole('button', {
        name: /create/i,
      }),
    );

    expect(
      await screen.findByText('unable to create repo'),
    ).toBeInTheDocument();
  });

  describe('create repo validation', () => {
    const validInputs = [
      ['good'],
      ['good-'],
      ['good_'],
      ['good1'],
      ['_good'],
      ['a'.repeat(63)],
    ];
    const invalidInputs = [
      [
        'bad repo',
        'Repo name can only contain alphanumeric characters, underscores, and dashes',
      ],
      [
        'bad!',
        'Repo name can only contain alphanumeric characters, underscores, and dashes',
      ],
      [
        'bad.',
        'Repo name can only contain alphanumeric characters, underscores, and dashes',
      ],
      ['a'.repeat(64), 'Repo name exceeds maximum allowed length'],
    ];
    test.each(validInputs)(
      'should not error with a valid repo name (%j)',
      async (input) => {
        render(<CreateRepoModal />);

        const modal = screen.getByRole('dialog');
        const nameInput = await within(modal).findByLabelText('Name', {
          exact: false,
        });

        await type(nameInput, input);

        expect(within(modal).queryByRole('alert')).not.toBeInTheDocument();
      },
    );
    test.each(invalidInputs)(
      'should error with an invalid repo name (%j)',
      async (input, assertionText) => {
        render(<CreateRepoModal />);

        const modal = screen.getByRole('dialog');
        const nameInput = await within(modal).findByLabelText('Name', {
          exact: false,
        });

        await type(nameInput, input);

        expect(within(modal).getByText(assertionText)).toBeInTheDocument();
      },
    );
  });
});
