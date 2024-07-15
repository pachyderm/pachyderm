import {render, screen} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {DeleteRepoRequest} from '@dash-frontend/api/pfs';
import {RequestError} from '@dash-frontend/api/utils/error';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import DeleteRepoModalComponent from '../DeleteRepoModal';

const server = setupServer();

describe('DeleteRepoModal', () => {
  const DeleteRepoModal = withContextProviders(() => {
    return (
      <DeleteRepoModalComponent repoId="images" setModalOpen={() => null} />
    );
  });

  beforeAll(() => {
    server.listen();
    window.history.replaceState('', '', '/lineage/default/repos/images');
  });

  afterAll(() => server.close());

  it('should display an error message if mutation fails', async () => {
    server.use(
      rest.post<DeleteRepoRequest, Empty, RequestError>(
        '/api/pfs_v2.API/DeleteRepo',
        (_req, res, ctx) => {
          return res(
            ctx.status(400),
            ctx.json({
              code: 11,
              message: 'unable to delete repo',
              details: [],
            }),
          );
        },
      ),
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
