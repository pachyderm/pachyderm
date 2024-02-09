import {render, screen, act, waitFor} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {ListRepoRequest, RepoInfo} from '@dash-frontend/api/pfs';
import {ListPipelineRequest, PipelineInfo} from '@dash-frontend/api/pps';
import {
  buildPipeline,
  buildRepo,
  mockPipelinesEmpty,
  mockReposEmpty,
} from '@dash-frontend/mocks';
import {
  click,
  withContextProviders,
  type,
  clear,
} from '@dash-frontend/testHelpers';

import SearchComponent from '../Search';

describe('Search', () => {
  const server = setupServer();

  const Search = withContextProviders(SearchComponent);

  beforeAll(() => server.listen());

  beforeEach(() => server.resetHandlers());

  afterAll(() => server.close());

  beforeEach(() => {
    window.history.replaceState({}, '', '/lineage/default');
  });

  afterEach(() => {
    window.localStorage.removeItem('pachyderm-console-default');
  });

  it('should hide and show the search dropdown based on focus and it should be empty.', async () => {
    server.use(mockReposEmpty());
    server.use(mockPipelinesEmpty());

    render(<Search />);

    const searchBar = await screen.findByRole('searchbox');
    expect(screen.queryByTestId('Search__dropdown')).not.toBeInTheDocument();

    act(() => searchBar.focus());
    expect(screen.getByTestId('Search__dropdown')).toBeInTheDocument();

    expect(screen.queryByText(/failed/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/recent/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/clear/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/pipeline/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/repo/i)).not.toBeInTheDocument();
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
    await type(searchBar, 'im');

    expect(await screen.findByText('im')).toHaveClass('underline');
    expect(await screen.findByText('ages')).not.toHaveClass('underline');
  });

  it('should keep the dropdown open when clicking an item', async () => {
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

    // REPO
    const searchBar = await screen.findByRole('searchbox');

    await type(searchBar, 'edges');
    const searchResult = await screen.findByRole('button', {
      name: 'See Commits',
    });
    await click(searchResult);
    expect(screen.getByTestId('Search__dropdown')).toBeInTheDocument();

    // PIPELINE
    await click((await screen.findAllByText('edges'))[0]);

    expect(await screen.findByTestId('Search__dropdown')).toBeInTheDocument();
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

    await type(searchBar, 'images');
    const searchResult = await screen.findByRole('button', {
      name: 'See Commits',
    });
    await click(searchResult);
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

    await type(searchBar, 'edges');
    await click((await screen.findAllByText('edges'))[0]);

    expect(window.location.pathname).toBe('/lineage/default/pipelines/edges');
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

    await type(searchBar, 'images');
    const searchResult = await screen.findByRole('button', {
      name: 'See Commits',
    });
    await click(searchResult);

    await clear(searchBar);
    const savedSearch = await screen.findByRole('button', {name: 'images'});
    await click(savedSearch);
    expect(searchBar).toHaveValue('images');
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

    const searchBar = await screen.findByRole('searchbox');

    await type(searchBar, 'images');
    await click(await screen.findByRole('button', {name: 'See Commits'}));

    await clear(searchBar);

    // this is the recent search item
    await waitFor(() => {
      expect(screen.getAllByText('images')).toHaveLength(1);
    });

    await click(await screen.findByRole('button', {name: 'Clear'}));
    expect(screen.queryByText('images')).not.toBeInTheDocument();
  });
});
