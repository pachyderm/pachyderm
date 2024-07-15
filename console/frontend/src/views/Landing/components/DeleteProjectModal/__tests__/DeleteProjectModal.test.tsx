import {render, screen} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {withContextProviders, type, click} from '@dash-frontend/testHelpers';

import DeleteProjectModalComponent from '../DeleteProjectModal';

const server = setupServer();

describe('DeleteProjectModal', () => {
  const DeleteProjectModal = withContextProviders(({onHide}) => {
    return (
      <DeleteProjectModalComponent
        show={true}
        onHide={onHide}
        projectName="ProjectA"
      />
    );
  });

  beforeAll(() => {
    server.listen();
  });

  afterAll(() => server.close());

  it('should display an error message if mutation fails', async () => {
    server.use(
      rest.post('/api/pps_v2.API/DeletePipelines', (_req, res, ctx) => {
        return res(
          ctx.status(400),
          ctx.json({
            message: 'unable to delete project',
          }),
        );
      }),
    );
    render(<DeleteProjectModal />);

    const confirmText = await screen.findByRole('textbox');
    await type(confirmText, 'ProjectA');

    await click(
      screen.getByRole('button', {
        name: /delete project/i,
      }),
    );

    expect(
      await screen.findByText('unable to delete project'),
    ).toBeInTheDocument();
  });
});
