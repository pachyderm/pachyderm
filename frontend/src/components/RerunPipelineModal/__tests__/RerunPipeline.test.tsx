import {mockRerunPipelineMutation} from '@graphqlTypes';
import {render, screen, waitFor} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockFalseGetAuthorize,
  mockGetSpoutPipeline,
  mockTrueGetAuthorize,
} from '@dash-frontend/mocks';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import RerunPipelineButtonComponent from '../components/RerunPipelineButton';
import RerunPipelineModalComponent from '../RerunPipelineModal';

const server = setupServer();

describe('RerunPipelineModal', () => {
  const RerunPipelineModal = withContextProviders(({pipelineId}) => {
    return (
      <RerunPipelineModalComponent
        show={true}
        onHide={() => null}
        pipelineId={pipelineId}
      />
    );
  });

  const RerunPipelineButton = withContextProviders(() => {
    return <RerunPipelineButtonComponent />;
  });

  beforeAll(() => {
    server.listen();
    window.history.replaceState('', '', '/lineage/default/pipelines/edges');
  });

  afterAll(() => server.close());

  it('should display an error message if mutation fails', async () => {
    server.use(
      mockRerunPipelineMutation((_req, res, ctx) => {
        return res(
          ctx.errors([
            {
              message: 'unable to rerun pipeline',
              path: ['rerunPipeline'],
            },
          ]),
        );
      }),
    );
    render(<RerunPipelineModal pipelineId="edges" />);

    await click(
      screen.getByRole('radio', {
        name: /process all datums/i,
      }),
    );

    await click(
      screen.getByRole('button', {
        name: /rerun pipeline/i,
      }),
    );

    expect(
      await screen.findByText('unable to rerun pipeline'),
    ).toBeInTheDocument();
  });

  it('should process all datums', async () => {
    server.use(
      mockRerunPipelineMutation((req, res, ctx) => {
        const {projectId, pipelineId, reprocess} = req.variables.args;

        if (
          projectId === 'default' &&
          pipelineId === 'edges' &&
          reprocess === true
        ) {
          return res(ctx.data({rerunPipeline: true}));
        }

        return res(
          ctx.errors([
            {
              message: 'unable to rerun pipeline',
              path: ['rerunPipeline'],
            },
          ]),
        );
      }),
    );
    render(<RerunPipelineModal pipelineId="edges" />);

    const submit = screen.getByRole('button', {
      name: /rerun pipeline/i,
    });

    expect(submit).toBeDisabled();

    await click(
      screen.getByRole('radio', {
        name: /process all datums/i,
      }),
    );

    await click(submit);
    expect(
      await screen.findByText('Pipeline was rerun successfully'),
    ).toBeInTheDocument();
  });

  it('should process only failed datums', async () => {
    server.use(
      mockRerunPipelineMutation((req, res, ctx) => {
        const {projectId, pipelineId, reprocess} = req.variables.args;

        if (
          projectId === 'default' &&
          pipelineId === 'edges' &&
          reprocess === false
        ) {
          return res(ctx.data({rerunPipeline: true}));
        }

        return res(
          ctx.errors([
            {
              message: 'unable to rerun pipeline',
              path: ['rerunPipeline'],
            },
          ]),
        );
      }),
    );
    render(<RerunPipelineModal pipelineId="edges" />);

    const submit = screen.getByRole('button', {
      name: /rerun pipeline/i,
    });

    expect(submit).toBeDisabled();

    await click(
      screen.getByRole('radio', {
        name: /process only failed datums/i,
      }),
    );

    await click(submit);
    expect(
      await screen.findByText('Pipeline was rerun successfully'),
    ).toBeInTheDocument();
  });

  describe('RerunPipelineButton', () => {
    beforeEach(() => {
      server.resetHandlers();
    });

    it('should disable button if the pipeline is a spout pipeline', async () => {
      window.history.replaceState({}, '', '/lineage/default/pipelines/montage');
      server.use(mockTrueGetAuthorize());
      server.use(mockGetSpoutPipeline());
      render(<RerunPipelineButton>Rerun Pipeline</RerunPipelineButton>);

      await waitFor(() =>
        expect(
          screen.getByRole('button', {
            name: /rerun pipeline/i,
          }),
        ).toBeDisabled(),
      );
    });

    it('should disable button if user does not have permission', async () => {
      server.use(mockFalseGetAuthorize());
      render(<RerunPipelineButton>Rerun Pipeline</RerunPipelineButton>);

      await waitFor(() =>
        expect(
          screen.getByRole('button', {
            name: /rerun pipeline/i,
          }),
        ).toBeDisabled(),
      );
    });

    it('should enable button if user has permission', async () => {
      server.use(mockTrueGetAuthorize());
      render(<RerunPipelineButton>Rerun Pipeline</RerunPipelineButton>);

      expect(
        screen.getByRole('button', {
          name: /rerun pipeline/i,
        }),
      ).toBeEnabled();
    });
  });
});
