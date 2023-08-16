import {mockDeleteRepoMutation} from '@graphqlTypes';
import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {withContextProviders, click} from '@dash-frontend/testHelpers';

import DeleteRepoModalComponent from '../DeleteRepoModal';

const server = setupServer();

describe('DeleteRepoModal', () => {
  const DeleteRepoModal = withContextProviders(() => {
    return <DeleteRepoModalComponent setModalOpen={() => null} />;
  });

  beforeAll(() => {
    server.listen();
    window.history.replaceState('', '', '/lineage/default/repos/images');
  });

  afterAll(() => server.close());

  it('should display an error message if mutation fails', async () => {
    server.use(
      mockDeleteRepoMutation((_req, res, ctx) => {
        return res(
          ctx.errors([
            {
              message: 'unable to delete repo',
              path: ['deleteRepo'],
            },
          ]),
        );
      }),
    );
    render(<DeleteRepoModal />);

    await click(
      screen.getByRole('button', {
        name: /delete/i,
      }),
    );

    expect(
      await screen.findByText('unable to delete repo'),
    ).toBeInTheDocument();
  });
});
