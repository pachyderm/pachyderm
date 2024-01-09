import {render, within, screen, waitFor} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Permission} from '@dash-frontend/api/auth';
import {Empty} from '@dash-frontend/api/googleTypes';
import {
  ListPipelineRequest,
  PipelineInfo,
  PipelineState,
} from '@dash-frontend/api/pps';
import {
  mockProjects,
  mockEmptyGetAuthorize,
  mockFalseGetAuthorize,
  mockGetVersionInfo,
  mockTrueGetAuthorize,
  buildPipeline,
  mockGetEnterpriseInfo,
} from '@dash-frontend/mocks';
import {withContextProviders, click, type} from '@dash-frontend/testHelpers';

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
    server.use(mockGetEnterpriseInfo());
    server.use(mockGetVersionInfo());
    server.use(mockProjects());
    server.use(
      rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
        '/api/pps_v2.API/ListPipeline',
        (req, res, ctx) => {
          if (req.body.projects?.[0].name === 'ProjectC') {
            return res(
              ctx.json([
                buildPipeline({
                  state: PipelineState.PIPELINE_FAILURE,
                }),
              ]),
            );
          }

          return res(ctx.json([buildPipeline()]));
        },
      ),
    );
    server.use(
      rest.post('/api/pfs_v2.API/ListRepo', (_req, res, ctx) => {
        return res(ctx.json([]));
      }),
    );
    server.use(
      rest.post('/api/pps_v2.API/ListJob', (_req, res, ctx) => {
        return res(ctx.json([]));
      }),
    );
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
      server.use(
        mockTrueGetAuthorize([
          Permission.PROJECT_CREATE,
          Permission.CLUSTER_AUTH_SET_CONFIG,
        ]),
      );
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

    expect(await screen.findAllByTestId('ProjectStatus__HEALTHY')).toHaveLength(
      2,
    );
    expect(
      await screen.findAllByTestId('ProjectStatus__UNHEALTHY'),
    ).toHaveLength(1);

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

    // Have to wait for everything to load first
    await waitFor(async () => {
      expect(screen.getAllByTestId('ProjectStatus__HEALTHY')).toHaveLength(2);
    });

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
});
