import {
  render,
  waitForElementToBeRemoved,
  screen,
  within,
} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {Permission} from '@dash-frontend/api/auth';
import {
  mockEmptyGetAuthorize,
  mockTrueGetAuthorize,
  mockGetEnterpriseInfoInactive,
  mockEmptyGetRoles,
  mockPipelines,
  mockRepos,
} from '@dash-frontend/mocks';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import PipelineListComponent from '../PipelineList';

describe('Pipelines', () => {
  const server = setupServer();

  const PipelineList = withContextProviders(() => {
    return <PipelineListComponent />;
  });

  beforeAll(() => server.listen());

  beforeEach(() => {
    window.history.replaceState('', '', '/project/default/pipelines');

    server.resetHandlers();
    server.use(mockEmptyGetAuthorize());
    server.use(mockRepos());
    server.use(mockPipelines());
    server.use(mockGetEnterpriseInfoInactive());
    server.use(mockEmptyGetRoles());
  });

  afterAll(() => server.close());

  it('should display pipeline details', async () => {
    render(<PipelineList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('PipelineStepsTable__loadingDots'),
    );

    const pipelines = screen.getAllByTestId('PipelineListRow__row');
    expect(pipelines[0]).toHaveTextContent('montage');
    expect(pipelines[0]).toHaveTextContent('Error - Failure');
    expect(pipelines[0]).toHaveTextContent('Error - Killed');
    expect(pipelines[0]).toHaveTextContent('v:1');
    expect(pipelines[0]).toHaveTextContent('Jul 24, 2023; 17:58');
    expect(pipelines[0]).toHaveTextContent('clusterAdmin');
    expect(pipelines[0]).toHaveTextContent(
      'A pipeline that combines images from the `images` and `edges` repositories into a montage',
    );
  });

  it('should allow the user to select a subset of pipelines', async () => {
    render(<PipelineList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('PipelineStepsTable__loadingDots'),
    );

    expect(
      screen.getByText('Select rows to view detailed info'),
    ).toBeInTheDocument();
    expect(await screen.findAllByTestId('PipelineListRow__row')).toHaveLength(
      2,
    );

    await click(screen.getAllByText('montage')[0]);
    await click(screen.getAllByText('edges')[0]);

    expect(
      screen.getByText('Detailed info for 2 pipelines'),
    ).toBeInTheDocument();
  });

  it('should sort pipelines', async () => {
    render(<PipelineList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('PipelineStepsTable__loadingDots'),
    );

    const expand = screen.getByRole('button', {
      name: /expand filters/i,
    });

    let pipelines = screen.getAllByTestId('PipelineListRow__row');
    expect(pipelines[0]).toHaveTextContent('montage');
    expect(pipelines[1]).toHaveTextContent('edges');

    await click(expand);
    await click(screen.getByRole('radio', {name: /Created: Oldest/}));

    pipelines = screen.getAllByTestId('PipelineListRow__row');
    expect(pipelines[0]).toHaveTextContent('edges');
    expect(pipelines[1]).toHaveTextContent('montage');

    await click(expand);
    await click(screen.getByRole('radio', {name: /Alphabetical: A-Z/}));

    pipelines = screen.getAllByTestId('PipelineListRow__row');
    expect(pipelines[0]).toHaveTextContent('edges');
    expect(pipelines[1]).toHaveTextContent('montage');

    await click(expand);
    await click(screen.getByRole('radio', {name: /Alphabetical: Z-A/}));

    pipelines = screen.getAllByTestId('PipelineListRow__row');
    expect(pipelines[0]).toHaveTextContent('montage');
    expect(pipelines[1]).toHaveTextContent('edges');

    await click(expand);
    await click(screen.getByRole('radio', {name: /Job Status/}));

    pipelines = screen.getAllByTestId('PipelineListRow__row');
    expect(pipelines[0]).toHaveTextContent('edges');
    expect(pipelines[1]).toHaveTextContent('montage');
  });

  it('should filter by pipeline state', async () => {
    render(<PipelineList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('PipelineStepsTable__loadingDots'),
    );

    expect(await screen.findAllByTestId('PipelineListRow__row')).toHaveLength(
      2,
    );
    await click(
      screen.getByRole('button', {
        name: /expand filters/i,
      }),
    );
    await click(screen.getByRole('checkbox', {name: 'Idle'}));

    const pipelines = await screen.findAllByTestId('PipelineListRow__row');
    expect(pipelines).toHaveLength(1);
    expect(pipelines[0]).toHaveTextContent('Idle - Running');
  });

  it('should filter by last job status', async () => {
    render(<PipelineList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('PipelineStepsTable__loadingDots'),
    );

    expect(await screen.findAllByTestId('PipelineListRow__row')).toHaveLength(
      2,
    );

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByRole('checkbox', {name: 'Success', hidden: false}));
    expect(screen.getByText('No matching results')).toBeInTheDocument();
  });

  describe('with permission', () => {
    beforeEach(() => {
      server.use(
        mockTrueGetAuthorize([
          Permission.REPO_MODIFY_BINDINGS,
          Permission.REPO_WRITE,
        ]),
      );
      server.use(mockEmptyGetRoles());
    });

    it('should allow users to open the roles modal', async () => {
      render(<PipelineList />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('PipelineStepsTable__loadingDots'),
      );

      const montageRow = screen.getAllByTestId('PipelineListRow__row')[0];
      await click(await within(montageRow).findByText('clusterAdmin'));

      const modal = await screen.findByRole('dialog');
      expect(modal).toBeInTheDocument();

      expect(
        within(modal).getByRole('heading', {
          name: 'Set Repo Level Roles: default/montage',
        }),
      ).toBeInTheDocument();
    });
  });
});
