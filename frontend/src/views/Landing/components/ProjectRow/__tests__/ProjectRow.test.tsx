import {
  ProjectStatus,
  mockGetAuthorizeQuery,
  Permission,
  mockDeleteProjectAndResourcesMutation,
} from '@graphqlTypes';
import {render, waitFor, within, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockFalseGetAuthorize,
  mockEmptyGetRoles,
  mockHealthyProjectStatus,
} from '@dash-frontend/mocks';
import {
  withContextProviders,
  click,
  type,
  clear,
} from '@dash-frontend/testHelpers';

import ProjectRowComponent from '../ProjectRow';

describe('ProjectRow RBAC', () => {
  const server = setupServer();

  const ProjectRow = withContextProviders(() => {
    return (
      <ProjectRowComponent
        project={{
          id: 'ProjectA',
          description: 'A description for project a',
          status: ProjectStatus.HEALTHY,
        }}
        isSelected={true}
        multiProject={true}
        setSelectedProject={() => null}
      />
    );
  });

  beforeAll(() => {
    server.use(mockHealthyProjectStatus());
    server.listen();
  });

  afterAll(() => server.close());

  it('should not allow CRUD actions without permission', async () => {
    server.use(mockFalseGetAuthorize());

    render(<ProjectRow />);

    await click(
      screen.getByRole('button', {
        name: 'ProjectA overflow menu',
      }),
    );
    expect(
      await screen.findByRole('menuitem', {
        name: /view project roles/i,
      }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole('menuitem', {
        name: /delete project/i,
      }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole('menuitem', {
        name: /edit project info/i,
      }),
    ).not.toBeInTheDocument();
  });

  it('should allow a user to delete a project with permission', async () => {
    server.use(
      mockGetAuthorizeQuery((req, res, ctx) => {
        const deleteAuth = req.variables.args.permissionsList.includes(
          Permission.PROJECT_DELETE,
        );
        return res(
          ctx.data({
            getAuthorize: {
              satisfiedList: [],
              missingList: [],
              authorized: deleteAuth ? true : false,
              principal: 'email@user.com',
            },
          }),
        );
      }),
    );
    server.use(
      mockDeleteProjectAndResourcesMutation((req, res, ctx) => {
        if (req.variables.args.name === 'ProjectA') {
          return res(
            ctx.data({
              deleteProjectAndResources: true,
            }),
          );
        } else return res(ctx.errors([]));
      }),
    );

    render(<ProjectRow />);

    // wait for page to populate
    expect(await screen.findByText('ProjectA')).toBeInTheDocument();
    expect(screen.queryByRole('dialog')).toBeNull();
    expect(
      screen.queryByRole('menuitem', {
        name: /delete project/i,
      }),
    ).toBeNull();
    await click(
      screen.getByRole('button', {
        name: 'ProjectA overflow menu',
      }),
    );
    const menuItem = await screen.findByRole('menuitem', {
      name: /delete project/i,
    });
    expect(menuItem).toBeVisible();
    await click(menuItem);

    const modal = await screen.findByRole('dialog');
    expect(modal).toBeInTheDocument();
    const projectNameInput = await within(modal).findByRole('textbox');
    expect(projectNameInput).toHaveValue('');

    await clear(projectNameInput);

    const confirmButton = within(modal).getByRole('button', {
      name: /delete project/i,
    });

    expect(confirmButton).toBeDisabled();
    await type(projectNameInput, 'Project');
    expect(confirmButton).toBeDisabled();
    await type(projectNameInput, 'A');
    expect(confirmButton).toBeEnabled();

    await click(confirmButton);
    await waitFor(() =>
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument(),
    );
  });

  it('should allow a user to update a project with permission', async () => {
    server.use(
      mockGetAuthorizeQuery((req, res, ctx) => {
        const createAuth = req.variables.args.permissionsList.includes(
          Permission.PROJECT_CREATE,
        );
        return res(
          ctx.data({
            getAuthorize: {
              satisfiedList: [],
              missingList: [],
              authorized: createAuth ? true : false,
              principal: 'email@user.com',
            },
          }),
        );
      }),
    );

    render(<ProjectRow />);

    await click(
      screen.getByRole('button', {
        name: 'ProjectA overflow menu',
      }),
    );
    await click(
      await screen.findByRole('menuitem', {
        name: /edit project info/i,
      }),
    );

    const modal = await screen.findByRole('dialog');
    expect(modal).toBeInTheDocument();

    expect(
      within(modal).getByRole('heading', {name: 'Edit Project: ProjectA'}),
    ).toBeInTheDocument();
  });

  it('should allow a user to update roles with permission', async () => {
    server.use(mockEmptyGetRoles());
    server.use(
      mockGetAuthorizeQuery((req, res, ctx) => {
        const rolesAuth = req.variables.args.permissionsList.includes(
          Permission.PROJECT_MODIFY_BINDINGS,
        );
        return res(
          ctx.data({
            getAuthorize: {
              satisfiedList: [],
              missingList: [],
              authorized: rolesAuth ? true : false,
              principal: 'email@user.com',
            },
          }),
        );
      }),
    );

    render(<ProjectRow />);

    await click(
      screen.getByRole('button', {
        name: 'ProjectA overflow menu',
      }),
    );
    await click(
      await screen.findByRole('menuitem', {
        name: /edit project roles/i,
      }),
    );

    const modal = await screen.findByRole('dialog');
    expect(modal).toBeInTheDocument();

    expect(
      within(modal).getByRole('heading', {
        name: 'Set Project Level Roles: ProjectA',
      }),
    ).toBeInTheDocument();
  });
});
