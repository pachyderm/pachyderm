import {render, screen} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  DeletePipelineRequest,
  DeletePipelinesResponse,
} from '@dash-frontend/api/pps';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import DeletePipelineModalComponent from '../DeletePipelineModal';

const server = setupServer();

describe('DeletePipelineModal', () => {
  const DeletePipelineModal = withContextProviders(() => {
    return (
      <DeletePipelineModalComponent
        setModalOpen={() => null}
        pipelineId="edges"
      />
    );
  });

  beforeAll(() => {
    server.listen();
    window.history.replaceState('', '', '/lineage/default/pipelines/edges');
  });

  afterAll(() => server.close());

  it('should display an error message if mutation fails', async () => {
    server.use(
      rest.post<DeletePipelineRequest, Empty, {message: string}>(
        '/api/pps_v2.API/DeletePipeline',
        (_req, res, ctx) => {
          return res(
            ctx.status(500),
            ctx.json({
              message:
                'delete branch tutorial/edges@master: branch master has [tutorial/montage@master tutorial/montage.meta@master] as subvenance, deleting it would break those branches',
            }),
          );
        },
      ),
    );
    render(<DeletePipelineModal />);

    await click(
      screen.getByRole('button', {
        name: /delete/i,
      }),
    );

    expect(
      await screen.findByText(
        'delete branch tutorial/edges@master: branch master has [tutorial/montage@master tutorial/montage.meta@master] as subvenance, deleting it would break those branches',
      ),
    ).toBeInTheDocument();
  });

  it('should clear a global id filter if one is present', async () => {
    window.history.replaceState(
      '',
      '',
      '/lineage/default/pipelines/edges?globalId=1234',
    );
    server.use(
      rest.post<DeletePipelineRequest, Empty, DeletePipelinesResponse>(
        '/api/pps_v2.API/DeletePipeline',
        (_req, res, ctx) => {
          return res(ctx.json({pipelines: []}));
        },
      ),
    );
    render(<DeletePipelineModal />);

    await click(
      screen.getByRole('button', {
        name: /delete/i,
      }),
    );

    expect(window.location.search).toBe('');
  });
});
