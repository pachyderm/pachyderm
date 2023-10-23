import {PipelineType, mockPipelineQuery} from '@graphqlTypes';
import {
  render,
  within,
  screen,
  waitForElementToBeRemoved,
} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  buildPipeline,
  mockEmptyGetAuthorize,
  mockGetMontagePipeline,
} from '@dash-frontend/mocks';
import {withContextProviders} from '@dash-frontend/testHelpers';

import PipelineInfoComponent from '../PipelineInfo';

describe('PipelineInfo', () => {
  const server = setupServer();

  const PipelineInfo = withContextProviders(() => <PipelineInfoComponent />);

  beforeAll(() => {
    server.listen();
    server.use(mockEmptyGetAuthorize());
    server.use(mockGetMontagePipeline());
  });

  afterAll(() => server.close());

  it('should display information about the pipeline', async () => {
    window.history.replaceState('', '', `/lineage/default/pipelines/montage`);

    render(<PipelineInfo />);

    expect(
      await screen.findByRole('heading', {name: 'Failure'}),
    ).toBeInTheDocument();

    const outputRepo = screen.getByLabelText('Output Repo');
    expect(outputRepo).toHaveTextContent('montage');
    expect(within(outputRepo).getByRole('link')).toHaveAttribute(
      'href',
      `/lineage/default/repos/montage`,
    );

    expect(screen.getByLabelText('Failure Reason')).toHaveTextContent(
      'Pipeline failed because we have no memory!',
    );
    expect(screen.getByLabelText('Pipeline Type')).toHaveTextContent(
      'Standard',
    );
    expect(screen.getByLabelText('Datum Timeout')).toHaveTextContent('N/A');
    expect(screen.getByLabelText('Datum Tries')).toHaveTextContent('3');
    expect(screen.getByLabelText('Job Timeout')).toHaveTextContent('N/A');
    expect(screen.getByLabelText('Output Branch')).toHaveTextContent('master');
    expect(screen.getByLabelText('Egress')).toHaveTextContent('No');
    expect(screen.getByLabelText('S3 Output Repo')).toHaveTextContent(`N/A`);
  });

  it('should hide items if a global id filter is applied', async () => {
    window.history.replaceState(
      '',
      '',
      `/lineage/default/pipelines/montage?globalIdFilter=be7e4aa9caf148bb886b469d19f482d1`,
    );

    render(<PipelineInfo />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    const outputRepo = screen.getByLabelText('Output Repo');
    expect(outputRepo).toHaveTextContent('montage');

    expect(
      screen.queryByRole('heading', {name: 'Failure'}),
    ).not.toBeInTheDocument();

    expect(screen.queryByLabelText('Failure Reason')).not.toBeInTheDocument();
  });

  it('shows an S3 output when it is present', async () => {
    window.history.replaceState(
      '',
      '',
      `/lineage/default/pipelines/doesnt-matter`,
    );

    server.use(
      mockPipelineQuery((_req, res, ctx) => {
        return res(
          ctx.data({
            pipeline: buildPipeline({
              s3OutputRepo: 'http://an/internet/link',
              type: PipelineType.STANDARD,
              egress: true,
            }),
          }),
        );
      }),
    );

    render(<PipelineInfo />);

    expect(await screen.findByLabelText('Egress')).toHaveTextContent('Yes');
    expect(screen.getByLabelText('S3 Output Repo')).toHaveTextContent(
      `http://an/internet/link`,
    );
  });

  it('hides items for a service pipeline', async () => {
    window.history.replaceState(
      '',
      '',
      `/lineage/default/pipelines/service-pipeline`,
    );

    server.use(
      mockPipelineQuery((_req, res, ctx) => {
        return res(
          ctx.data({
            pipeline: buildPipeline({
              name: 'service-pipeline',
              s3OutputRepo: 'http://an/internet/link',
              type: PipelineType.SERVICE,
            }),
          }),
        );
      }),
    );

    render(<PipelineInfo />);

    expect(await screen.findByLabelText('Output Repo')).toHaveTextContent(
      'service-pipeline',
    );

    expect(screen.queryByTestId('Datum Timeout')).not.toBeInTheDocument();
    expect(screen.queryByTestId('Datum Tries')).not.toBeInTheDocument();
    expect(screen.queryByTestId('Job Timeout')).not.toBeInTheDocument();
    expect(screen.getByLabelText('Output Branch')).toHaveTextContent('master');
    expect(screen.queryByTestId('Egress')).not.toBeInTheDocument();
    expect(screen.queryByTestId('S3 Output Repo')).not.toBeInTheDocument();
  });

  it('hides items for a spout pipeline', async () => {
    window.history.replaceState(
      '',
      '',
      `/lineage/default/pipelines/spout-pipeline`,
    );

    server.use(
      mockPipelineQuery((_req, res, ctx) => {
        return res(
          ctx.data({
            pipeline: buildPipeline({
              name: 'spout-pipeline',
              type: PipelineType.SPOUT,
            }),
          }),
        );
      }),
    );
    render(<PipelineInfo />);

    expect(await screen.findByLabelText('Output Repo')).toHaveTextContent(
      'spout-pipeline',
    );

    expect(screen.getByLabelText('Pipeline Type')).toHaveTextContent('Spout');

    expect(screen.queryByTestId('Datum Timeout')).not.toBeInTheDocument();
    expect(screen.queryByTestId('Datum Tries')).not.toBeInTheDocument();
    expect(screen.queryByTestId('Job Timeout')).not.toBeInTheDocument();
    expect(screen.getByLabelText('Output Branch')).toHaveTextContent('master');
    expect(screen.queryByTestId('Egress')).not.toBeInTheDocument();
    expect(screen.queryByTestId('S3 Output Repo')).not.toBeInTheDocument();
  });
});
