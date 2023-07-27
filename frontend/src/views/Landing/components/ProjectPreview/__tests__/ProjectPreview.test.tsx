import {mockProjectDetailsQuery} from '@graphqlTypes';
import {render, screen, within} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockEmptyProjectDetails,
  MOCK_PROJECT_DETAILS,
  MOCK_PROJECT_DETAILS_NO_JOBS,
  mockHealthyProjectStatus,
} from '@dash-frontend/mocks';
import {withContextProviders} from '@dash-frontend/testHelpers';

import ProjectPreviewComponent from '../ProjectPreview';

describe('ProjectPreview', () => {
  const server = setupServer();

  const ProjectPreview = withContextProviders(() => (
    <ProjectPreviewComponent
      project={{
        description: 'Project preview description',
        id: 'ProjectA',
      }}
    />
  ));

  beforeAll(() => {
    server.use(mockHealthyProjectStatus());
    server.listen();
  });

  afterAll(() => server.close());

  it('should not display the project details when the project is empty', async () => {
    server.use(mockEmptyProjectDetails());
    render(<ProjectPreview />);

    expect(
      await screen.findByText('Create your first repo/pipeline!'),
    ).toBeInTheDocument();
  });

  it('should display an empty state if there are no jobs', async () => {
    server.use(mockEmptyProjectDetails());
    server.use(
      mockProjectDetailsQuery((_req, res, ctx) => {
        return res(ctx.data(MOCK_PROJECT_DETAILS_NO_JOBS));
      }),
    );
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
      mockProjectDetailsQuery((_req, res, ctx) => {
        return res(ctx.data(MOCK_PROJECT_DETAILS));
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
});
