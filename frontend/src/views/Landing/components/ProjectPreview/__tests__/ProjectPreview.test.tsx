import {render, screen, within} from '@testing-library/react';
import range from 'lodash/range';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {RepoInfo, ListRepoRequest} from '@dash-frontend/api/pfs';
import {
  PipelineInfo,
  ListPipelineRequest,
  JobState,
} from '@dash-frontend/api/pps';
import {
  mockGetEnterpriseInfoInactive,
  buildJob,
  buildPipeline,
  buildRepo,
} from '@dash-frontend/mocks';
import {withContextProviders} from '@dash-frontend/testHelpers';

import ProjectPreviewComponent from '../ProjectPreview';

describe('ProjectPreview', () => {
  const server = setupServer();

  const ProjectPreview = withContextProviders(() => (
    <ProjectPreviewComponent
      project={{
        description: 'Project preview description',
        project: {name: 'ProjectA'},
      }}
    />
  ));

  beforeAll(() => server.listen());

  beforeEach(() => {
    server.use(mockGetEnterpriseInfoInactive());
    server.use(
      rest.post<ListRepoRequest, Empty, RepoInfo[]>(
        '/api/pfs_v2.API/ListRepo',
        (_req, res, ctx) => {
          return res(
            ctx.json([
              buildRepo({sizeBytesUpperBound: '3000'}),
              buildRepo({sizeBytesUpperBound: '0'}),
              buildRepo({sizeBytesUpperBound: '0'}),
            ]),
          );
        },
      ),
    );
    server.use(
      rest.post<ListPipelineRequest, Empty, PipelineInfo[]>(
        '/api/pps_v2.API/ListPipeline',
        (_req, res, ctx) => {
          return res(ctx.json([buildPipeline(), buildPipeline()]));
        },
      ),
    );
    server.use(
      rest.post('/api/pps_v2.API/ListJob', (_req, res, ctx) => {
        return res(ctx.json([]));
      }),
    );
  });

  afterAll(() => server.close());

  it('should not display the project details when the project is empty', async () => {
    server.use(
      rest.post('/api/pfs_v2.API/ListRepo', (_req, res, ctx) => {
        return res(ctx.json([]));
      }),
    );
    server.use(
      rest.post('/api/pps_v2.API/ListPipeline', (_req, res, ctx) => {
        return res(ctx.json([]));
      }),
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
});
