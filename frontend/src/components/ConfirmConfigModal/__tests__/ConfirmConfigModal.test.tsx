import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {mockGetEnterpriseInfoInactive} from '@dash-frontend/mocks';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import ConfirmConfigModalComponent from '../ConfirmConfigModal';

describe('ConfirmConfigModal', () => {
  const server = setupServer();

  const ConfirmConfigModal = withContextProviders(() => {
    return (
      <ConfirmConfigModalComponent
        level="Cluster"
        show
        onSubmit={() => null}
        loading={false}
        affectedPipelines={[
          {name: 'PipelineA', project: {name: 'ProjectA'}},
          {name: 'PipelineB', project: {name: 'ProjectA'}},
          {name: 'PipelineC', project: {name: 'ProjectB'}},
          {name: 'PipelineD', project: {name: 'ProjectB'}},
        ]}
      />
    );
  });

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    server.resetHandlers();
    server.use(mockGetEnterpriseInfoInactive());
  });

  afterAll(() => server.close());

  it('should link to affected pipelines by project and pipeline name', async () => {
    render(<ConfirmConfigModal />);

    expect(
      screen.getByRole('radio', {name: 'Save Cluster Defaults'}),
    ).toBeChecked();

    await click(
      screen.getByRole('radio', {
        name: /save cluster defaults and regenerate pipelines/i,
      }),
    );

    // project name should only appear once
    expect(screen.getByRole('link', {name: /projecta/i})).toHaveAttribute(
      'href',
      '/lineage/ProjectA',
    );
    expect(screen.getByRole('link', {name: /projectb/i})).toHaveAttribute(
      'href',
      '/lineage/ProjectB',
    );

    expect(screen.getByRole('link', {name: /pipelinea/i})).toHaveAttribute(
      'href',
      '/lineage/ProjectA/pipelines/PipelineA',
    );
    expect(screen.getByRole('link', {name: /pipelinec/i})).toHaveAttribute(
      'href',
      '/lineage/ProjectB/pipelines/PipelineC',
    );
    expect(
      screen.getByText('4 pipelines will be affected'),
    ).toBeInTheDocument();
  });

  it('should allow users to sort by project or pipeline name', async () => {
    render(<ConfirmConfigModal />);

    await click(
      screen.getByRole('radio', {
        name: /save cluster defaults and regenerate pipelines/i,
      }),
    );

    let list = screen.getAllByRole('link');

    expect(list[0]).toHaveTextContent('ProjectA');
    expect(list[1]).toHaveTextContent('PipelineA');
    expect(list[3]).toHaveTextContent('PipelineB');
    expect(list[4]).toHaveTextContent('ProjectB');
    expect(list[5]).toHaveTextContent('PipelineC');
    expect(list[7]).toHaveTextContent('PipelineD');

    await click(
      screen.getByRole('button', {
        name: /sort by project in descending order/i,
      }),
    );

    list = screen.getAllByRole('link');
    expect(list[0]).toHaveTextContent('ProjectB');
    expect(list[1]).toHaveTextContent('PipelineC');
    expect(list[3]).toHaveTextContent('PipelineD');
    expect(list[4]).toHaveTextContent('ProjectA');
    expect(list[5]).toHaveTextContent('PipelineA');
    expect(list[7]).toHaveTextContent('PipelineB');

    await click(
      screen.getByRole('button', {
        name: /sort by name in descending order/i,
      }),
    );

    list = screen.getAllByRole('link');
    expect(list[0]).toHaveTextContent('ProjectA');
    expect(list[1]).toHaveTextContent('PipelineA');
    expect(list[3]).toHaveTextContent('PipelineB');
    expect(list[4]).toHaveTextContent('ProjectB');
    expect(list[5]).toHaveTextContent('PipelineC');
    expect(list[7]).toHaveTextContent('PipelineD');
  });

  it('should allow users to show and hide the table', async () => {
    render(<ConfirmConfigModal />);

    expect(screen.queryByRole('table')).not.toBeInTheDocument();

    await click(
      screen.getByRole('radio', {
        name: /save cluster defaults and regenerate pipelines/i,
      }),
    );

    expect(screen.getByRole('table')).toBeInTheDocument();

    await click(screen.getByRole('button', {name: /hide list/i}));

    expect(screen.queryByRole('table')).not.toBeInTheDocument();

    await click(screen.getByRole('button', {name: /show list/i}));

    expect(screen.getByRole('table')).toBeInTheDocument();
  });
});
