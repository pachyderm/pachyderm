import {render, screen} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {ListRepoRequest, RepoInfo} from '@dash-frontend/api/pfs';
import {
  JobState,
  ListPipelineRequest,
  PipelineInfo,
  PipelineState,
} from '@dash-frontend/api/pps';
import {buildPipeline, buildRepo} from '@dash-frontend/mocks';
import {withContextProviders, click, type} from '@dash-frontend/testHelpers';

import ProjectStatusSearchComponent from '../ProjectStatusSearch';

describe('ProjectStatusSearch', () => {
  const server = setupServer();

  const ProjectStatusSearch = withContextProviders(
    ProjectStatusSearchComponent,
  );

  beforeAll(() => server.listen());

  beforeEach(() => server.resetHandlers());

  afterAll(() => server.close());

  beforeEach(() => {
    window.history.replaceState({}, '', '/lineage/default');
    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) => res(ctx.json([])),
      ),
    );
  });

  afterEach(() => {
    window.localStorage.removeItem('pachyderm-console-default');
  });

  it('the status search is not interactable when project is healthy', async () => {
    server.use(
      rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
        '/api/pps_v2.API/ListPipeline',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildPipeline({
                pipeline: {name: 'aa'},
                lastJobState: JobState.JOB_SUCCESS,
                state: PipelineState.PIPELINE_RUNNING,
              }),
              buildPipeline({
                pipeline: {name: 'ac'},
                lastJobState: JobState.JOB_CREATED,
                state: PipelineState.PIPELINE_STANDBY,
              }),
              buildPipeline({
                pipeline: {name: 'ab'},
                lastJobState: JobState.JOB_RUNNING,
                state: PipelineState.PIPELINE_PAUSED,
              }),
              buildPipeline({
                pipeline: {name: 'ab2'},
                lastJobState: JobState.JOB_EGRESSING,
                state: PipelineState.PIPELINE_STARTING,
              }),
              buildPipeline({
                pipeline: {name: 'ab1'},
                lastJobState: JobState.JOB_STARTING,
                state: PipelineState.PIPELINE_RESTARTING,
              }),
            ]),
          ),
      ),
    );

    render(<ProjectStatusSearch />);
    expect(await screen.findByText('Healthy Project')).toBeInTheDocument();
    expect(
      screen.queryByRole('button', {name: /unhealthy Project/i}),
    ).not.toBeInTheDocument();
  });

  it('the status search is interactable when the project is unhealthy', async () => {
    server.use(
      rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
        '/api/pps_v2.API/ListPipeline',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildPipeline({
                pipeline: {name: 'aa'},
                lastJobState: JobState.JOB_FAILURE,
                state: PipelineState.PIPELINE_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: 'ac'},
                lastJobState: JobState.JOB_KILLED,
                state: PipelineState.PIPELINE_CRASHING,
              }),
            ]),
          ),
      ),
    );

    render(<ProjectStatusSearch />);
    expect(
      await screen.findByRole('button', {name: /unhealthy project/i}),
    ).toBeInTheDocument();
  });

  it('failed pipeline and subjobs are sorted into two groups', async () => {
    server.use(
      rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
        '/api/pps_v2.API/ListPipeline',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildPipeline({
                pipeline: {name: 'ps1'},
                lastJobState: JobState.JOB_FAILURE,
                state: PipelineState.PIPELINE_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: 's1'},
                lastJobState: JobState.JOB_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: 'p1'},
                state: PipelineState.PIPELINE_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: '1'},
              }),
            ]),
          ),
      ),
    );

    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildRepo({repo: {name: 'ps1'}}),
              buildRepo({repo: {name: 's1'}}),
              buildRepo({repo: {name: 'p1'}}),
              buildRepo({repo: {name: '1'}}),
            ]),
          ),
      ),
    );

    render(<ProjectStatusSearch />);

    await click(
      await screen.findByRole('button', {name: /unhealthy project/i}),
    );

    await click(await screen.findByText(/2 subjob failures/i));

    let items = screen.getAllByTestId('SearchResultItem__container');
    expect(items.map((el: HTMLElement) => el.textContent)).toEqual([
      'ps1',
      's1',
    ]);

    await click(await screen.findByText(/2 pipeline failures/i));

    items = screen.getAllByTestId('SearchResultItem__container');
    expect(items.map((el: HTMLElement) => el.textContent)).toEqual([
      'ps1',
      'p1',
    ]);
  });

  it('search bar should filter the results and update header counts', async () => {
    server.use(
      rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
        '/api/pps_v2.API/ListPipeline',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildPipeline({
                pipeline: {name: 'ps1'},
                lastJobState: JobState.JOB_FAILURE,
                state: PipelineState.PIPELINE_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: 's1'},
                lastJobState: JobState.JOB_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: 'p1'},
                state: PipelineState.PIPELINE_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: '1'},
              }),
            ]),
          ),
      ),
    );

    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildRepo({repo: {name: 'ps1'}}),
              buildRepo({repo: {name: 's1'}}),
              buildRepo({repo: {name: 'p1'}}),
              buildRepo({repo: {name: '1'}}),
            ]),
          ),
      ),
    );

    render(<ProjectStatusSearch />);

    await click(
      await screen.findByRole('button', {name: /unhealthy project/i}),
    );

    await screen.findByText(/2 pipeline failures/i);
    await click(await screen.findByText(/2 subjob failures/i));
    let items = screen.getAllByTestId('SearchResultItem__container');
    expect(items.map((el: HTMLElement) => el.textContent)).toEqual([
      'ps1',
      's1',
    ]);

    const searchBar = await screen.findByRole('searchbox');
    await type(searchBar, 's');

    await screen.findByText(/1 subjob failed/i);
    items = screen.getAllByTestId('SearchResultItem__container');
    expect(items.map((el: HTMLElement) => el.textContent)).toEqual(['s1']);

    await click(await screen.findByText(/0 pipeline failures/i));
    await screen.findByText('No matching Pipelines.');
  });

  it('should clear the search bar', async () => {
    server.use(
      rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
        '/api/pps_v2.API/ListPipeline',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildPipeline({
                pipeline: {name: 'ps1'},
                lastJobState: JobState.JOB_FAILURE,
                state: PipelineState.PIPELINE_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: 's1'},
                lastJobState: JobState.JOB_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: 'p1'},
                state: PipelineState.PIPELINE_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: '1'},
              }),
            ]),
          ),
      ),
    );

    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildRepo({repo: {name: 'ps1'}}),
              buildRepo({repo: {name: 's1'}}),
              buildRepo({repo: {name: 'p1'}}),
              buildRepo({repo: {name: '1'}}),
            ]),
          ),
      ),
    );

    render(<ProjectStatusSearch />);

    await click(
      await screen.findByRole('button', {name: /unhealthy project/i}),
    );

    await click(await screen.findByText(/2 subjob failures/i));
    let items = screen.getAllByTestId('SearchResultItem__container');
    expect(items.map((el: HTMLElement) => el.textContent)).toEqual([
      'ps1',
      's1',
    ]);

    const searchBar = await screen.findByRole('searchbox');
    await type(searchBar, 's');

    await screen.findByText(/1 subjob failed/i);
    items = screen.getAllByTestId('SearchResultItem__container');
    expect(items.map((el: HTMLElement) => el.textContent)).toEqual(['s1']);

    await click(await screen.findByLabelText(/clear search/i));
    expect(searchBar).toHaveTextContent('');
    await screen.findByText(/2 subjob failures/i);
    items = screen.getAllByTestId('SearchResultItem__container');
    expect(items.map((el: HTMLElement) => el.textContent)).toEqual([
      'ps1',
      's1',
    ]);
  });

  it('pipeline and subjob enums are correctly caught as failed state and unrunnable jobs are filtered out', async () => {
    server.use(
      rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
        '/api/pps_v2.API/ListPipeline',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildPipeline({
                pipeline: {name: 'lastJobState1'},
                lastJobState: JobState.JOB_CREATED,
              }),
              buildPipeline({
                pipeline: {name: 'lastJobState2'},
                lastJobState: JobState.JOB_EGRESSING,
              }),
              buildPipeline({
                pipeline: {name: 'lastJobState3'},
                lastJobState: JobState.JOB_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: 'lastJobState4'},
                lastJobState: JobState.JOB_FINISHING,
              }),
              buildPipeline({
                pipeline: {name: 'lastJobState5'},
                lastJobState: JobState.JOB_KILLED,
              }),
              buildPipeline({
                pipeline: {name: 'lastJobState6'},
                lastJobState: JobState.JOB_RUNNING,
              }),
              buildPipeline({
                pipeline: {name: 'lastJobState7'},
                lastJobState: JobState.JOB_STARTING,
              }),
              buildPipeline({
                pipeline: {name: 'lastJobState8'},
                lastJobState: JobState.JOB_STATE_UNKNOWN,
              }),
              buildPipeline({
                pipeline: {name: 'lastJobState9'},
                lastJobState: JobState.JOB_SUCCESS,
              }),
              buildPipeline({
                pipeline: {name: 'lastJobState10'},
                lastJobState: JobState.JOB_UNRUNNABLE,
              }),

              buildPipeline({
                pipeline: {name: 'state1'},
                state: PipelineState.PIPELINE_CRASHING,
              }),
              buildPipeline({
                pipeline: {name: 'state2'},
                state: PipelineState.PIPELINE_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: 'state3'},
                state: PipelineState.PIPELINE_PAUSED,
              }),
              buildPipeline({
                pipeline: {name: 'state4'},
                state: PipelineState.PIPELINE_RESTARTING,
              }),
              buildPipeline({
                pipeline: {name: 'state5'},
                state: PipelineState.PIPELINE_RUNNING,
              }),
              buildPipeline({
                pipeline: {name: 'state6'},
                state: PipelineState.PIPELINE_STANDBY,
              }),
              buildPipeline({
                pipeline: {name: 'state7'},
                state: PipelineState.PIPELINE_STARTING,
              }),
              buildPipeline({
                pipeline: {name: 'state8'},
                state: PipelineState.PIPELINE_STATE_UNKNOWN,
              }),
            ]),
          ),
      ),
    );
    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildRepo({repo: {name: 'lastJobState1'}}),
              buildRepo({repo: {name: 'lastJobState2'}}),
              buildRepo({repo: {name: 'lastJobState3'}}),
              buildRepo({repo: {name: 'lastJobState4'}}),
              buildRepo({repo: {name: 'lastJobState5'}}),
              buildRepo({repo: {name: 'lastJobState6'}}),
              buildRepo({repo: {name: 'lastJobState7'}}),
              buildRepo({repo: {name: 'lastJobState8'}}),
              buildRepo({repo: {name: 'lastJobState9'}}),
              buildRepo({repo: {name: 'lastJobState10'}}),
              buildRepo({repo: {name: 'state1'}}),
              buildRepo({repo: {name: 'state2'}}),
              buildRepo({repo: {name: 'state3'}}),
              buildRepo({repo: {name: 'state4'}}),
              buildRepo({repo: {name: 'state5'}}),
              buildRepo({repo: {name: 'state6'}}),
              buildRepo({repo: {name: 'state7'}}),
              buildRepo({repo: {name: 'state8'}}),
            ]),
          ),
      ),
    );
    render(<ProjectStatusSearch />);

    await click(
      await screen.findByRole('button', {name: /unhealthy project/i}),
    );

    expect(await screen.findByText(/2 pipeline failures/i)).toBeInTheDocument();
    expect(await screen.findByText(/2 subjob failures/i)).toBeInTheDocument();
  });

  it('failed pipeline and subjobs are sorted alphanumerically', async () => {
    server.use(
      rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
        '/api/pps_v2.API/ListPipeline',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildPipeline({
                pipeline: {name: 'aa'},
                lastJobState: JobState.JOB_FAILURE,
                state: PipelineState.PIPELINE_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: 'ac'},
                lastJobState: JobState.JOB_FAILURE,
                state: PipelineState.PIPELINE_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: 'ab'},
                lastJobState: JobState.JOB_FAILURE,
                state: PipelineState.PIPELINE_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: 'ab2'},
                lastJobState: JobState.JOB_FAILURE,
                state: PipelineState.PIPELINE_FAILURE,
              }),
              buildPipeline({
                pipeline: {name: 'ab1'},
                lastJobState: JobState.JOB_FAILURE,
                state: PipelineState.PIPELINE_FAILURE,
              }),
            ]),
          ),
      ),
    );
    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildRepo({repo: {name: 'aa'}}),
              buildRepo({repo: {name: 'ac'}}),
              buildRepo({repo: {name: 'ab'}}),
              buildRepo({repo: {name: 'ab2'}}),
              buildRepo({repo: {name: 'ab1'}}),
            ]),
          ),
      ),
    );
    render(<ProjectStatusSearch />);

    await click(
      await screen.findByRole('button', {name: /unhealthy project/i}),
    );

    expect(await screen.findByText(/5 pipeline failures/i)).toBeInTheDocument();
    let items = screen.getAllByTestId('SearchResultItem__container');
    expect(items.map((el: HTMLElement) => el.textContent)).toEqual([
      'aa',
      'ab',
      'ab1',
      'ab2',
      'ac',
    ]);

    expect(await screen.findByText(/5 subjob failures/i)).toBeInTheDocument();
    items = screen.getAllByTestId('SearchResultItem__container');
    expect(items.map((el: HTMLElement) => el.textContent)).toEqual([
      'aa',
      'ab',
      'ab1',
      'ab2',
      'ac',
    ]);
  });

  it('should route to selected pipeline', async () => {
    server.use(
      rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
        '/api/pps_v2.API/ListPipeline',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildPipeline({
                pipeline: {name: 'edges'},
                lastJobState: JobState.JOB_FAILURE,
                state: PipelineState.PIPELINE_FAILURE,
              }),
            ]),
          ),
      ),
    );
    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) => res(ctx.json([buildRepo({repo: {name: 'edges'}})])),
      ),
    );
    render(<ProjectStatusSearch />);

    await click(
      await screen.findByRole('button', {name: /unhealthy project/i}),
    );

    await click((await screen.findAllByText('edges'))[0]);

    expect(window.location.pathname).toBe('/lineage/default/pipelines/edges');
  });

  it('should display "killed" if the job is of the killed error type', async () => {
    server.use(
      rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
        '/api/pps_v2.API/ListPipeline',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildPipeline({
                pipeline: {name: 'edges'},
                lastJobState: JobState.JOB_KILLED,
                state: PipelineState.PIPELINE_RUNNING,
              }),
            ]),
          ),
      ),
    );
    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) => res(ctx.json([buildRepo({repo: {name: 'edges'}})])),
      ),
    );
    render(<ProjectStatusSearch />);

    await click(
      await screen.findByRole('button', {name: /unhealthy project/i}),
    );

    await click(await screen.findByText(/1 subjob failed/i));
    const item = screen.getByTestId('SearchResultItem__container');
    expect(item).toHaveTextContent('edgesKilled');
  });

  it('should display an empty message when there are no items', async () => {
    server.use(
      rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
        '/api/pps_v2.API/ListPipeline',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildPipeline({
                pipeline: {name: 'edges'},
                lastJobState: JobState.JOB_KILLED,
                state: PipelineState.PIPELINE_RUNNING,
              }),
            ]),
          ),
      ),
    );
    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) => res(ctx.json([buildRepo({repo: {name: 'edges'}})])),
      ),
    );
    render(<ProjectStatusSearch />);

    await click(
      await screen.findByRole('button', {name: /unhealthy project/i}),
    );

    const searchBar = await screen.findByRole('searchbox');
    await type(searchBar, 'zzzz');

    await click(await screen.findByText(/0 subjob failures/i));
    await screen.findByText('No matching Subjobs.');
    await click(await screen.findByText(/0 pipeline failures/i));
    await screen.findByText('No matching Pipelines.');
  });
});
