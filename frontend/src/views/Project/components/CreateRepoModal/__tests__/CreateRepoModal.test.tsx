import {mockCreateRepoMutation} from '@graphqlTypes';
import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

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
      mockCreateRepoMutation((_req, res, ctx) => {
        return res(
          ctx.errors([
            {
              message: 'unable to create repo',
              path: ['createRepo'],
            },
          ]),
        );
      }),
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
});
