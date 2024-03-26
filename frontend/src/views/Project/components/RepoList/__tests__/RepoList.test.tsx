import {
  render,
  screen,
  waitFor,
  waitForElementToBeRemoved,
  within,
} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  CommitInfo,
  ListCommitRequest,
  ListRepoRequest,
  RepoInfo,
} from '@dash-frontend/api/pfs';
import {
  mockEmptyGetAuthorize,
  mockTrueGetAuthorize,
  mockRepos,
  buildCommit,
  mockEmptyGetRoles,
  mockGetEnterpriseInfo,
  mockPipelinesEmpty,
  mockGetVersionInfo,
  generatePagingRepos,
} from '@dash-frontend/mocks';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import RepoListComponent from '../RepoList';

const mockMontageLastCommit = () =>
  rest.post<ListCommitRequest, Empty, CommitInfo[]>(
    '/api/pfs_v2.API/ListCommit',
    async (req, res, ctx) => {
      const body = await req.json();
      if (body['repo']['name'] === 'montage') {
        return res(
          ctx.json([
            buildCommit({
              commit: {
                id: '544e62c9a16f4a0aa995eb993ded1e52',
                repo: {
                  name: 'montage',
                  project: {name: 'default'},
                },
                branch: {
                  name: 'master',
                  __typename: 'Branch',
                },
              },
              description: 'commit not finished',
              started: '2023-07-24T17:58:38Z',
              finished: '2023-07-24T18:01:40Z',
              sizeBytesUpperBound: '797',
            }),
          ]),
        );
      }
    },
  );

const mockEdgesLastCommit = () =>
  rest.post<ListCommitRequest, Empty, CommitInfo[]>(
    '/api/pfs_v2.API/ListCommit',
    async (req, res, ctx) => {
      const body = await req.json();
      if (body['repo']['name'] === 'edges') {
        return res(
          ctx.json([
            buildCommit({
              commit: {
                id: '544e62c9a16f4a0aa995eb993ded1e52',
                repo: {
                  name: 'edges',
                  project: {name: 'default'},
                },
                branch: {
                  name: 'master',
                  __typename: 'Branch',
                },
              },
              started: '2023-07-24T17:58:25Z',
              finished: '2023-07-24T17:58:38Z',
              sizeBytesUpperBound: '0',
            }),
          ]),
        );
      }
    },
  );

const mockImagesLastCommit = () =>
  rest.post<ListCommitRequest, Empty, CommitInfo[]>(
    '/api/pfs_v2.API/ListCommit',
    async (req, res, ctx) => {
      const body = await req.json();
      if (body['repo']['name'] === 'images') {
        return res(
          ctx.json([
            buildCommit({
              commit: {
                id: '4eb1aa567dab483f93a109db4641ee75',
                repo: {
                  name: 'images',
                  project: {name: 'default'},
                },
                branch: {
                  name: 'master',
                  __typename: 'Branch',
                },
              },
              started: '2023-07-24T17:58:25Z',
              finished: undefined,
              sizeBytesUpperBound: '0',
            }),
          ]),
        );
      }
    },
  );

const mockEmptyLastCommit = () =>
  rest.post<ListCommitRequest, Empty, CommitInfo[]>(
    '/api/pfs_v2.API/ListCommit',
    async (req, res, ctx) => {
      const body = await req.json();
      if (body['repo']['name'] === 'empty') {
        return res(ctx.json([]));
      }
    },
  );

describe('Repo List', () => {
  const server = setupServer();

  const RepoList = withContextProviders(() => {
    return <RepoListComponent />;
  });

  beforeAll(() => {
    server.listen();
    server.use(mockGetVersionInfo());
    server.use(mockEmptyGetRoles());
    server.use(mockEmptyGetAuthorize());
    server.use(mockGetEnterpriseInfo());
    server.use(mockRepos());
    server.use(mockMontageLastCommit());
    server.use(mockEdgesLastCommit());
    server.use(mockImagesLastCommit());
    server.use(mockEmptyLastCommit());
    server.use(mockPipelinesEmpty());
  });

  beforeEach(() => {
    window.history.replaceState('', '', '/project/default/repos');
  });

  afterAll(() => server.close());

  it('should display repo details', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    const repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[1]).toHaveTextContent('edges');
    expect(repos[1]).toHaveTextContent('22.79 kB');
    expect(repos[1]).toHaveTextContent('Jul 24, 2023; 17:58');

    await waitFor(() =>
      expect(screen.getAllByTestId('RepoListRow__row')[1]).toHaveTextContent(
        '544e62..',
      ),
    );

    expect(repos[1]).toHaveTextContent(
      'Output repo for pipeline default/edges.',
    );
    expect(repos[1]).toHaveTextContent('clusterAdmin');

    expect(repos[3]).toHaveTextContent('empty');
    expect(repos[3]).toHaveTextContent('Jul 24, 2023; 17:58');
    expect(repos[3]).toHaveTextContent('0 B');
    expect(repos[3]).toHaveTextContent('N/A');
    expect(repos[3]).toHaveTextContent('clusterAdmin, repoOwner');
  });

  it('should allow the user to select a repo', async () => {
    render(<RepoList />);

    expect(
      screen.getByText('Select a row to view detailed info'),
    ).toBeInTheDocument();

    expect(await screen.findAllByTestId('RepoListRow__row')).toHaveLength(4);
    await click(screen.getByText('edges'));

    expect(screen.getByText('Detailed info for edges')).toBeInTheDocument();
  });

  it('should sort repos', async () => {
    render(<RepoList />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    await waitFor(() =>
      expect(screen.getAllByTestId('RepoListRow__row')[1]).toHaveTextContent(
        '544e62..',
      ),
    );

    const expand = screen.getByRole('button', {
      name: /expand filters/i,
    });

    let repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[0]).toHaveTextContent('montage');
    expect(repos[1]).toHaveTextContent('edges');
    expect(repos[2]).toHaveTextContent('images');
    expect(repos[3]).toHaveTextContent('empty');

    await click(expand);
    await click(screen.getByRole('radio', {name: /Created: Oldest/}));

    repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[0]).toHaveTextContent('empty');
    expect(repos[1]).toHaveTextContent('images');
    expect(repos[2]).toHaveTextContent('edges');
    expect(repos[3]).toHaveTextContent('montage');

    await click(expand);
    await click(screen.getByRole('radio', {name: /Alphabetical: A-Z/}));

    repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[0]).toHaveTextContent('edges');
    expect(repos[1]).toHaveTextContent('empty');
    expect(repos[2]).toHaveTextContent('images');
    expect(repos[3]).toHaveTextContent('montage');

    await click(expand);
    await click(screen.getByRole('radio', {name: /Alphabetical: Z-A/}));

    repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[0]).toHaveTextContent('montage');
    expect(repos[1]).toHaveTextContent('images');
    expect(repos[2]).toHaveTextContent('empty');
    expect(repos[3]).toHaveTextContent('edges');

    await click(expand);
    await click(screen.getByRole('radio', {name: /Size/}));

    repos = screen.getAllByTestId('RepoListRow__row');
    expect(repos[0]).toHaveTextContent('empty');
    expect(repos[1]).toHaveTextContent('images');
    expect(repos[2]).toHaveTextContent('edges');
    expect(repos[3]).toHaveTextContent('montage');
  });

  describe('with repo edit permission', () => {
    beforeAll(() => {
      server.use(mockTrueGetAuthorize());
      server.use(mockEmptyGetRoles());
    });

    it('should allow users to open the roles modal', async () => {
      render(<RepoList />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      const montageRow = screen.getAllByTestId('RepoListRow__row')[0];
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

  describe('repo paging', () => {
    beforeEach(() => {
      const repos = generatePagingRepos(30);
      server.use(
        rest.post<ListRepoRequest, Empty, RepoInfo[]>(
          '/api/pfs_v2.API/ListRepo',
          async (req, res, ctx) => {
            const body = await req.json();
            const {page} = body;
            const {pageSize, pageIndex} = page;

            if (pageSize === '15' && pageIndex === '0') {
              return res(ctx.json(repos.slice(0, 15)));
            }

            if (pageSize === '15' && pageIndex === '1') {
              return res(ctx.json(repos.slice(15, 29)));
            }

            if (pageSize === '50' && pageIndex === '0') {
              return res(ctx.json(repos));
            }

            return res(ctx.json(repos));
          },
        ),
      );
      server.use(
        rest.post<ListCommitRequest, Empty, CommitInfo[]>(
          '/api/pfs_v2.API/ListCommit',
          async (_req, res, ctx) => {
            return res(ctx.json([]));
          },
        ),
      );
    });

    it('should allow users to navigate through paged repos', async () => {
      render(<RepoList />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      let repos = screen.getAllByTestId('RepoListRow__row');
      expect(repos).toHaveLength(15);
      expect(repos[0]).toHaveTextContent('repo-0');
      expect(repos[14]).toHaveTextContent('repo-14');

      let pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();
      await click(within(pager).getByTestId('Pager__forward'));

      expect(await screen.findByText('repo-15')).toBeInTheDocument();
      repos = screen.getAllByTestId('RepoListRow__row');
      expect(repos).toHaveLength(14);
      expect(repos[0]).toHaveTextContent('repo-15');
      repos = screen.getAllByTestId('RepoListRow__row');
      expect(repos[0]).toHaveTextContent('repo-15');
      expect(repos[13]).toHaveTextContent('repo-28');

      pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__forward')).toBeDisabled();
      await click(within(pager).getByTestId('Pager__backward'));

      repos = screen.getAllByTestId('RepoListRow__row');
      expect(repos).toHaveLength(15);
      expect(repos[0]).toHaveTextContent('repo-0');
      expect(repos[14]).toHaveTextContent('repo-14');
    });

    it('should allow users to update page size', async () => {
      render(<RepoList />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      expect(await screen.findByText('repo-0')).toBeInTheDocument();

      let repos = screen.getAllByTestId('RepoListRow__row');
      expect(repos).toHaveLength(15);

      const pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__forward')).toBeEnabled();
      expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();

      await click(within(pager).getByTestId('DropdownButton__button'));
      await click(within(pager).getByText(50));

      expect(await screen.findByText('repo-16')).toBeInTheDocument();
      repos = screen.getAllByTestId('RepoListRow__row');
      expect(repos).toHaveLength(30);
    });
  });
});
