import {
  render,
  waitFor,
  waitForElementToBeRemoved,
  screen,
  within,
} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Permission} from '@dash-frontend/api/auth';
import {Empty} from '@dash-frontend/api/googleTypes';
import {
  InspectPipelineRequest,
  ListJobRequest,
  JobInfo,
  JobState,
  PipelineInfo,
} from '@dash-frontend/api/pps';
import {
  mockEmptyGetRoles,
  mockDiffFile,
  mockGetImageCommits,
  mockRepoImages,
  mockRepoEdges,
  mockEmptyCommits,
  mockGetEnterpriseInfoInactive,
  mockPipelines,
  mockRepos,
  mockGetMontagePipeline,
  mockGetMontageJob5C,
  mockGetMontageJob1D,
  mockInspectJobMontage5C,
  buildJob,
  buildPipeline,
  mockRepoMontage,
  mockGetServicePipeline,
  mockGetSpoutPipeline,
  mockEmptyJob,
  mockFalseGetAuthorize,
  mockTrueGetAuthorize,
  mockGetImageCommitsNoBranch,
  mockGetVersionInfo,
} from '@dash-frontend/mocks';
import {click, withContextProviders} from '@dash-frontend/testHelpers';

import ProjectSidebar from '../ProjectSidebar';

describe('ProjectSidebar', () => {
  const server = setupServer();

  const Project = withContextProviders(ProjectSidebar);

  beforeAll(() => server.listen());

  beforeEach(() => {
    server.resetHandlers();
    server.use(mockGetVersionInfo());
    server.use(mockPipelines());
    server.use(mockRepos());
    server.use(mockGetMontagePipeline());
    server.use(mockGetMontageJob5C());
    server.use(mockInspectJobMontage5C());
    server.use(
      mockTrueGetAuthorize([
        Permission.REPO_MODIFY_BINDINGS,
        Permission.REPO_WRITE,
        Permission.REPO_READ,
      ]),
    );
    server.use(mockRepoImages());
    server.use(mockEmptyGetRoles());
    server.use(mockGetEnterpriseInfoInactive());
    server.use(mockGetImageCommits());
  });

  afterAll(() => server.close());

  it('should not display the sidebar if not on a sidebar route', async () => {
    window.history.replaceState('', '', '/project/default');

    render(<Project />);

    expect(
      screen.queryByTestId('ProjectSidebar__sidebar'),
    ).not.toBeInTheDocument();
  });

  describe('PipelineDetails', () => {
    it('should display top pipeline details', async () => {
      window.history.replaceState('', '', '/lineage/default/pipelines/montage');

      render(<Project />);

      expect(
        await screen.findByRole('heading', {
          name: 'montage',
        }),
      ).toBeInTheDocument();

      expect(
        screen.getAllByText(
          /a pipeline that combines images from the `images` and `edges` repositories into a montage\./i,
        ),
      ).toHaveLength(3);

      expect(
        screen.getByRole('definition', {
          name: /most recent job id/i,
        }),
      ).toHaveTextContent('5c1aa9bc87dd411ba5a1be0c80a3ebc2');
    });

    it('should display the job overview tab', async () => {
      window.history.replaceState('', '', '/lineage/default/pipelines/montage');

      render(<Project />);

      const overviewTab = await screen.findByRole('tabpanel', {
        name: /job overview/i,
      });

      expect(
        within(overviewTab).getByRole('heading', {name: /success/i}),
      ).toBeInTheDocument();

      expect(
        within(overviewTab).getByRole('link', {name: /inspect job/i}),
      ).toHaveAttribute(
        'href',
        '/lineage/default/pipelines/montage/jobs/5c1aa9bc87dd411ba5a1be0c80a3ebc2/logs/datum?prevPath=%2Flineage%2Fdefault%2Fpipelines%2Fmontage',
      );

      expect(
        within(overviewTab).getByRole('definition', {
          name: /created/i,
        }),
      ).toHaveTextContent('Aug 1, 2023; 14:20');

      expect(
        within(overviewTab).getByRole('definition', {
          name: /start/i,
        }),
      ).toHaveTextContent('Aug 1, 2023; 14:20');

      const runtimeDropdown = within(overviewTab).getByRole('button', {
        name: /cumulative time/i,
      });
      expect(runtimeDropdown).toHaveTextContent('3 s');
      await click(runtimeDropdown);

      expect(
        within(overviewTab).getByRole('definition', {
          name: /download/i,
        }),
      ).toHaveTextContent('1 s');

      expect(
        within(overviewTab).getByRole('definition', {
          name: /processing/i,
        }),
      ).toHaveTextContent('1 s');

      expect(
        within(overviewTab).getByRole('definition', {
          name: /upload/i,
        }),
      ).toHaveTextContent('1 s');

      expect(
        within(overviewTab).getByRole('link', {name: /images/i}),
      ).toHaveAttribute('href', '/lineage/default/repos/images');

      expect(
        await within(overviewTab).findByRole('link', {name: /edges/i}),
      ).toHaveAttribute('href', '/lineage/default/repos/edges');

      expect(
        within(overviewTab).getByRole('link', {name: /montage/i}),
      ).toHaveAttribute('href', '/lineage/default/pipelines/montage');

      expect(
        within(overviewTab).getByRole('link', {
          name: /5c1aa9bc87dd411ba5a1be0c80a3ebc2/i,
        }),
      ).toHaveAttribute(
        'href',
        '/lineage/default/repos/montage/commit/5c1aa9bc87dd411ba5a1be0c80a3ebc2/?prevPath=%2Flineage%2Fdefault%2Fpipelines%2Fmontage',
      );

      expect(within(overviewTab).getAllByText(/"joinOn"/)).toHaveLength(2);
    });

    it('should display a specific job overview when a global id filter is applied', async () => {
      server.use(mockGetMontageJob1D());
      window.history.replaceState(
        '',
        '',
        '/lineage/default/pipelines/montage?globalIdFilter=1dc67e479f03498badcc6180be4ee6ce',
      );

      render(<Project />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      expect(
        screen.getByRole('definition', {name: 'Global ID'}),
      ).toBeInTheDocument();

      expect(
        screen.getByRole('link', {name: /previous subjobs/i}),
      ).toHaveAttribute(
        'href',
        '/lineage/default/pipelines/montage/jobs/1dc67e479f03498badcc6180be4ee6ce/logs?globalIdFilter=1dc67e479f03498badcc6180be4ee6ce&prevPath=%2Flineage%2Fdefault%2Fpipelines%2Fmontage%3FglobalIdFilter%3D1dc67e479f03498badcc6180be4ee6ce',
      );

      const overviewTab = await screen.findByRole('tabpanel', {
        name: /job overview/i,
      });

      expect(
        within(overviewTab).getByRole('heading', {name: /killed/i}),
      ).toBeInTheDocument();

      expect(
        within(overviewTab).getByRole('link', {name: /inspect job/i}),
      ).toHaveAttribute(
        'href',
        '/lineage/default/pipelines/montage/jobs/1dc67e479f03498badcc6180be4ee6ce/logs/datum?globalIdFilter=1dc67e479f03498badcc6180be4ee6ce&prevPath=%2Flineage%2Fdefault%2Fpipelines%2Fmontage%3FglobalIdFilter%3D1dc67e479f03498badcc6180be4ee6ce',
      );
    });

    it('should hide the spec tab when a global id filter is applied', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/default/pipelines/montage?globalIdFilter=5c1aa9bc87dd411ba5a1be0c80a3ebc2',
      );

      render(<Project />);

      expect(
        await screen.findByRole('heading', {
          name: 'montage',
        }),
      ).toBeInTheDocument();

      expect(
        screen.queryByRole('tab', {name: /spec/i}),
      ).not.toBeInTheDocument();
    });

    it('should display effective specs in the pipeline spec tab', async () => {
      const spy = jest.spyOn(document, 'createElement');
      window.history.replaceState('', '', '/lineage/default/pipelines/montage');

      render(<Project />);

      await click(await screen.findByRole('tab', {name: /spec/i}));

      const specTab = screen.getByRole('tabpanel', {name: /spec/i});
      const effectiveSpec = within(specTab).getByLabelText('Effective Spec');

      expect(
        within(effectiveSpec).getByText(/resourceRequests/),
      ).toBeInTheDocument();

      await click(within(effectiveSpec).getByLabelText('Minimize Spec'));

      expect(
        within(effectiveSpec).queryByText(/resourceRequests/),
      ).not.toBeInTheDocument();

      await click(within(effectiveSpec).getByLabelText('Maximize Spec'));

      // gear icon button
      await click(
        within(effectiveSpec).getByLabelText('Pipeline Spec Options'),
      );

      await click(within(effectiveSpec).getByRole('menuitem', {name: 'Copy'}));
      expect(navigator.clipboard.writeText).toHaveBeenLastCalledWith(
        `pipeline:
  project:
    name: default
  name: montage
transform:
  image: dpokidov/imagemagick:7.1.0-23
  cmd:
    - sh
  stdin:
    - >-
      montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find -L
      /pfs/images /pfs/edges -type f | sort) /pfs/out/montage.png
input:
  cross:
    - pfs:
        repo: images
        glob: /
    - pfs:
        repo: edges
        glob: /
description: >-
  A pipeline that combines images from the \`images\` and \`edges\` repositories
  into a montage.
resourceRequests:
  cpu: 1
  disk: 1Gi
  memory: 256Mi
sidecarResourceRequests:
  cpu: 1
  disk: 1Gi
  memory: 256Mi
`,
      );

      await click(
        within(effectiveSpec).getByLabelText('Pipeline Spec Options'),
      );
      await click(
        within(effectiveSpec).getByRole('menuitem', {name: 'Download JSON'}),
      );

      // test that `useDownloadText` created an anchor tag
      expect(spy).toHaveBeenCalledWith('a');

      await click(
        within(effectiveSpec).getByLabelText('Pipeline Spec Options'),
      );
      await click(
        within(effectiveSpec).getByRole('menuitem', {name: 'Download YAML'}),
      );

      expect(spy).toHaveBeenCalledWith('a');

      expect(
        screen.getAllByTestId('dynamicEffectiveSpecDecorations__userAvatarSVG'),
      ).toHaveLength(7);
    });

    it('should display user specs in the pipeline spec tab', async () => {
      const spy = jest.spyOn(document, 'createElement');
      window.history.replaceState('', '', '/lineage/default/pipelines/montage');

      render(<Project />);

      await click(await screen.findByRole('tab', {name: /spec/i}));

      const specTab = screen.getByRole('tabpanel', {name: /spec/i});
      const effectiveSpec = within(specTab).getByLabelText('Submitted Spec');

      expect(
        within(effectiveSpec).queryByText(/resourceRequests/),
      ).not.toBeInTheDocument();
      expect(within(effectiveSpec).getByText(/transform/)).toBeInTheDocument();

      await click(within(effectiveSpec).getByLabelText('Minimize Spec'));

      expect(
        within(effectiveSpec).queryByText(/transform/),
      ).not.toBeInTheDocument();

      await click(within(effectiveSpec).getByLabelText('Maximize Spec'));

      // gear icon button
      await click(
        within(effectiveSpec).getByLabelText('Pipeline Spec Options'),
      );

      await click(within(effectiveSpec).getByRole('menuitem', {name: 'Copy'}));
      expect(navigator.clipboard.writeText).toHaveBeenLastCalledWith(
        `pipeline:
  project:
    name: default
  name: montage
transform:
  image: dpokidov/imagemagick:7.1.0-23
  cmd:
    - sh
  stdin:
    - >-
      montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find -L
      /pfs/images /pfs/edges -type f | sort) /pfs/out/montage.png
input:
  cross:
    - pfs:
        repo: images
        glob: /
    - pfs:
        repo: edges
        glob: /
description: >-
  A pipeline that combines images from the \`images\` and \`edges\` repositories
  into a montage.
`,
      );

      await click(
        within(effectiveSpec).getByLabelText('Pipeline Spec Options'),
      );
      await click(
        within(effectiveSpec).getByRole('menuitem', {name: 'Download JSON'}),
      );

      // test that `useDownloadText` created an anchor tag
      expect(spy).toHaveBeenLastCalledWith('a');

      await click(
        within(effectiveSpec).getByLabelText('Pipeline Spec Options'),
      );
      await click(
        within(effectiveSpec).getByRole('menuitem', {name: 'Download YAML'}),
      );

      expect(spy).toHaveBeenLastCalledWith('a');
    });

    it('should display pipeline logs button', async () => {
      window.history.replaceState('', '', '/lineage/default/pipelines/montage');

      render(<Project />);
      const topLogsLink = await screen.findByRole('link', {
        name: 'Previous Subjobs',
      });
      expect(topLogsLink).toHaveAttribute(
        'href',
        '/lineage/default/pipelines/montage/jobs/5c1aa9bc87dd411ba5a1be0c80a3ebc2/logs?prevPath=%2Flineage%2Fdefault%2Fpipelines%2Fmontage',
      );
      expect(topLogsLink).toBeEnabled();

      const inspectLogsLink = await screen.findByRole('link', {
        name: 'Inspect Job',
      });
      expect(inspectLogsLink).toHaveAttribute(
        'href',
        '/lineage/default/pipelines/montage/jobs/5c1aa9bc87dd411ba5a1be0c80a3ebc2/logs/datum?prevPath=%2Flineage%2Fdefault%2Fpipelines%2Fmontage',
      );
      expect(inspectLogsLink).toBeEnabled();
    });

    it('should display datum logs link with filter applied', async () => {
      window.history.replaceState('', '', '/lineage/default/pipelines/montage');
      render(<Project />);

      const logsLink = await screen.findByRole('link', {name: '1 Success'});
      expect(logsLink).toHaveAttribute(
        'href',
        '/lineage/default/pipelines/montage/jobs/5c1aa9bc87dd411ba5a1be0c80a3ebc2/logs/datum?prevPath=%2Flineage%2Fdefault%2Fpipelines%2Fmontage&datumFilters=SUCCESS',
      );
    });

    it('should show a linked project input node', async () => {
      server.use(
        rest.post<InspectPipelineRequest, Empty, PipelineInfo>(
          '/api/pps_v2.API/InspectPipeline',
          (req, res, ctx) => {
            return res(ctx.json(buildPipeline()));
          },
        ),
      );
      server.use(
        rest.post<ListJobRequest, Empty, JobInfo[]>(
          '/api/pps_v2.API/ListJob',
          (req, res, ctx) => {
            return res(
              ctx.json([
                buildJob({
                  job: {
                    id: '23b9af7d5d4343219bc8e02ff44cd55a',
                    pipeline: {
                      name: 'Node_2',
                      project: {
                        name: 'Multi-Project-Pipeline-A',
                      },
                    },
                  },
                  details: {
                    input: {
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
                    },
                  },
                }),
              ]),
            );
          },
        ),
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

    it('should allow users to stop a running job', async () => {
      server.use(
        rest.post<ListJobRequest, Empty, JobInfo[]>(
          '/api/pps_v2.API/ListJob',
          async (_req, res, ctx) => {
            return res(
              ctx.json([
                buildJob({
                  state: JobState.JOB_RUNNING,
                }),
              ]),
            );
          },
        ),
      );
      window.history.replaceState('', '', '/lineage/default/pipelines/montage');

      render(<Project />);

      const overviewTab = await screen.findByRole('tabpanel', {
        name: /job overview/i,
      });

      expect(
        within(overviewTab).getByRole('heading', {name: /running/i}),
      ).toBeInTheDocument();

      expect(
        within(overviewTab).getByRole('button', {name: /stop job/i}),
      ).toBeInTheDocument();
    });

    it('should not allow users to stop a finished job', async () => {
      window.history.replaceState('', '', '/lineage/default/pipelines/montage');

      render(<Project />);

      const overviewTab = await screen.findByRole('tabpanel', {
        name: /job overview/i,
      });

      expect(
        within(overviewTab).getByRole('heading', {name: /success/i}),
      ).toBeInTheDocument();

      expect(
        within(overviewTab).queryByRole('button', {name: /stop job/i}),
      ).not.toBeInTheDocument();
    });

    it('should allow users to open the roles modal', async () => {
      server.use(mockRepoMontage());

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

    it('should show service ip for a service pipeline of type LoadBalancer', async () => {
      server.use(mockGetServicePipeline());

      window.history.replaceState({}, '', '/lineage/default/pipelines/montage');
      render(<Project />);

      const field = await screen.findByRole('definition', {
        name: /service ip/i,
      });

      expect(field).toHaveTextContent(/localhost/i);

      expect(within(field).getByRole('link')).toHaveAttribute(
        'href',
        'http://localhost:80',
      );
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
      expect(screen.queryByText('Most Recent Job ID')).not.toBeInTheDocument();
    });

    it('should hide job info if no access and display empty state', async () => {
      server.use(mockFalseGetAuthorize());
      server.use(mockEmptyJob());

      window.history.replaceState({}, '', '/lineage/default/pipelines/montage');
      render(<Project />);

      const overviewTab = await screen.findByRole('tabpanel', {
        name: /job overview/i,
      });

      expect(
        screen.getByRole('button', {
          name: /previous subjobs/i,
        }),
      ).toBeDisabled();

      expect(
        within(overviewTab).getByRole('heading', {
          name: "You don't have permission to view this pipeline",
        }),
      ).toBeInTheDocument();

      expect(
        within(overviewTab).getByText(
          "You'll need a role of repoReader or higher on the connected repo to view job data about this pipeline.",
        ),
      ).toBeInTheDocument();

      expect(
        within(overviewTab).getByRole('link', {
          name: 'Read more about authorization',
        }),
      ).toBeInTheDocument();

      expect(
        screen.queryByRole('definition', {
          name: /most recent job id/i,
        }),
      ).not.toBeInTheDocument();
    });
  });

  describe('repos', () => {
    beforeEach(() => {
      server.use(mockDiffFile());
    });

    it('should display repo details', async () => {
      window.history.replaceState('', '', '/lineage/default/repos/images');

      render(<Project />);

      await screen.findByRole('heading', {name: 'images'});

      expect(
        screen.getByRole('definition', {
          name: /commit message/i,
        }),
      ).toHaveTextContent('added mako');

      expect(
        screen.getByRole('definition', {
          name: /repo created/i,
        }),
      ).toHaveTextContent('Jul 24, 2023; 17:58');

      expect(
        screen.getByRole('definition', {
          name: /most recent commit start/i,
        }),
      ).toHaveTextContent('Jul 24, 2023; 17:58');

      expect(await screen.findByText('139.24 kB')).toBeInTheDocument();
      await screen.findByText('4a83c74809664f899261baccdb47cd90');
      expect(screen.getByText('+ 58.65 kB')).toBeInTheDocument();
      await screen.findByText('4eb1aa...@master');

      await waitFor(() =>
        expect(
          screen.queryByTestId('CommitList__loadingdots'),
        ).not.toBeInTheDocument(),
      );

      expect(
        screen.getAllByRole('link', {
          name: 'Inspect Commit',
        })[0],
      ).toHaveAttribute(
        'href',
        '/lineage/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/?prevPath=%2Flineage%2Fdefault%2Frepos%2Fimages',
      );

      const previousCommits = screen.queryAllByTestId('CommitList__commit');
      expect(previousCommits).toHaveLength(2);
      expect(previousCommits[0]).toHaveTextContent(/4eb1aa...@master/);
      expect(previousCommits[0]).toHaveTextContent('Jul 24, 2023; 17:58');
      expect(previousCommits[1]).toHaveTextContent(/c43fff...@master/);
      expect(previousCommits[1]).toHaveTextContent('Jul 24, 2023; 17:58');
      expect(
        within(previousCommits[0]).queryByRole('link', {
          name: 'Inspect Commit',
        }),
      ).not.toBeInTheDocument();
      expect(
        within(previousCommits[1]).getByRole('link', {
          name: 'Inspect Commit',
        }),
      ).toHaveAttribute(
        'href',
        '/lineage/default/repos/images/commit/c43fffd650a24b40b7d9f1bf90fcfdbe/?prevPath=%2Flineage%2Fdefault%2Frepos%2Fimages',
      );
    });

    it('should display repo details for commit without branch', async () => {
      window.history.replaceState('', '', '/lineage/default/repos/images');

      server.use(mockGetImageCommitsNoBranch());

      render(<Project />);

      await screen.findByRole('heading', {name: 'images'});

      expect(
        screen.getByRole('definition', {
          name: /commit message/i,
        }),
      ).toHaveTextContent('I deleted this branch');

      expect(
        screen.getByRole('definition', {
          name: /repo created/i,
        }),
      ).toHaveTextContent('Jul 24, 2023; 17:58');

      expect(
        screen.getByRole('definition', {
          name: /most recent commit start/i,
        }),
      ).toHaveTextContent('Jul 24, 2023; 17:58');

      expect(await screen.findByText('139.24 kB')).toBeInTheDocument();
      await screen.findByText('g2bb3e50cd124b76840145a8c18f8892');
      expect(screen.getByText('+ 58.65 kB')).toBeInTheDocument();
      await screen.findByText('73fe17...');

      await waitFor(() =>
        expect(
          screen.queryByTestId('CommitList__loadingdots'),
        ).not.toBeInTheDocument(),
      );

      expect(
        screen.getAllByRole('link', {
          name: 'Inspect Commit',
        })[0],
      ).toHaveAttribute(
        'href',
        '/lineage/default/repos/images/commit/g2bb3e50cd124b76840145a8c18f8892/?prevPath=%2Flineage%2Fdefault%2Frepos%2Fimages',
      );
      const previousCommits = screen.queryAllByTestId('CommitList__commit');
      expect(previousCommits).toHaveLength(1);
      expect(previousCommits[0]).toHaveTextContent(/73fe17.../);
      expect(previousCommits[0]).toHaveTextContent('Jul 24, 2023; 17:58');
    });

    it('should display specific commit details when a global id filter is applied', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/default/repos/images?globalIdFilter=c43fffd650a24b40b7d9f1bf90fcfdbe',
      );

      render(<Project />);

      expect(
        await screen.findByRole('heading', {name: 'images'}),
      ).toBeInTheDocument();
      expect(
        screen.getByText('c43fffd650a24b40b7d9f1bf90fcfdbe'),
      ).toBeInTheDocument();
      expect(
        screen.getByRole('definition', {name: 'Global ID'}),
      ).toBeInTheDocument();
      expect(
        screen.getByRole('definition', {name: 'Global ID Commit Start'}),
      ).toBeInTheDocument();
      expect(await screen.findByText('+ 58.65 kB')).toBeInTheDocument();

      expect(
        screen.getByRole('link', {
          name: 'Previous Commits',
        }),
      ).toHaveAttribute(
        'href',
        '/lineage/default/repos/images/commit/c43fffd650a24b40b7d9f1bf90fcfdbe/?globalIdFilter=c43fffd650a24b40b7d9f1bf90fcfdbe&prevPath=%2Flineage%2Fdefault%2Frepos%2Fimages%3FglobalIdFilter%3Dc43fffd650a24b40b7d9f1bf90fcfdbe',
      );

      expect(
        screen.getByRole('link', {
          name: 'Inspect Commit',
        }),
      ).toHaveAttribute(
        'href',
        '/lineage/default/repos/images/commit/c43fffd650a24b40b7d9f1bf90fcfdbe/?globalIdFilter=c43fffd650a24b40b7d9f1bf90fcfdbe&prevPath=%2Flineage%2Fdefault%2Frepos%2Fimages%3FglobalIdFilter%3Dc43fffd650a24b40b7d9f1bf90fcfdbe',
      );

      expect(screen.queryAllByTestId('CommitList__commit')).toHaveLength(0);
    });

    it('should show no data when the repo has no data', async () => {
      server.use(mockEmptyCommits());

      window.history.replaceState('', '', '/lineage/default/repos/images');

      render(<Project />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('RepoDetails__repoNameSkeleton'),
      );

      expect(
        await screen.findByText(`This repo doesn't have any data`),
      ).toBeInTheDocument();

      const docsLink = (
        await screen.findByText('View our documentation about managing data')
      ).closest('a');

      expect(docsLink).toHaveAttribute(
        'href',
        'https://docs.pachyderm.com/0.0.x/prepare-data/ingest-data/',
      );
      expect(docsLink).toHaveAttribute('target', '_blank');
      expect(docsLink).toHaveAttribute('rel', 'noopener noreferrer');
    });

    it('should not display logs button', async () => {
      window.history.replaceState('', '', '/lineage/default/repos/montage');

      render(<Project />);
      expect(screen.queryByText('Read Logs')).not.toBeInTheDocument();
    });

    it('should display a link to repo outputs', async () => {
      server.use(mockRepoEdges());
      server.use(mockEmptyCommits());
      window.history.replaceState('', '', '/lineage/default/repos/edges');

      render(
        <Project
          pipelineOutputsMap={{
            d0e1e9a51269508c3f11c0e64c721c3ea6204838: [
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
      window.history.replaceState('', '', '/lineage/default/repos/images');

      render(<Project />);
      await click((await screen.findAllByText('Set Roles'))[0]);

      expect(
        screen.getByText('Set Repo Level Roles: default/images'),
      ).toBeInTheDocument();
    });

    it('should show a link to file browser for most recent commit', async () => {
      window.history.replaceState('', '', '/lineage/default/repos/images');

      render(<Project />);

      const fileBrowserLink = await screen.findByRole('link', {
        name: 'Previous Commits',
      });
      expect(fileBrowserLink).toHaveAttribute(
        'href',
        '/lineage/default/repos/images/commit/4a83c74809664f899261baccdb47cd90/?prevPath=%2Flineage%2Fdefault%2Frepos%2Fimages',
      );
    });

    it('should hide commit info if no access and display empty state', async () => {
      server.use(mockFalseGetAuthorize());
      server.use(mockEmptyCommits());

      window.history.replaceState('', '', '/lineage/default/repos/images');
      render(<Project />);

      await screen.findByRole('heading', {name: 'images'});

      expect(
        screen.getByRole('button', {
          name: 'Previous Commits',
        }),
      ).toBeDisabled();

      expect(
        screen.getByRole('button', {
          name: 'See All Roles',
        }),
      ).toBeEnabled();

      expect(
        screen.getByRole('heading', {
          name: "You don't have permission to view this repo",
        }),
      ).toBeInTheDocument();

      expect(
        screen.getByText(
          "You'll need a role of repoReader or higher to view commit data about this repo.",
        ),
      ).toBeInTheDocument();

      expect(
        screen.getByRole('link', {
          name: 'Read more about authorization',
        }),
      ).toBeInTheDocument();

      expect(screen.queryAllByTestId('CommitList__commit')).toHaveLength(0);
      expect(
        screen.queryByRole('link', {
          name: 'Previous Commits',
        }),
      ).not.toBeInTheDocument();
    });
  });
});
