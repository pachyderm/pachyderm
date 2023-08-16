import {mockDeletePipelineMutation} from '@graphqlTypes';
import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {withContextProviders, click} from '@dash-frontend/testHelpers';

import DeletePipelineModalComponent from '../DeletePipelineModal';

const server = setupServer();

describe('DeletePipelineModal', () => {
  const DeletePipelineModal = withContextProviders(() => {
    return <DeletePipelineModalComponent setModalOpen={() => null} />;
  });

  beforeAll(() => {
    server.listen();
    window.history.replaceState('', '', '/lineage/default/pipelines/edges');
  });

  afterAll(() => server.close());

  it('should display an error message if mutation fails', async () => {
    server.use(
      mockDeletePipelineMutation((_req, res, ctx) => {
        return res(
          ctx.errors([
            {
              message: 'unable to delete pipeline',
              path: ['deletePipeline'],
            },
          ]),
        );
      }),
    );
    render(<DeletePipelineModal />);

    await click(
      screen.getByRole('button', {
        name: /delete/i,
      }),
    );

    expect(
      await screen.findByText('unable to delete pipeline'),
    ).toBeInTheDocument();
  });
});
