import {
  JobState,
  NodeState,
  PipelineState,
  PipelineType,
  mockCommitDiffQuery,
  mockJobQuery,
  mockPipelinesQuery,
  mockRepoQuery,
} from '@graphqlTypes';
import {
  render,
  waitFor,
  waitForElementToBeRemoved,
  screen,
  within,
} from '@testing-library/react';
import {setupServer} from 'msw/lib/node';
import React from 'react';

import {
  buildJob,
  mockEmptyGetRoles,
  mockGetCommitA4,
  mockGetImageCommits,
  mockGetMontageJob_5C,
  mockGetMontagePipeline,
  mockGetServicePipeline,
  mockGetSpoutPipeline,
  mockRepoImages,
  mockTrueGetAuthorize,
} from '@dash-frontend/mocks';
import {mockGetDag, mockGetLargerDag} from '@dash-frontend/mocks/dag';
import {click, withContextProviders} from '@dash-frontend/testHelpers';

import ProjectSidebar from '../ProjectSidebar';

describe('ProjectSidebar', () => {
  const server = setupServer();

  const Project = withContextProviders(ProjectSidebar);

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    server.use(mockGetDag());
    server.use(mockGetMontagePipeline());
    server.use(mockGetMontageJob_5C());
    server.use(mockTrueGetAuthorize());
    server.use(mockRepoImages());
    server.use(mockEmptyGetRoles());
  });

  afterEach(() => server.resetHandlers());

  afterAll(() => server.close());

  it('should not display the sidebar if not on a sidebar route', async () => {
    window.history.replaceState('', '', '/project/default');

    render(<Project />);

    expect(
      screen.queryByTestId('ProjectSidebar__sidebar'),
    ).not.toBeInTheDocument();
  });

  describe('PipelineDetails', () => {
    it('should display pipeline details', async () => {
      server.use(mockGetMontagePipeline());
      window.history.replaceState('', '', '/lineage/default/pipelines/montage');

      render(<Project />);

      expect(
        await screen.findByRole('heading', {
          name: 'montage',
        }),
      ).toBeInTheDocument();
    });

    it('should display pipeline logs button', async () => {
      server.use(mockGetMontagePipeline());
      server.use(mockGetMontageJob_5C());
      window.history.replaceState('', '', '/lineage/default/pipelines/montage');

      render(<Project />);
      const topLogsLink = await screen.findByRole('link', {
        name: 'Inspect Jobs',
      });
      expect(topLogsLink).toHaveAttribute(
        'href',
        '/lineage/default/pipelines/montage/jobs/5c1aa9bc87dd411ba5a1be0c80a3ebc2/logs',
      );
      expect(topLogsLink).toBeEnabled();

      const inspectLogsLink = await screen.findByRole('link', {
        name: 'Inspect Job',
      });
      expect(inspectLogsLink).toHaveAttribute(
        'href',
        '/lineage/default/pipelines/montage/jobs/5c1aa9bc87dd411ba5a1be0c80a3ebc2/logs/datum',
      );
      expect(inspectLogsLink).toBeEnabled();
    });

    it('should display datum logs link with filter applied', async () => {
      server.use(mockGetMontageJob_5C());
      window.history.replaceState('', '', '/lineage/default/pipelines/montage');
      render(<Project />);

      const logsLink = await screen.findByRole('link', {name: '1 Success'});
      expect(logsLink).toHaveAttribute(
        'href',
        '/lineage/default/pipelines/montage/jobs/5c1aa9bc87dd411ba5a1be0c80a3ebc2/logs/datum?datumFilters=SUCCESS',
      );
    });

    it('should disable the delete button when there are downstream pipelines', async () => {
      server.use(mockGetMontagePipeline());
      server.use(mockGetLargerDag());
      window.history.replaceState('', '', '/lineage/default/pipelines/montage');

      render(<Project />);

      expect(
        await screen.findByRole('heading', {name: 'montage'}),
      ).toBeInTheDocument();

      expect(
        await screen.findByRole('button', {
          name: /delete/i,
        }),
      ).toBeDisabled();
    });

    it('should enable the delete button when there are no downstream pipelines', async () => {
      server.use(mockGetMontagePipeline());
      server.use(mockGetDag());

      window.history.replaceState('', '', '/lineage/default/pipelines/montage');

      render(<Project />);

      expect(
        await screen.findByRole('heading', {
          name: 'montage',
        }),
      ).toBeInTheDocument();

      expect(
        await screen.findByRole('button', {
          name: /delete/i,
        }),
      ).toBeEnabled();
    });

    it('should show a linked project input node', async () => {
      server.use(
        mockJobQuery((req, res, ctx) => {
          return res(
            ctx.data({
              job: buildJob({
                id: '"23b9af7d5d4343219bc8e02ff44cd55a"',
                inputString: JSON.stringify({
                  pfs: {
                    project: 'Multi-Project-Pipeline-B',
                    name: 'Node_1',
                    repo: 'Node_1',
                    repoType: '',
                    branch: 'master',
                    commit: '',
                    glob: '',
                    joinOn: '',
                    outerJoin: false,
                    groupBy: '',
                    lazy: false,
                    emptyFiles: false,
                    s3: false,
                  },
                  join: [],
                  group: [],
                  cross: [],
                  union: [],
                }),
              }),
            }),
          );
        }),
      );

      window.history.replaceState(
        {},
        '',
        '/lineage/Multi-Project-Pipeline-A/pipelines/Node_2/job',
      );
      render(<Project />);

      expect((await screen.findAllByText('Success'))[0]).toBeInTheDocument();

      expect(
        await screen.findByText('Node_1 (Project Multi-Project-Pipeline-B)'),
      ).toBeInTheDocument();
    });

    it('should allow users to open the roles modal', async () => {
      server.use(mockGetMontagePipeline());
      server.use(mockGetMontageJob_5C());
      server.use(mockTrueGetAuthorize());
      server.use(mockRepoImages());

      window.history.replaceState('', '', '/lineage/default/pipelines/montage');

      render(<Project />);

      await click(
        await screen.findByRole('button', {
          name: /set roles via repo/i,
        }),
      );

      expect(await screen.findByRole('dialog')).toBeInTheDocument();
      expect(
        screen.getByText('Set Repo Level Roles: default/montage'),
      ).toBeInTheDocument();
    });

    it('should default to the info tab for a service pipeline', async () => {
      server.use(mockGetServicePipeline());
      server.use(mockGetMontageJob_5C());

      window.history.replaceState({}, '', '/lineage/default/pipelines/montage');
      render(<Project />);

      expect(
        await screen.findByRole('heading', {
          name: /running/i,
        }),
      ).toBeInTheDocument();

      expect(screen.getAllByRole('tab')).toHaveLength(2);
      expect(
        screen.getByRole('definition', {
          name: /pipeline type/i,
        }),
      ).toHaveTextContent(/service/i);
    });

    it('should default to the info tab for a spout pipeline', async () => {
      server.use(mockGetSpoutPipeline());

      window.history.replaceState({}, '', '/lineage/default/pipelines/montage');
      render(<Project />);

      expect(
        await screen.findByRole('heading', {
          name: /running/i,
        }),
      ).toBeInTheDocument();

      expect(screen.getAllByRole('tab')).toHaveLength(2);
      expect(
        screen.getByRole('definition', {
          name: /pipeline type/i,
        }),
      ).toHaveTextContent(/spout/i);
    });

    it('should hide recent job info for a spout pipeline', async () => {
      server.use(mockGetSpoutPipeline());

      window.history.replaceState({}, '', '/lineage/default/pipelines/montage');
      render(<Project />);

      expect(
        await screen.findByRole('heading', {
          name: /running/i,
        }),
      ).toBeInTheDocument();
      expect(
        screen.queryByText('Most Recent Job Start'),
      ).not.toBeInTheDocument();
      expect(screen.queryByText('Most Recent Job ID')).not.toBeInTheDocument();
    });
  });

  describe('repos', () => {
    beforeEach(() => {
      server.use(mockGetCommitA4());

      server.use(mockRepoImages());

      server.use(mockGetDag());

      server.use(
        mockCommitDiffQuery((req, res, ctx) =>
          res(
            ctx.data({
              commitDiff: {
                size: 58644,
                sizeDisplay: '58.65 kB',
                filesUpdated: {
                  count: 0,
                  sizeDelta: 0,
                  __typename: 'DiffCount',
                },
                filesAdded: {
                  count: 1,
                  sizeDelta: 58644,
                  __typename: 'DiffCount',
                },
                filesDeleted: {
                  count: 0,
                  sizeDelta: 0,
                  __typename: 'DiffCount',
                },
                __typename: 'Diff',
              },
            }),
          ),
        ),
      );

      server.use(mockGetImageCommits());
    });

    it('should display repo details', async () => {
      window.history.replaceState('', '', '/lineage/default/repos/images');

      render(<Project />);

      await screen.findByRole('heading', {name: 'images'});
      expect(screen.getAllByText('139.24 kB')).toHaveLength(2);
      await screen.findByText('4a83c74809664f899261baccdb47cd90');
      expect(screen.getByText('+ 58.65 kB')).toBeInTheDocument();

      await waitFor(() =>
        expect(
          screen.queryByTestId('CommitList__loadingdots'),
        ).not.toBeInTheDocument(),
      );

      expect(
        screen.getByRole('link', {
          name: 'Inspect Current Commit',
        }),
      ).toHaveAttribute(
        'href',
        '/lineage/default/repos/images/branch/master/commit/4a83c74809664f899261baccdb47cd90/?prevPath=%2Flineage%2Fdefault%2Frepos%2Fimages',
      );

      const previousCommits = screen.queryAllByTestId('CommitList__commit');
      expect(previousCommits).toHaveLength(2);
      expect(previousCommits[0]).toHaveTextContent(/4a83c...@master/);
      expect(previousCommits[1]).toHaveTextContent(/c43ff...@master/);
      expect(
        within(previousCommits[0]).getByRole('link', {
          name: 'Inspect Commit',
        }),
      ).toHaveAttribute(
        'href',
        '/lineage/default/repos/images/branch/master/commit/4a83c74809664f899261baccdb47cd90/?prevPath=%2Flineage%2Fdefault%2Frepos%2Fimages',
      );
      expect(
        within(previousCommits[1]).getByRole('link', {
          name: 'Inspect Commit',
        }),
      ).toHaveAttribute(
        'href',
        '/lineage/default/repos/images/branch/master/commit/c43fffd650a24b40b7d9f1bf90fcfdbe/?prevPath=%2Flineage%2Fdefault%2Frepos%2Fimages',
      );
    });

    it('should show no branches when the repo has no branches', async () => {
      server.use(
        mockRepoQuery((req, res, ctx) =>
          res(
            ctx.data({
              repo: {
                branches: [],
                createdAt: 1614426189,
                description: '',
                id: 'test',
                name: 'test',
                sizeDisplay: '0 B',
                sizeBytes: 0,
                access: true,
                projectId: 'Data-Cleaning-Process',
                authInfo: {
                  rolesList: ['repoOwner'],
                  __typename: 'AuthInfo',
                },
                __typename: 'Repo',
              },
            }),
          ),
        ),
      );

      window.history.replaceState('', '', '/lineage/default/repos/test');

      render(<Project />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('RepoDetails__repoNameSkeleton'),
      );

      const emptyMessage = await screen.findByText(
        `This repo doesn't have any branches`,
      );

      expect(emptyMessage).toBeInTheDocument();
    });

    it('should not display logs button', async () => {
      window.history.replaceState('', '', '/lineage/default/repos/montage');

      render(<Project />);
      expect(screen.queryByText('Read Logs')).not.toBeInTheDocument();
    });

    it('should disable the delete button when there are associated pipelines', async () => {
      server.use(mockGetLargerDag());
      window.history.replaceState('', '', '/lineage/default/repos/montage');

      render(<Project />);
      const deleteButton = await screen.findByTestId('DeleteRepoButton__link');
      expect(deleteButton).toBeDisabled();
    });

    it('should display a link to repo outputs', async () => {
      server.use(
        mockRepoQuery((req, res, ctx) =>
          res(
            ctx.data({
              repo: {
                branches: [
                  {
                    name: 'master',
                    __typename: 'Branch',
                  },
                ],
                createdAt: 1614126189,
                description: '',
                id: 'edges',
                name: 'edges',
                sizeDisplay: '0 B',
                sizeBytes: 0,
                access: true,
                projectId: 'Egress-Examples',
                authInfo: {
                  rolesList: ['repoOwner'],
                  __typename: 'AuthInfo',
                },
                __typename: 'Repo',
              },
            }),
          ),
        ),
      );
      window.history.replaceState(
        '',
        '',
        '/lineage/Egress-Examples/repos/edges',
      );

      render(
        <Project
          pipelineOutputsMap={{
            'Egress-Examples_edges': [
              {
                id: 'edges_output',
                name: 'edges_output',
              },
              {
                id: 'egress_output',
                name: 'egress_output',
              },
            ],
          }}
        />,
      );

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('RepoDetails__repoNameSkeleton'),
      );

      expect(screen.getByText('edges_output')).toBeInTheDocument();
      expect(screen.getByText('egress_output')).toBeInTheDocument();
    });

    it('should allow users to open the roles modal', async () => {
      server.use(
        mockRepoQuery((req, res, ctx) =>
          res(
            ctx.data({
              repo: {
                branches: [
                  {
                    name: 'master',
                    __typename: 'Branch',
                  },
                ],
                createdAt: 1614126189,
                description: '',
                id: 'edges',
                name: 'edges',
                sizeDisplay: '0 B',
                sizeBytes: 0,
                access: true,
                projectId: 'Egress-Examples',
                authInfo: {
                  rolesList: ['repoOwner'],
                  __typename: 'AuthInfo',
                },
                __typename: 'Repo',
              },
            }),
          ),
        ),
      );
      window.history.replaceState(
        '',
        '',
        '/lineage/Egress-Examples/repos/edges',
      );

      render(<Project />);
      await click((await screen.findAllByText('Set Roles'))[0]);

      expect(
        screen.getByText('Set Repo Level Roles: Egress-Examples/edges'),
      ).toBeInTheDocument();
    });

    it('should show a link to file browser for most recent commit', async () => {
      window.history.replaceState('', '', '/lineage/default/repos/montage');

      render(<Project />);

      const fileBrowserLink = await screen.findByRole('link', {
        name: 'Inspect Commits',
      });
      expect(fileBrowserLink).toHaveAttribute(
        'href',
        '/lineage/default/repos/montage/branch/master/commit/4a83c74809664f899261baccdb47cd90/?prevPath=%2Flineage%2Fdefault%2Frepos%2Fmontage',
      );
    });
  });
});
