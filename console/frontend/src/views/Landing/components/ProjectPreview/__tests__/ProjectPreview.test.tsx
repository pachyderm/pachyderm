import {render, screen, within} from '@testing-library/react';
import range from 'lodash/range';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  ReposSummaryRequest,
  ReposSummaryResponse,
} from '@dash-frontend/api/pfs';
import {
  JobState,
  PipelinesSummaryRequest,
  PipelinesSummaryResponse,
} from '@dash-frontend/api/pps';
import {
  mockGetEnterpriseInfoInactive,
  buildJob,
  mockRepoSummaries,
  mockPipelineSummaries,
  mockGetVersionInfo,
} from '@dash-frontend/mocks';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import ProjectPreviewComponent from '../ProjectPreview';

describe('ProjectPreview', () => {
  const server = setupServer();

  const ProjectPreview = withContextProviders(() => (
    <ProjectPreviewComponent
      project={{
        description: 'Project preview description',
        project: {name: 'ProjectA'},
        metadata: {metadataProjectKey: 'metadataProjectValue'},
      }}
      allProjectNames={['ProjectA']}
    />
  ));

  beforeAll(() => server.listen());

  beforeEach(() => {
    server.use(mockGetVersionInfo());
    server.use(mockGetEnterpriseInfoInactive());
    server.use(mockRepoSummaries());
    server.use(mockPipelineSummaries());
    server.use(
      rest.post('/api/pps_v2.API/ListJob', (_req, res, ctx) => {
        return res(ctx.json([]));
      }),
    );
  });

  afterAll(() => server.close());

  it('should not display the project details when the project is empty', async () => {
    server.use(
      rest.post<ReposSummaryRequest, Empty, ReposSummaryResponse>(
        '/api/pfs_v2.API/ReposSummary',
        (_req, res, ctx) => {
          return res(
            ctx.json({
              summaries: [
                {
                  project: {name: 'ProjectA'},
                  userRepoCount: '0',
                  sizeBytes: '0',
                },
              ],
            }),
          );
        },
      ),
    );
    server.use(
      rest.post<PipelinesSummaryRequest, Empty, PipelinesSummaryResponse>(
        '/api/pps_v2.API/PipelinesSummary',
        (_req, res, ctx) => {
          return res(
            ctx.json({
              summaries: [],
            }),
          );
        },
      ),
    );

    render(<ProjectPreview />);

    expect(
      await screen.findByText('Create your first repo/pipeline!'),
    ).toBeInTheDocument();
  });

  it('should display an empty state if there are no jobs', async () => {
    render(<ProjectPreview />);

    expect(await screen.findByText('3/2')).toBeInTheDocument();
    expect(screen.getByText('3 kB')).toBeInTheDocument();
    expect(screen.getByText('Healthy')).toBeInTheDocument();
    expect(screen.queryByRole('list')).not.toBeInTheDocument();
    expect(screen.getByText('No Jobs Found')).toBeInTheDocument();
    expect(
      screen.getByText(
        'Create your first job! If there are any pipeline errors, fix those before you create a job.',
      ),
    ).toBeInTheDocument();
  });

  it('should display the project details', async () => {
    server.use(
      rest.post('/api/pps_v2.API/ListJob', (_req, res, ctx) => {
        return res(
          ctx.json([
            buildJob({
              job: {id: '23b9af7d5d4343219bc8e02ff44cd55a'},
              state: JobState.JOB_SUCCESS,
              created: '2023-07-24T17:52:14.000Z',
            }),
            buildJob({
              job: {
                id: '33b9af7d5d4343219bc8e02ff44cd55a',
              },
              state: JobState.JOB_FAILURE,
              created: '2023-07-24T17:54:24.000Z',
            }),
            buildJob({
              job: {id: '7798fhje5d4343219bc8e02ff4acd33a'},
              state: JobState.JOB_FINISHING,
              created: '2023-07-24T17:56:34.000Z',
            }),
            buildJob({
              job: {id: 'o90du4js5d4343219bc8e02ff4acd33a'},
              state: JobState.JOB_KILLED,
              created: '2023-07-24T17:58:44.000Z',
            }),
          ]),
        );
      }),
    );
    render(<ProjectPreview />);

    expect(await screen.findByText('3/2')).toBeInTheDocument();
    expect(screen.getByText('3 kB')).toBeInTheDocument();
    expect(screen.getByText('Healthy')).toBeInTheDocument();
    expect(screen.getByText('Last 4 Jobs')).toBeInTheDocument();

    const jobsList = screen.getByRole('list');
    expect(within(jobsList).getAllByRole('listitem')).toHaveLength(4);
    expect(
      within(jobsList).getByRole('link', {name: /Success Created/i}),
    ).toHaveAttribute(
      'href',
      '/project/ProjectA/jobs/subjobs?selectedJobs=23b9af7d5d4343219bc8e02ff44cd55a',
    );
    expect(
      within(jobsList).getByRole('link', {name: /Failure Created/i}),
    ).toHaveAttribute(
      'href',
      '/project/ProjectA/jobs/subjobs?selectedJobs=33b9af7d5d4343219bc8e02ff44cd55a',
    );
    expect(
      within(jobsList).getByRole('link', {name: /Finishing Created/i}),
    ).toHaveAttribute(
      'href',
      '/project/ProjectA/jobs/subjobs?selectedJobs=7798fhje5d4343219bc8e02ff4acd33a',
    );
    expect(
      within(jobsList).getByRole('link', {name: /Killed Created/i}),
    ).toHaveAttribute(
      'href',
      '/project/ProjectA/jobs/subjobs?selectedJobs=o90du4js5d4343219bc8e02ff4acd33a',
    );
  });

  it('should only show 10 jobs', async () => {
    server.use(
      rest.post('/api/pps_v2.API/ListJob', (_req, res, ctx) => {
        const jobs = range(0, 50).map((i) =>
          buildJob({job: {id: i.toString()}}),
        );

        return res(ctx.json(jobs));
      }),
    );
    render(<ProjectPreview />);

    expect(await screen.findByText('3/2')).toBeInTheDocument();
    expect(screen.getByText('Last 10 Jobs')).toBeInTheDocument();
  });

  describe('user metadata tab', () => {
    it('should display project metadata', async () => {
      render(<ProjectPreview />);
      expect(await screen.findByText('3/2')).toBeInTheDocument();

      await click(screen.getByRole('tab', {name: /user metadata/i}));

      expect(
        await screen.findByRole('cell', {name: 'metadataProjectValue'}),
      ).toBeInTheDocument();
      expect(
        screen.getByRole('cell', {name: 'metadataProjectKey'}),
      ).toBeInTheDocument();
    });

    it('should allow users to open the edit metadata modal to add and remove values', async () => {
      render(<ProjectPreview />);
      expect(await screen.findByText('3/2')).toBeInTheDocument();

      await click(screen.getByRole('tab', {name: /user metadata/i}));
      await click(screen.getAllByRole('button', {name: /edit/i})[0]);
      const modal = await screen.findByRole('dialog');

      expect(
        await within(modal).findByRole('heading', {
          name: 'Edit Project Metadata',
        }),
      ).toBeInTheDocument();
      expect(within(modal).getAllByRole('row')).toHaveLength(3);

      await click(screen.getByRole('button', {name: /add new/i}));
      await click(screen.getByRole('button', {name: /add new/i}));

      expect(within(modal).getAllByRole('row')).toHaveLength(5);

      await click(screen.getByRole('button', {name: /delete metadata row 2/i}));

      expect(within(modal).getAllByRole('row')).toHaveLength(4);
    });
  });
});
