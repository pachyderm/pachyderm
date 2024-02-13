import {render, screen, within} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Permission} from '@dash-frontend/api/auth';
import {Empty} from '@dash-frontend/api/googleTypes';
import {
  InspectPipelineRequest,
  PipelineInfo,
  PipelineInfoPipelineType,
  PipelineState,
} from '@dash-frontend/api/pps';
import {
  mockPipelinesEmpty,
  mockPipelines,
  mockGetEdgesPipeline,
  mockRepos,
  mockTrueGetAuthorize,
  mockFalseGetAuthorize,
  buildPipeline,
  mockGetEnterpriseInfo,
  mockGetVersionInfo,
} from '@dash-frontend/mocks';
import {withContextProviders, click, hover} from '@dash-frontend/testHelpers';

import PipelineActionsMenuComponent from '../PipelineActionsMenu';

describe('Pipeline Actions Menu', () => {
  const server = setupServer();

  const PipelineActionsMenu = withContextProviders(() => (
    <PipelineActionsMenuComponent pipelineId="edges" />
  ));

  beforeAll(() => server.listen());

  beforeEach(() => {
    window.history.replaceState('', '', '/lineage/default/pipelines/edges');
    server.resetHandlers();
    server.use(mockGetVersionInfo());
    server.use(mockPipelinesEmpty());
    server.use(mockRepos());
    server.use(mockGetEdgesPipeline());
    server.use(mockGetEnterpriseInfo());
    server.use(
      mockTrueGetAuthorize([
        Permission.REPO_DELETE,
        Permission.REPO_WRITE,
        Permission.PROJECT_CREATE_REPO,
      ]),
    );
  });

  afterAll(() => server.close());

  it('should allow users to view a pipeline in the DAG', async () => {
    window.history.replaceState('', '', '/project/default/pipelines');
    render(<PipelineActionsMenu />);

    await click(screen.getByRole('button', {name: 'Pipeline Actions'}));
    await click(
      await screen.findByRole('menuitem', {
        name: /view in dag/i,
      }),
    );

    expect(window.location.pathname).toBe('/lineage/default/pipelines/edges');
  });

  it('should not allow users to view a pipeline in the DAG if already in the DAG', async () => {
    render(<PipelineActionsMenu />);

    await click(screen.getByRole('button', {name: 'Pipeline Actions'}));
    expect(
      screen.queryByRole('menuitem', {
        name: /view in dag/i,
      }),
    ).not.toBeInTheDocument();
  });

  it('should allow users to open the rerun pipeline modal', async () => {
    render(<PipelineActionsMenu />);

    await click(screen.getByRole('button', {name: 'Pipeline Actions'}));
    await click(
      await screen.findByRole('menuitem', {
        name: /rerun pipeline/i,
      }),
    );

    const modal = await screen.findByRole('dialog');
    expect(modal).toBeInTheDocument();

    expect(
      within(modal).getByRole('heading', {
        name: 'Rerun Pipeline: default/edges',
      }),
    ).toBeInTheDocument();
  });

  it('should not allow users to open the rerun pipeline modal for a spout pipeline', async () => {
    server.use(
      rest.post<InspectPipelineRequest, Empty, PipelineInfo>(
        '/api/pps_v2.API/InspectPipeline',
        async (_req, res, ctx) => {
          return res(
            ctx.json(
              buildPipeline({
                pipeline: {name: 'montage'},
                type: PipelineInfoPipelineType.PIPELINE_TYPE_SPOUT,
              }),
            ),
          );
        },
      ),
    );
    render(<PipelineActionsMenu />);

    await click(screen.getByRole('button', {name: 'Pipeline Actions'}));
    expect(
      await screen.findByRole('menuitem', {
        name: /rerun pipeline/i,
      }),
    ).toBeDisabled();
  });

  it('should allow users to open the delete pipeline modal', async () => {
    render(<PipelineActionsMenu />);

    await click(screen.getByRole('button', {name: 'Pipeline Actions'}));
    await click(
      await screen.findByRole('menuitem', {
        name: /delete pipeline/i,
      }),
    );

    expect(
      within(await screen.findByRole('dialog')).getByRole('heading', {
        name: 'Are you sure you want to delete this Pipeline?',
      }),
    ).toBeInTheDocument();
  });

  it('should disable the delete button when there are associated pipelines', async () => {
    server.use(mockPipelines());
    render(<PipelineActionsMenu />);

    await click(screen.getByRole('button', {name: 'Pipeline Actions'}));

    const deleteButton = await screen.findByRole('menuitem', {
      name: /delete pipeline/i,
    });
    expect(deleteButton).toBeDisabled();
    await hover(deleteButton);
    expect(
      screen.getByText(
        "This pipeline can't be deleted while it has downstream pipelines.",
      ),
    ).toBeInTheDocument();
  });

  describe('permissions', () => {
    it('should disable pipeline actions without proper permissions', async () => {
      server.use(mockFalseGetAuthorize());
      render(<PipelineActionsMenu />);

      await click(screen.getByRole('button', {name: 'Pipeline Actions'}));

      const deleteButton = await screen.findByRole('menuitem', {
        name: /delete pipeline/i,
      });
      expect(deleteButton).toBeDisabled();
      await hover(deleteButton);
      expect(
        screen.getByText('You need at least repoOwner to delete this.'),
      ).toBeInTheDocument();

      const rerunButton = await screen.findByRole('menuitem', {
        name: /rerun pipeline/i,
      });
      expect(rerunButton).toBeDisabled();
      await hover(rerunButton);
      expect(
        screen.getByText(
          'You need at least repoWriter to rerun this pipeline.',
        ),
      ).toBeInTheDocument();

      const updateButton = await screen.findByRole('menuitem', {
        name: /update pipeline/i,
      });
      expect(updateButton).toBeDisabled();
      await hover(updateButton);
      expect(
        screen.getByText(
          'You need at least repoWriter to update this pipeline.',
        ),
      ).toBeInTheDocument();

      const duplicateButton = await screen.findByRole('menuitem', {
        name: /duplicate pipeline/i,
      });
      expect(duplicateButton).toBeDisabled();
      await hover(duplicateButton);
      expect(
        screen.getByText(
          'You need at least projectWriter to create pipelines.',
        ),
      ).toBeInTheDocument();

      const pauseButton = await screen.findByRole('menuitem', {
        name: /pause pipeline/i,
      });
      expect(pauseButton).toBeDisabled();
      await hover(pauseButton);
      expect(
        screen.getByText(
          'You need at least repoWriter to pause this pipeline.',
        ),
      ).toBeInTheDocument();
    });

    it('should disable restarting a pipeline without proper permissions', async () => {
      server.use(mockFalseGetAuthorize());
      server.use(
        rest.post<InspectPipelineRequest, Empty, PipelineInfo>(
          '/api/pps_v2.API/InspectPipeline',
          async (_req, res, ctx) => {
            return res(
              ctx.json(
                buildPipeline({
                  pipeline: {name: 'edges'},
                  state: PipelineState.PIPELINE_PAUSED,
                }),
              ),
            );
          },
        ),
      );
      render(<PipelineActionsMenu />);

      await click(screen.getByRole('button', {name: 'Pipeline Actions'}));

      const restartButton = await screen.findByRole('menuitem', {
        name: /restart pipeline/i,
      });
      expect(restartButton).toBeDisabled();
      await hover(restartButton);
      expect(
        screen.getByText(
          'You need at least repoWriter to restart this pipeline.',
        ),
      ).toBeInTheDocument();
    });

    it('should disable running a cron pipeline without proper permissions', async () => {
      server.use(mockFalseGetAuthorize());
      server.use(
        rest.post<InspectPipelineRequest, Empty, PipelineInfo>(
          '/api/pps_v2.API/InspectPipeline',
          async (_req, res, ctx) => {
            return res(
              ctx.json(
                buildPipeline({
                  pipeline: {name: 'edges'},
                  details: {
                    input: {
                      cron: {
                        name: 'tick',
                        spec: '@every 10h',
                      },
                    },
                  },
                }),
              ),
            );
          },
        ),
      );
      render(<PipelineActionsMenu />);

      await click(screen.getByRole('button', {name: 'Pipeline Actions'}));

      const runCronButton = await screen.findByRole('menuitem', {
        name: /run cron/i,
      });
      expect(runCronButton).toBeDisabled();
      await hover(runCronButton);
      expect(
        screen.getByText(
          'You need at least repoWriter to trigger this cron pipeline',
        ),
      ).toBeInTheDocument();
    });
  });
});
