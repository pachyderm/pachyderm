import {
  render,
  waitFor,
  within,
  screen,
  waitForElementToBeRemoved,
} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Permission} from '@dash-frontend/api/auth';
import {Empty} from '@dash-frontend/api/googleTypes';
import {
  DeleteProjectRequest,
  DeleteReposRequest,
  DeleteReposResponse,
  ReposSummaryRequest,
  ReposSummaryResponse,
} from '@dash-frontend/api/pfs';
import {
  DeletePipelinesRequest,
  DeletePipelinesResponse,
  PipelinesSummaryRequest,
  PipelinesSummaryResponse,
} from '@dash-frontend/api/pps';
import {
  mockFalseGetAuthorize,
  mockEmptyGetRoles,
  mockGetEnterpriseInfoInactive,
  mockTrueGetAuthorize,
  mockGetVersionInfo,
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
          project: {name: 'ProjectA'},
          description: 'A description for project a',
        }}
        isSelected={true}
        multiProject={true}
        setSelectedProject={() => null}
        setMyProjectsCount={() => null}
        setAllProjectsCount={() => null}
        allProjectNames={['ProjectA']}
      />
    );
  });

  beforeAll(() => {
    server.use(
      rest.post<ReposSummaryRequest, Empty, ReposSummaryResponse>(
        '/api/pfs_v2.API/ReposSummary',
        (_req, res, ctx) => {
          return res(
            ctx.json({
              summaries: [
                {
                  project: {name: 'default'},
                  userRepoCount: '0',
                  sizeBytes: '0',
                },
              ],
            }),
          );
        },
      ),
    );
    server.use(
      rest.post<PipelinesSummaryRequest, Empty, PipelinesSummaryResponse>(
        '/api/pps_v2.API/PipelinesSummary',
        (_req, res, ctx) => {
          return res(
            ctx.json({
              summaries: [
                {
                  project: {name: 'ProjectA'},
                  activePipelines: '3',
                  pausedPipelines: '0',
                  failedPipelines: '0',
                  unhealthyPipelines: '0',
                },
              ],
            }),
          );
        },
      ),
    );
    server.listen();
    server.use(mockGetVersionInfo());
    server.use(mockGetEnterpriseInfoInactive());
    server.use(mockFalseGetAuthorize());
  });

  afterAll(() => server.close());

  it('should route users to the DAG page', async () => {
    render(<ProjectRow />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));
    await click(
      screen.getByRole('button', {
        name: /view project projecta/i,
      }),
    );
    expect(window.location.pathname).toBe('/lineage/ProjectA');
  });

  it('should route users to the Repos list', async () => {
    render(<ProjectRow />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    await click(
      screen.getByRole('button', {
        name: /view projecta repos/i,
      }),
    );
    expect(window.location.pathname).toBe('/project/ProjectA/repos');
  });

  it('should route users to the Pipelines list', async () => {
    render(<ProjectRow />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    await click(
      screen.getByRole('button', {
        name: /view projecta pipelines/i,
      }),
    );
    expect(window.location.pathname).toBe('/project/ProjectA/pipelines');
  });

  it('should not allow CRUD actions without permission', async () => {
    server.use(mockFalseGetAuthorize());

    render(<ProjectRow />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

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
    expect(
      screen.queryByRole('menuitem', {
        name: /edit project defaults/i,
      }),
    ).not.toBeInTheDocument();
  });

  it('should allow a user to delete a project with permission', async () => {
    server.use(mockTrueGetAuthorize([Permission.PROJECT_DELETE]));
    server.use(
      rest.post<DeletePipelinesRequest, Empty, DeletePipelinesResponse>(
        '/api/pps_v2.API/DeletePipelines',
        (req, res, ctx) => {
          if (req.body.projects?.[0].name === 'ProjectA') {
            return res(
              ctx.json({
                pipelines: [{name: 'Pipeline 1'}],
              }),
            );
          }

          return res(ctx.status(400), ctx.json({}));
        },
      ),
    );
    server.use(
      rest.post<DeleteReposRequest, Empty, DeleteReposResponse>(
        '/api/pfs_v2.API/DeleteRepos',
        (req, res, ctx) => {
          if (req.body.projects?.[0].name === 'ProjectA') {
            return res(
              ctx.json({
                repos: [{name: 'Repo 1'}],
              }),
            );
          }

          return res(ctx.status(400), ctx.json({}));
        },
      ),
    );
    server.use(
      rest.post<DeleteProjectRequest, Empty, Empty>(
        '/api/pfs_v2.API/DeleteProject',
        (req, res, ctx) => {
          if (req.body.project?.name === 'ProjectA') {
            return res(ctx.json({}));
          }

          return res(ctx.status(400), ctx.json({}));
        },
      ),
    );

    render(<ProjectRow />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

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
    server.use(mockTrueGetAuthorize([Permission.PROJECT_CREATE]));

    render(<ProjectRow />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

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
    server.use(mockTrueGetAuthorize([Permission.PROJECT_MODIFY_BINDINGS]));

    render(<ProjectRow />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

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

  it('should route users to project defaults with permission', async () => {
    server.use(mockTrueGetAuthorize([Permission.PROJECT_SET_DEFAULTS]));

    render(<ProjectRow />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    await click(
      screen.getByRole('button', {
        name: 'ProjectA overflow menu',
      }),
    );
    await click(
      await screen.findByRole('menuitem', {
        name: /edit project defaults/i,
      }),
    );

    expect(window.location.pathname).toBe('/project/ProjectA/defaults');
  });
});
