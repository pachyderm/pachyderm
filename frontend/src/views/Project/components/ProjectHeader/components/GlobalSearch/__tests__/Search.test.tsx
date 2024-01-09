import {render, screen, act} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {ListRepoRequest, RepoInfo} from '@dash-frontend/api/pfs';
import {
  InspectJobSetRequest,
  JobInfo,
  ListPipelineRequest,
  PipelineInfo,
} from '@dash-frontend/api/pps';
import {
  buildJob,
  buildPipeline,
  buildRepo,
  mockPipelinesEmpty,
  mockReposEmpty,
} from '@dash-frontend/mocks';
import {click, withContextProviders, type} from '@dash-frontend/testHelpers';

import SearchComponent from '../Search';

describe('Search', () => {
  const server = setupServer();

  const Search = withContextProviders(SearchComponent);

  beforeAll(() => server.listen());

  beforeEach(() => server.resetHandlers());

  afterAll(() => server.close());

  const assertDropdown = () => {
    const dropdown = screen.getByTestId('Search__dropdown');

    return {
      toBeShown: () => expect(dropdown).toHaveClass('open'),
      toBeHidden: () => expect(dropdown).not.toHaveClass('open'),
    };
  };

  beforeEach(() => {
    window.history.replaceState({}, '', '/lineage/default');
  });

  afterEach(() => {
    window.localStorage.removeItem('pachyderm-console-default');
  });

  it('should display empty state messages', async () => {
    server.use(mockReposEmpty());
    server.use(mockPipelinesEmpty());

    render(<Search />);

    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    act(() => searchBar.focus());
    assertDropdown().toBeShown();
    expect(
      screen.getByText('There are no recent searches.'),
    ).toBeInTheDocument();
    await type(searchBar, 'anything');
    expect(
      await screen.findByText(
        'No matching pipelines, jobs or Global ID found.',
      ),
    ).toBeInTheDocument();
    assertDropdown().toBeShown();
  });

  it('should underline matched search term', async () => {
    server.use(mockPipelinesEmpty());
    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) =>
          res(ctx.json([buildRepo({repo: {name: 'images'}})])),
      ),
    );
    render(<Search />);

    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    await type(searchBar, 'im');
    assertDropdown().toBeShown();
    expect(await screen.findByText('im')).toHaveClass('underline');
    expect(await screen.findByText('ages')).not.toHaveClass('underline');
  });

  it('should save recent searches and populate search bar', async () => {
    server.use(mockPipelinesEmpty());
    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) =>
          res(ctx.json([buildRepo({repo: {name: 'images'}})])),
      ),
    );
    render(<Search />);

    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    await type(searchBar, 'images');
    assertDropdown().toBeShown();
    const searchResult = await screen.findByRole('button', {
      name: 'See Commits',
    });
    await click(searchResult);
    assertDropdown().toBeHidden();

    act(() => searchBar.focus());
    assertDropdown().toBeShown();
    const savedSearch = await screen.findByRole('button', {name: 'images'});
    await click(savedSearch);
    expect(searchBar).toHaveValue('images');
    assertDropdown().toBeShown();
  });

  it('should clear saved search history', async () => {
    server.use(mockPipelinesEmpty());
    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) =>
          res(ctx.json([buildRepo({repo: {name: 'images'}})])),
      ),
    );
    render(<Search />);

    expect(
      screen.getByText('There are no recent searches.'),
    ).toBeInTheDocument();

    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    await type(searchBar, 'images');
    assertDropdown().toBeShown();
    await click(await screen.findByRole('button', {name: 'See Commits'}));
    assertDropdown().toBeHidden();

    act(() => searchBar.focus());
    assertDropdown().toBeShown();
    expect(
      screen.queryByText('There are no recent searches.'),
    ).not.toBeInTheDocument();
    await click(await screen.findByRole('button', {name: 'Clear'}));
    expect(
      screen.getByText('There are no recent searches.'),
    ).toBeInTheDocument();
    assertDropdown().toBeShown();
  });

  it('should route to selected repo', async () => {
    server.use(mockPipelinesEmpty());
    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) =>
          res(ctx.json([buildRepo({repo: {name: 'images'}})])),
      ),
    );
    render(<Search />);

    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    await type(searchBar, 'images');
    assertDropdown().toBeShown();
    const searchResult = await screen.findByRole('button', {
      name: 'See Commits',
    });
    await click(searchResult);

    assertDropdown().toBeHidden();
    expect(window.location.pathname).toBe('/lineage/default/repos/images');
  });

  it('should route to selected pipeline', async () => {
    server.use(
      rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
        '/api/pps_v2.API/ListPipeline',
        (_req, res, ctx) =>
          res(ctx.json([buildPipeline({pipeline: {name: 'edges'}})])),
      ),
    );
    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) => res(ctx.json([buildRepo({repo: {name: 'edges'}})])),
      ),
    );
    render(<Search />);

    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    await type(searchBar, 'edges');
    assertDropdown().toBeShown();
    await click((await screen.findAllByText('edges'))[1]);

    assertDropdown().toBeHidden();
    expect(window.location.pathname).toBe('/lineage/default/pipelines/edges');
    assertDropdown().toBeHidden();
  });

  it('should route to jobs for selected id', async () => {
    server.use(
      rest.post<InspectJobSetRequest, Empty, JobInfo[]>(
        '/api/pps_v2.API/InspectJobSet',
        (_req, res, ctx) =>
          res(
            ctx.json([
              buildJob({job: {id: '1dc67e479f03498badcc6180be4ee6ce'}}),
            ]),
          ),
      ),
    );
    render(<Search />);

    const searchBar = await screen.findByRole('searchbox');
    assertDropdown().toBeHidden();

    await type(searchBar, '1dc67e479f03498badcc6180be4ee6ce');
    assertDropdown().toBeShown();
    await click(await screen.findByText('1dc67e479f03498badcc6180be4ee6ce'));

    assertDropdown().toBeHidden();
    expect(window.location.pathname).toBe('/project/default/jobs/subjobs');
    expect(window.location.search).toBe(
      '?selectedJobs=1dc67e479f03498badcc6180be4ee6ce',
    );
  });
});
