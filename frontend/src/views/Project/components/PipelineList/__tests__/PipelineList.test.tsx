import {
  render,
  waitForElementToBeRemoved,
  screen,
  within,
  waitFor,
} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Permission} from '@dash-frontend/api/auth';
import {Empty} from '@dash-frontend/api/googleTypes';
import {InspectRepoRequest, RepoInfo} from '@dash-frontend/api/pfs';
import {ListPipelineRequest, PipelineInfo} from '@dash-frontend/api/pps';
import {
  mockTrueGetAuthorize,
  mockGetEnterpriseInfoInactive,
  mockEmptyGetRoles,
  mockPipelines,
  mockGetVersionInfo,
  generatePagingPipelines,
  REPO_INFO_EMPTY,
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
    server.use(mockGetVersionInfo());
    server.use(mockPipelines());
    server.use(mockGetEnterpriseInfoInactive());
    server.use(mockEmptyGetRoles());
    server.use(mockTrueGetAuthorize());
    server.use(
      rest.post<InspectRepoRequest, Empty, RepoInfo>(
        '/api/pfs_v2.API/InspectRepo',
        async (_req, res, ctx) => {
          return res(ctx.json(REPO_INFO_EMPTY));
        },
      ),
    );
  });

  afterAll(() => server.close());

  it('should display pipeline details', async () => {
    render(<PipelineList />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    const pipelines = screen.getAllByTestId('PipelineListRow__row');
    expect(pipelines[0]).toHaveTextContent('montage');
    expect(pipelines[0]).toHaveTextContent('Error - Failure');
    expect(pipelines[0]).toHaveTextContent('Error - Killed');
    expect(pipelines[0]).toHaveTextContent('v:1');
    expect(pipelines[0]).toHaveTextContent('Jul 24, 2023; 17:58');
    expect(pipelines[0]).toHaveTextContent(
      'A pipeline that combines images from the `images` and `edges` repositories into a montage',
    );
    await waitFor(() =>
      expect(
        screen.getAllByTestId('PipelineListRow__row')[0],
      ).toHaveTextContent('clusterAdmin'),
    );
  });

  it('should allow the user to select a subset of pipelines', async () => {
    render(<PipelineList />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

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

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

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

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

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

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

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

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      await waitFor(() =>
        expect(
          screen.getAllByTestId('PipelineListRow__row')[0],
        ).toHaveTextContent('clusterAdmin'),
      );
      const montageRow = screen.getAllByTestId('PipelineListRow__row')[0];
      await click(
        await within(montageRow).findByText('clusterAdmin, repoOwner'),
      );

      const modal = await screen.findByRole('dialog');
      expect(modal).toBeInTheDocument();

      expect(
        within(modal).getByRole('heading', {
          name: 'Set Repo Level Roles: default/montage',
        }),
      ).toBeInTheDocument();
    });
  });

  describe('pipeline paging', () => {
    beforeEach(() => {
      const pipelines = generatePagingPipelines(30);
      server.use(
        rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
          '/api/pps_v2.API/ListPipeline',
          async (req, res, ctx) => {
            const body = await req.json();
            const {page} = body;
            const {pageSize, pageIndex} = page;

            if (pageSize === '15' && pageIndex === '0') {
              return res(ctx.json(pipelines.slice(0, 15)));
            }

            if (pageSize === '15' && pageIndex === '1') {
              return res(ctx.json(pipelines.slice(15, 29)));
            }

            if (pageSize === '50' && pageIndex === '0') {
              return res(ctx.json(pipelines));
            }

            return res(ctx.json(pipelines));
          },
        ),
      );
    });

    it('should allow users to navigate through paged pipelines', async () => {
      render(<PipelineList />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      let pipelines = screen.getAllByTestId('PipelineListRow__row');
      expect(pipelines).toHaveLength(15);
      expect(pipelines[0]).toHaveTextContent('pipeline-0');
      expect(pipelines[14]).toHaveTextContent('pipeline-14');

      let pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();
      await click(within(pager).getByTestId('Pager__forward'));

      expect(await screen.findByText('pipeline-15')).toBeInTheDocument();
      pipelines = screen.getAllByTestId('PipelineListRow__row');
      expect(pipelines).toHaveLength(14);
      expect(pipelines[0]).toHaveTextContent('pipeline-15');
      pipelines = screen.getAllByTestId('PipelineListRow__row');
      expect(pipelines[0]).toHaveTextContent('pipeline-15');
      expect(pipelines[13]).toHaveTextContent('pipeline-28');

      pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__forward')).toBeDisabled();
      await click(within(pager).getByTestId('Pager__backward'));

      pipelines = screen.getAllByTestId('PipelineListRow__row');
      expect(pipelines).toHaveLength(15);
      expect(pipelines[0]).toHaveTextContent('pipeline-0');
      expect(pipelines[14]).toHaveTextContent('pipeline-14');
    });

    it('should allow users to update page size', async () => {
      render(<PipelineList />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      expect(await screen.findByText('pipeline-0')).toBeInTheDocument();

      let pipelines = screen.getAllByTestId('PipelineListRow__row');
      expect(pipelines).toHaveLength(15);

      const pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__forward')).toBeEnabled();
      expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();

      await click(within(pager).getByTestId('DropdownButton__button'));
      await click(within(pager).getByText(50));

      expect(await screen.findByText('pipeline-16')).toBeInTheDocument();
      pipelines = screen.getAllByTestId('PipelineListRow__row');
      expect(pipelines).toHaveLength(30);
    });
  });
});

//
