import {mockCreateProjectMutation} from '@graphqlTypes';
import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {mockProjects} from '@dash-frontend/mocks';
import {withContextProviders, type, click} from '@dash-frontend/testHelpers';

import CreateProjectModalComponent from '../CreateProjectModal';

const server = setupServer();

describe('CreateProjectModal', () => {
  const CreateProjectModal = withContextProviders(({onHide}) => {
    return <CreateProjectModalComponent show={true} onHide={onHide} />;
  });

  beforeAll(() => {
    server.listen();
    server.use(mockProjects());
  });

  afterAll(() => server.close());

  it('should error if the project already exists', async () => {
    render(<CreateProjectModal />);

    const nameInput = await screen.findByLabelText('Name', {
      exact: false,
    });

    await type(nameInput, 'ProjectA');

    expect(
      await screen.findByText('Project name already in use'),
    ).toBeInTheDocument();
  });

  it('should display an error message if mutation fails', async () => {
    server.use(
      mockCreateProjectMutation((_req, res, ctx) => {
        return res(
          ctx.errors([
            {
              message: 'unable to create project',
              path: ['createProject'],
            },
          ]),
        );
      }),
    );
    render(<CreateProjectModal />);

    const nameInput = await screen.findByLabelText('Name', {
      exact: false,
    });

    await type(nameInput, 'ProjectError');
    await click(
      screen.getByRole('button', {
        name: /create/i,
      }),
    );

    expect(
      await screen.findByText('unable to create project'),
    ).toBeInTheDocument();
  });

  const validInputs = [
    ['goodproject'],
    ['good-project'],
    ['good_project'],
    ['goodproject1'],
    ['a'.repeat(51)],
  ];
  const invalidInputs = [
    [
      'bad project',
      'Name can only contain alphanumeric characters, underscores, and dashes',
    ],
    [
      'bad!',
      'Name can only contain alphanumeric characters, underscores, and dashes',
    ],
    [
      '_bad',
      'Name can only contain alphanumeric characters, underscores, and dashes',
    ],
    [
      'bad.',
      'Name can only contain alphanumeric characters, underscores, and dashes',
    ],
    ['a'.repeat(52), 'Project name exceeds maximum allowed length'],
  ];
  test.each(validInputs)(
    'should not error with a valid project name (%j)',
    async (input) => {
      render(<CreateProjectModal />);

      const nameInput = await screen.findByLabelText('Name', {
        exact: false,
      });

      await type(nameInput, input);

      expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    },
  );
  test.each(invalidInputs)(
    'should error with an invalid project name (%j)',
    async (input, assertionText) => {
      render(<CreateProjectModal />);

      const nameInput = await screen.findByLabelText('Name', {
        exact: false,
      });

      await type(nameInput, input);

      expect(screen.getByText(assertionText)).toBeInTheDocument();
    },
  );
});
