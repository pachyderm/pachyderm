import {mockUpdateProjectMutation, ProjectStatus} from '@graphqlTypes';
import {render, screen, within} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockProjects,
  mockEmptyProjectDetails,
  mockGetVersionInfo,
  mockEmptyGetAuthorize,
} from '@dash-frontend/mocks';
import {
  withContextProviders,
  click,
  type,
  clear,
} from '@dash-frontend/testHelpers';

import LandingComponent from '../../../Landing';

const server = setupServer();

describe('UpdateProjectModal', () => {
  const Landing = withContextProviders(() => {
    return <LandingComponent />;
  });

  beforeAll(() => {
    server.listen();
    server.use(mockEmptyGetAuthorize());
    server.use(mockGetVersionInfo());
    server.use(mockProjects());
    server.use(mockEmptyProjectDetails());
  });

  afterAll(() => server.close());

  it.skip('should update a project description with no text then with text', async () => {
    server.use(
      mockUpdateProjectMutation((req, res, ctx) => {
        const {args} = req.variables;
        return res(
          ctx.data({
            updateProject: {
              id: args.name,
              description: args.description,
              status: ProjectStatus.HEALTHY,
              __typename: 'Project',
            },
          }),
        );
      }),
    );
    render(<Landing />);

    // wait for page to populate
    expect(await screen.findByText('ProjectA')).toBeInTheDocument();

    // open the modal on the correct row
    expect(screen.queryByRole('dialog')).toBeNull();
    expect(
      screen.queryByRole('menuitem', {
        name: /edit project info/i,
      }),
    ).toBeNull();
    await click(
      screen.getByRole('button', {
        name: /projecta overflow menu/i,
      }),
    );
    const menuItem = await screen.findByRole('menuitem', {
      name: /edit project info/i,
    });
    expect(menuItem).toBeVisible();
    await click(menuItem);

    const modal = await screen.findByRole('dialog');
    expect(modal).toBeInTheDocument();

    const descriptionInput = await within(modal).findByRole('textbox', {
      name: /description/i,
      exact: false,
    });
    expect(descriptionInput).toHaveValue('A description for project a'); // Element should be prepopulated with the existing description.

    await clear(descriptionInput);

    await click(
      within(modal).getByRole('button', {
        name: /confirm changes/i,
      }),
    );
    const row = screen.getByRole('row', {
      name: /projecta/i,
      exact: false,
    });
    expect(await within(row).findByText('N/A')).toBeInTheDocument();

    // reopen it to edit and save some data
    expect(
      screen.queryByRole('menuitem', {
        name: /edit project info/i,
      }),
    ).toBeNull();
    await click(
      screen.getByRole('button', {
        name: /projecta overflow menu/i,
      }),
    );
    const menuItem2 = await screen.findByRole('menuitem', {
      name: /edit project info/i,
    });
    expect(menuItem2).toBeVisible();
    await click(menuItem2);

    const modal2 = await screen.findByRole('dialog');
    expect(modal2).toBeInTheDocument();

    const descriptionInput2 = await within(modal2).findByRole('textbox', {
      name: /description/i,
      exact: false,
    });
    expect(descriptionInput2).toHaveValue('');

    await clear(descriptionInput2);
    await type(descriptionInput2, 'new desc');

    await click(
      within(modal2).getByRole('button', {
        name: /confirm changes/i,
      }),
    );

    const row2 = await screen.findByRole('row', {
      name: /projecta/i,
      exact: false,
    });

    expect(await within(row2).findByText('new desc')).toBeInTheDocument();
  });
});
