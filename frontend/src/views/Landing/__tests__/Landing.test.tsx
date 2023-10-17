import {
  ProjectStatus,
  mockCreateProjectMutation,
  mockUpdateProjectMutation,
  mockProjectStatusQuery,
} from '@graphqlTypes';
import {render, within, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockProjects,
  mockEmptyProjectDetails,
  mockEmptyGetAuthorize,
  mockFalseGetAuthorize,
  mockGetVersionInfo,
  mockTrueGetAuthorize,
} from '@dash-frontend/mocks';
import {
  withContextProviders,
  click,
  type,
  clear,
} from '@dash-frontend/testHelpers';

import {Landing as LandingComponent} from '../Landing';

describe('Landing', () => {
  const server = setupServer();

  const Landing = withContextProviders(() => {
    return <LandingComponent />;
  });

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    server.resetHandlers();
    server.use(mockEmptyGetAuthorize());
    server.use(mockGetVersionInfo());
    server.use(mockProjects());
    server.use(
      mockProjectStatusQuery((req, res, ctx) => {
        return res(
          ctx.data({
            projectStatus: {
              __typename: 'Project',
              id: req.variables.id,
              status:
                req.variables.id !== 'ProjectC'
                  ? ProjectStatus.HEALTHY
                  : ProjectStatus.UNHEALTHY,
            },
          }),
        );
      }),
    );
    server.use(mockEmptyProjectDetails());
  });

  afterAll(() => server.close());

  describe('Cluster Config button', () => {
    it('appears in CE', async () => {
      render(<Landing />);

      expect(
        await screen.findByRole('button', {name: /cluster defaults/i}),
      ).toBeEnabled();
    });

    it('appears as a Cluster Admin', async () => {
      server.use(mockTrueGetAuthorize());
      render(<Landing />);

      expect(
        await screen.findByRole('button', {name: /cluster defaults/i}),
      ).toBeEnabled();
    });

    it('hides when not Cluster Admin', async () => {
      server.use(mockFalseGetAuthorize());
      render(<Landing />);

      expect(
        await screen.findByRole('heading', {
          name: 'ProjectA',
          level: 5,
        }),
      ).toBeInTheDocument();

      expect(
        screen.queryByRole('button', {name: /cluster defaults/i}),
      ).not.toBeInTheDocument();
    });
  });

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

  it('should allow the user to sort by created date newest', async () => {
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
      name: /Sort by/,
    });
    await click(sortDropdown);
    const nameSort = await screen.findByRole('menuitem', {name: 'Newest'});
    await click(nameSort);

    const projectNamesZA = within(projectsPanel).getAllByRole('heading', {
      level: 5,
    });
    expect(
      projectNamesZA.map((projectName) => projectName.textContent),
    ).toEqual(['ProjectA', 'ProjectB', 'ProjectC']);

    const creationDates = screen.getAllByTestId('ProjectRow__created');
    expect(creationDates[0]).toHaveTextContent('Sep 13, 2020; 12:26');
    expect(creationDates[1]).toHaveTextContent('Jul 14, 2017; 2:40');
    expect(creationDates[2]).toHaveTextContent('May 13, 2014; 16:53');
  });

  it('should allow the user to sort by created date oldest', async () => {
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
      name: /Sort by/,
    });
    await click(sortDropdown);
    const nameSort = await screen.findByRole('menuitem', {name: 'Oldest'});
    await click(nameSort);

    const projectNamesZA = within(projectsPanel).getAllByRole('heading', {
      level: 5,
    });
    expect(
      projectNamesZA.map((projectName) => projectName.textContent),
    ).toEqual(['ProjectC', 'ProjectB', 'ProjectA']);

    const creationDates = screen.getAllByTestId('ProjectRow__created');
    expect(creationDates[0]).toHaveTextContent('May 13, 2014; 16:53');
    expect(creationDates[1]).toHaveTextContent('Jul 14, 2017; 2:40');
    expect(creationDates[2]).toHaveTextContent('Sep 13, 2020; 12:26');
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

  it('should not allow a user to create a project without permission', async () => {
    server.use(mockFalseGetAuthorize());

    render(<Landing />);

    expect(
      await screen.findByRole('heading', {
        name: 'ProjectA',
        level: 5,
      }),
    ).toBeInTheDocument();

    expect(
      screen.queryByRole('button', {
        name: /create project/i,
      }),
    ).not.toBeInTheDocument();
  });

  describe('Create Project Modal', () => {
    beforeAll(() => {
      server.use(mockEmptyGetAuthorize());
    });

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
                createdAt: {seconds: 1600000000, nanos: 248610000},
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
        await screen.findByRole('heading', {
          name: /new-project/i,
        }),
      ).toBeInTheDocument();
    });
  });

  describe('UpdateProjectModal', () => {
    it('should update a project description with no text then with text', async () => {
      server.use(
        mockUpdateProjectMutation((req, res, ctx) => {
          const {args} = req.variables;
          return res(
            ctx.data({
              updateProject: {
                id: args.name,
                description: args.description,
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
      });

      expect(await within(row2).findByText('new desc')).toBeInTheDocument();
    });
  });
});
