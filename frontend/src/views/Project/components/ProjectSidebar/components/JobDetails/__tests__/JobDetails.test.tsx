import {render, waitFor, within, screen} from '@testing-library/react';
import {fromUnixTime, formatDistanceToNowStrict} from 'date-fns';
import React from 'react';
import {Route} from 'react-router';

import {
  withContextProviders,
  click,
  getUrlState,
} from '@dash-frontend/testHelpers';
import {PROJECT_JOB_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  fileBrowserRoute,
  jobRoute,
} from '@dash-frontend/views/Project/utils/routes';

import JobDetails from '..';

describe('Job Details', () => {
  const TestBed = withContextProviders(() => {
    return <Route path={PROJECT_JOB_PATH} component={JobDetails} />;
  });

  const projectId = 'Data-Cleaning-Process';
  const jobId = '23b9af7d5d4343219bc8e02ff4acd33a';

  beforeEach(() => {
    window.history.replaceState('', '', jobRoute({projectId, jobId}));
  });

  it('should display the list of pipelines that produced jobs in the set', async () => {
    render(<TestBed />);

    await waitFor(() =>
      expect(
        screen.queryByTestId('JobDetails__loading'),
      ).not.toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(
        screen.queryByTestId('Description__InputsSkeleton'),
      ).not.toBeInTheDocument(),
    );

    const withinNavList = within(screen.getByRole('list'));

    const likelihoodsLink = withinNavList.getByRole('link', {
      name: `likelihoods`,
    });
    expect(likelihoodsLink).toBeInTheDocument();
    expect(likelihoodsLink).toHaveAttribute(
      'href',
      decodeURI(jobRoute({projectId, jobId, pipelineId: 'likelihoods'}, false)),
    );

    const modelsLink = withinNavList.getByRole('link', {
      name: `models`,
    });
    expect(modelsLink).toBeInTheDocument();
    expect(modelsLink).toHaveAttribute(
      'href',
      decodeURI(jobRoute({projectId, jobId, pipelineId: 'models'}, false)),
    );

    const jointCallLink = withinNavList.getByRole('link', {
      name: `joint_call`,
    });
    expect(jointCallLink).toBeInTheDocument();
    expect(jointCallLink).toHaveAttribute(
      'href',
      decodeURI(jobRoute({projectId, jobId, pipelineId: 'joint_call'}, false)),
    );

    const splitLink = withinNavList.getByRole('link', {
      name: `split`,
    });
    expect(splitLink).toBeInTheDocument();
    expect(splitLink).toHaveAttribute(
      'href',
      decodeURI(jobRoute({projectId, jobId, pipelineId: 'split'}, false)),
    );

    const testLink = withinNavList.getByRole('link', {
      name: `test`,
    });
    expect(testLink).toBeInTheDocument();
    expect(testLink).toHaveAttribute(
      'href',
      decodeURI(jobRoute({projectId, jobId, pipelineId: 'test'}, false)),
    );

    await click(testLink);

    await waitFor(() =>
      expect(
        screen.queryByTestId('Description__InputsSkeleton'),
      ).not.toBeInTheDocument(),
    );

    expect(window.location.pathname).toBe(
      jobRoute({projectId, jobId, pipelineId: 'test'}, false),
    );
  });

  it('should default to the first job in the set if not specified', async () => {
    render(<TestBed />);

    await waitFor(() =>
      expect(
        screen.queryByTestId('JobDetails__loading'),
      ).not.toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(
        screen.queryByTestId('Description__InputsSkeleton'),
      ).not.toBeInTheDocument(),
    );

    expect(window.location.pathname).toBe(
      jobRoute({projectId, jobId, pipelineId: 'likelihoods'}, false),
    );
    await waitFor(() =>
      expect(
        screen.queryByTestId('InfoPanel__loading'),
      ).not.toBeInTheDocument(),
    );

    expect(screen.getByTestId('InfoPanel__pipeline')).toHaveTextContent(
      'likelihoods',
    );
    expect(screen.getByTestId('InfoPanel__state')).toHaveTextContent('Failure');
    expect(screen.getByTestId('RuntimeStats__started')).toHaveTextContent(
      formatDistanceToNowStrict(fromUnixTime(1614136190), {addSuffix: true}),
    );
    expect(screen.getByTestId('InfoPanel__success')).toHaveTextContent(/^0$/);
    expect(screen.getByTestId('InfoPanel__failed')).toHaveTextContent(/^100$/);
    expect(screen.getByTestId('InfoPanel__skipped')).toHaveTextContent(/^0$/);
    expect(screen.getByTestId('InfoPanel__recovered')).toHaveTextContent(/^0$/);
    expect(screen.getByTestId('InfoPanel__total')).toHaveTextContent(/^100$/);
  });

  it('should display correct pipeline job based on url', async () => {
    window.history.replaceState(
      '',
      '',
      jobRoute({projectId, jobId, pipelineId: 'models'}),
    );

    render(<TestBed />);

    await waitFor(() =>
      expect(
        screen.queryByTestId('JobDetails__loading'),
      ).not.toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(
        screen.queryByTestId('Description__InputsSkeleton'),
      ).not.toBeInTheDocument(),
    );

    const pipelineName = await screen.findByTestId('InfoPanel__pipeline');
    expect(pipelineName).toHaveTextContent('models');
  });

  it('should display job durations', async () => {
    window.history.replaceState(
      '',
      '',
      jobRoute({
        projectId: 'Solar-Panel-Data-Sorting',
        jobId: '33b9af7d5d4343219bc8e02ff44cd55a',
        pipelineId: 'montage',
      }),
    );

    render(<TestBed />);

    await waitFor(() =>
      expect(
        screen.queryByTestId('JobDetails__loading'),
      ).not.toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(
        screen.queryByTestId('Description__InputsSkeleton'),
      ).not.toBeInTheDocument(),
    );
    await click(await screen.findByTestId('RuntimeStats__duration'));

    expect(
      screen.queryByTestId('RuntimeStats__durationDetails'),
    ).toMatchSnapshot();
  });

  it('should display job failure', async () => {
    window.history.replaceState(
      '',
      '',
      jobRoute({
        projectId,
        jobId,
        pipelineId: 'likelihoods',
      }),
    );

    render(<TestBed />);

    await waitFor(() =>
      expect(
        screen.queryByTestId('JobDetails__loading'),
      ).not.toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(
        screen.queryByTestId('Description__InputsSkeleton'),
      ).not.toBeInTheDocument(),
    );
    expect(await screen.findByTestId('InfoPanel__state')).toHaveTextContent(
      'Failure',
    );
    expect(await screen.findByTestId('InfoPanel__reason')).toHaveTextContent(
      'failed',
    );
  });

  it('should correctly render extra details in JSON blob', async () => {
    window.history.replaceState(
      '',
      '',
      jobRoute({
        projectId: 'Solar-Panel-Data-Sorting',
        jobId: '33b9af7d5d4343219bc8e02ff44cd55a',
        pipelineId: 'montage',
      }),
    );

    render(<TestBed />);

    await waitFor(() =>
      expect(
        screen.queryByTestId('JobDetails__loading'),
      ).not.toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(
        screen.queryByTestId('Description__InputsSkeleton'),
      ).not.toBeInTheDocument(),
    );

    expect(screen.queryByTestId('InfoPanel__details')).toMatchSnapshot();
  });

  it('should fallback to first job if pipeline job cannot be found', async () => {
    window.history.replaceState(
      '',
      '',
      jobRoute({projectId, jobId, pipelineId: 'bogus'}),
    );

    render(<TestBed />);

    await waitFor(() =>
      expect(
        screen.queryByTestId('JobDetails__loading'),
      ).not.toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(
        screen.queryByTestId('Description__InputsSkeleton'),
      ).not.toBeInTheDocument(),
    );

    expect(window.location.pathname).toBe(
      jobRoute({projectId, jobId, pipelineId: 'likelihoods'}, false),
    );
  });

  it('should allow the user to navigate to the output commit', async () => {
    window.history.replaceState(
      '',
      '',
      jobRoute({projectId, jobId, pipelineId: 'models'}),
    );

    render(<TestBed />);

    await waitFor(() =>
      expect(
        screen.queryByTestId('JobDetails__loading'),
      ).not.toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(
        screen.queryByTestId('Description__InputsSkeleton'),
      ).not.toBeInTheDocument(),
    );
    await screen.findByTestId('InfoPanel__commitLink');

    const outputCommitLink = screen.getByTestId('InfoPanel__commitLink');

    await click(outputCommitLink);

    expect(window.location.pathname).toBe(
      fileBrowserRoute(
        {
          repoId: 'models',
          branchId: 'master',
          commitId: jobId,
          projectId,
        },
        false,
      ),
    );

    expect(getUrlState()).toMatchObject({
      prevFileBrowserPath: decodeURI(
        jobRoute({projectId, jobId, pipelineId: 'models'}, false),
      ),
    });
  });
});
