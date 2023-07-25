import {
  CreateProjectArgs,
  CreateProjectMutation,
  ProjectStatus,
  mockDeleteProjectAndResourcesMutation,
  mockCreateProjectMutation,
} from '@graphqlTypes';
import {render, waitFor, within, screen} from '@testing-library/react';
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

import LandingComponent from '../Landing';

describe('Landing', () => {
  const server = setupServer();

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

  it('should display all project names and status', async () => {
    render(<Landing />);

    expect(
      await screen.findByRole('heading', {
        name: 'ProjectA',
        level: 5,
      }),
    ).toBeInTheDocument();

    expect(
      screen.getByRole('heading', {
        name: 'ProjectB',
        level: 5,
      }),
    ).toBeInTheDocument();

    expect(
      screen.getByRole('heading', {
        name: 'ProjectC',
        level: 5,
      }),
    ).toBeInTheDocument();

    expect(screen.getAllByRole('row', {})).toHaveLength(3);

    expect(
      screen.getByRole('tab', {
        name: /projects/i,
      }),
    ).toBeInTheDocument();

    expect(screen.getAllByTestId('ProjectStatus__HEALTHY')).toHaveLength(2);
    expect(screen.getAllByTestId('ProjectStatus__UNHEALTHY')).toHaveLength(1);

    expect(screen.getAllByText(/A description for project/)).toHaveLength(3);
  });

  it('should allow users to non-case sensitive search for projects by name', async () => {
    render(<Landing />);

    expect(
      await screen.findByRole('heading', {
        name: 'ProjectA',
        level: 5,
      }),
    ).toBeInTheDocument();
    expect(
      await screen.findByRole('heading', {
        name: 'ProjectB',
        level: 5,
      }),
    ).toBeInTheDocument();

    const searchBox = await screen.findByRole('searchbox');

    await type(searchBox, 'projecta');

    await screen.findByRole('heading', {
      name: 'ProjectA',
      level: 5,
    });
    expect(
      screen.queryByRole('heading', {
        name: 'ProjectB',
        level: 5,
      }),
    ).not.toBeInTheDocument();
  });

  it('should allow a user to view a project based in default lineage view', async () => {
    render(<Landing />);

    expect(window.location.pathname).not.toBe('/lineage/ProjectA');

    const viewProjectButton = await screen.findByRole('button', {
      name: /View project ProjectA/i,
    });
    await click(viewProjectButton);
    expect(window.location.pathname).toBe('/lineage/ProjectA');
  });

  it('should allow the user to sort by name', async () => {
    render(<Landing />);

    expect(
      await screen.findByRole('heading', {
        name: 'ProjectA',
        level: 5,
      }),
    ).toBeInTheDocument();

    const projectsPanel = screen.getByRole('tabpanel', {
      name: /projects 3/i,
    });
    const projectNamesAZ = within(projectsPanel).getAllByRole('heading', {
      level: 5,
    });

    // Starts in A-Z order
    expect(
      projectNamesAZ.map((projectName) => projectName.textContent),
    ).toEqual(['ProjectA', 'ProjectB', 'ProjectC']);

    const sortDropdown = await screen.findByRole('button', {
      name: 'Sort by: Name A-Z',
    });
    await click(sortDropdown);
    const nameSort = await screen.findByRole('menuitem', {name: 'Name Z-A'});
    await click(nameSort);

    const projectNamesZA = within(projectsPanel).getAllByRole('heading', {
      level: 5,
    });
    expect(
      projectNamesZA.map((projectName) => projectName.textContent),
    ).toEqual(['ProjectC', 'ProjectB', 'ProjectA']);
  });

  it('should allow the user to filter projects by status', async () => {
    render(<Landing />);
    const projects = await screen.findByTestId('Landing__view');
    expect(
      await within(projects).findAllByTestId('ProjectStatus__HEALTHY'),
    ).toHaveLength(2);
    expect(
      await within(projects).findAllByTestId('ProjectStatus__UNHEALTHY'),
    ).toHaveLength(1);

    const filterDropdown = await screen.findByRole('button', {
      name: 'Show: All',
    });
    await click(filterDropdown);
    const healthyButton = await screen.findByLabelText('Healthy');
    await click(healthyButton);

    expect(
      await screen.findByRole('button', {name: 'Show: Unhealthy'}),
    ).toBeInTheDocument();

    expect(
      await within(projects).findAllByTestId('ProjectStatus__UNHEALTHY'),
    ).toHaveLength(1);
    expect(
      within(projects).queryByTestId('ProjectStatus__HEALTHY'),
    ).not.toBeInTheDocument();

    const unHealthyButton = await screen.findByLabelText('Unhealthy');
    await click(unHealthyButton);

    expect(
      await screen.findByRole('button', {name: 'Show: None'}),
    ).toBeInTheDocument();

    expect(
      within(projects).queryByTestId('ProjectStatus__HEALTHY'),
    ).not.toBeInTheDocument();
    expect(
      within(projects).queryByTestId('ProjectStatus__UNHEALTHY'),
    ).not.toBeInTheDocument();

    await click(healthyButton);

    expect(
      await screen.findByRole('button', {name: 'Show: Healthy'}),
    ).toBeInTheDocument();

    expect(
      await within(projects).findAllByTestId('ProjectStatus__HEALTHY'),
    ).toHaveLength(2);
    expect(
      within(projects).queryByTestId('ProjectStatus__UNHEALTHY'),
    ).not.toBeInTheDocument();
  });

  it('should give users a pachctl command to set their active project', async () => {
    render(<Landing />);

    expect(await screen.findByText('ProjectA')).toBeInTheDocument();

    await click(
      screen.getByRole('button', {
        name: 'ProjectA overflow menu',
      }),
    );
    await click(
      await screen.findByRole('menuitem', {
        name: /set active project/i,
      }),
    );

    expect(
      await screen.findByText('Set Active Project: "ProjectA"'),
    ).toBeInTheDocument();
    await click(screen.getByRole('button', {name: 'Copy'}));

    expect(window.document.execCommand).toHaveBeenCalledWith('copy');
  });

  describe('Create Project Modal', () => {
    it('should create a project', async () => {
      server.use(
        mockCreateProjectMutation((req, res, ctx) => {
          const {args} = req.variables;
          return res(
            ctx.data({
              createProject: {
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

      const createButton = await screen.findByText('Create Project');
      await click(createButton);

      const modal = screen.getByRole('dialog');
      const nameInput = await within(modal).findByLabelText('Name', {
        exact: false,
      });
      const descriptionInput = await within(modal).findByLabelText(
        'Description (optional)',
        {
          exact: false,
        },
      );

      await type(nameInput, 'New-Project');
      await type(descriptionInput, 'description text');

      expect(screen.queryByText('New-Project')).not.toBeInTheDocument();

      await click(within(modal).getByText('Create'));
      expect(
        screen.getByRole('heading', {
          name: /new-project/i,
        }),
      ).toBeInTheDocument();
    });

    it('should error if the project already exists', async () => {
      render(<Landing />);

      const createButton = await screen.findByText('Create Project');
      await click(createButton);

      const modal = screen.getByRole('dialog');
      const nameInput = await within(modal).findByLabelText('Name', {
        exact: false,
      });

      await type(nameInput, 'ProjectA');

      expect(
        await within(modal).findByText('Project name already in use'),
      ).toBeInTheDocument();
    });
  });

  describe('Delete Project Modal', () => {
    it('should delete a project', async () => {
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

      render(<Landing />);
      // wait for page to populate
      expect(await screen.findByText('ProjectA')).toBeInTheDocument();
      // open the modal on the correct row
      expect(screen.queryByRole('dialog')).toBeNull();
      expect(
        screen.queryByRole('menuitem', {
          name: /delete project/i,
        }),
      ).toBeNull();
      await click(
        screen.getByRole('button', {
          name: 'ProjectA overflow menu', // should be case-sensitive since projects are not case sensitive
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
  });
});
