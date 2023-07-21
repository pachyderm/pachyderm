import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockProjects,
  mockGetVersionInfo,
  mockEmptyGetAuthorize,
} from '@dash-frontend/mocks';
import {withContextProviders, type} from '@dash-frontend/testHelpers';

import CreateProjectModalComponent from '../CreateProjectModal';

const server = setupServer();

describe('CreateProjectModal', () => {
  const CreateProjectModal = withContextProviders(({onHide}) => {
    return <CreateProjectModalComponent show={true} onHide={onHide} />;
  });

  beforeAll(() => {
    server.listen();
    server.use(mockEmptyGetAuthorize());
    server.use(mockGetVersionInfo());
    server.use(mockProjects());
  });

  afterAll(() => server.close());

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
