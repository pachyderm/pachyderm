import {mockUpdateProjectMutation} from '@graphqlTypes';
import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {withContextProviders, clear, click} from '@dash-frontend/testHelpers';

import UpdateProjectModalComponent from '../UpdateProjectModal';

const server = setupServer();

describe('UpdateProjectModal', () => {
  const UpdateProjectModal = withContextProviders(({onHide}) => {
    return (
      <UpdateProjectModalComponent
        show={true}
        onHide={onHide}
        projectName="ProjectA"
        description="project a description"
      />
    );
  });

  beforeAll(() => {
    server.listen();
  });

  afterAll(() => server.close());

  it('should display an error message if mutation fails', async () => {
    server.use(
      mockUpdateProjectMutation((_req, res, ctx) => {
        return res(
          ctx.errors([
            {
              message: 'unable to update project',
              path: ['updateProject'],
            },
          ]),
        );
      }),
    );
    render(<UpdateProjectModal />);

    const descriptionInput = await screen.findByRole('textbox', {
      name: /description/i,
    });
    await clear(descriptionInput);

    await click(
      screen.getByRole('button', {
        name: /confirm changes/i,
      }),
    );

    expect(
      await screen.findByText('unable to update project'),
    ).toBeInTheDocument();
  });
});
