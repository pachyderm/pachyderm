import {render, waitFor, within} from '@testing-library/react';
import {formatDistanceToNow, fromUnixTime} from 'date-fns';
import React from 'react';
import {Route} from 'react-router';

import {withContextProviders, click} from '@dash-frontend/testHelpers';
import {JOB_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {jobRoute} from '@dash-frontend/views/Project/utils/routes';

import JobDetails from '../';

describe('Job Details', () => {
  const TestBed = withContextProviders(() => {
    return <Route path={JOB_PATH} component={JobDetails} />;
  });

  const projectId = '2';
  const jobId = '23b9af7d5d4343219bc8e02ff4acd33a';

  beforeEach(() => {
    window.history.replaceState('', '', jobRoute({projectId, jobId}));
  });

  it('should display the list of pipelines that produced jobs in the set', async () => {
    const {queryByTestId, getByRole} = render(<TestBed />);

    await waitFor(() =>
      expect(queryByTestId('JobDetails__loading')).not.toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(
        queryByTestId('Description__InputsSkeleton'),
      ).not.toBeInTheDocument(),
    );

    const withinNavList = within(getByRole('list'));

    const likelihoodsLink = withinNavList.getByRole('link', {
      name: `Pipeline job ${jobId} Failure: likelihoods`,
    });
    expect(likelihoodsLink).toBeInTheDocument();
    expect(likelihoodsLink).toHaveAttribute(
      'href',
      jobRoute({projectId, jobId, pipelineId: 'likelihoods'}, false),
    );

    const modelsLink = withinNavList.getByRole('link', {
      name: `Pipeline job ${jobId} Egressing: models`,
    });
    expect(modelsLink).toBeInTheDocument();
    expect(modelsLink).toHaveAttribute(
      'href',
      jobRoute({projectId, jobId, pipelineId: 'models'}, false),
    );

    const jointCallLink = withinNavList.getByRole('link', {
      name: `Pipeline job ${jobId} Killed: joint_call`,
    });
    expect(jointCallLink).toBeInTheDocument();
    expect(jointCallLink).toHaveAttribute(
      'href',
      jobRoute({projectId, jobId, pipelineId: 'joint_call'}, false),
    );

    const splitLink = withinNavList.getByRole('link', {
      name: `Pipeline job ${jobId} Running: split`,
    });
    expect(splitLink).toBeInTheDocument();
    expect(splitLink).toHaveAttribute(
      'href',
      jobRoute({projectId, jobId, pipelineId: 'split'}, false),
    );

    const testLink = withinNavList.getByRole('link', {
      name: `Pipeline job ${jobId} Starting: test`,
    });
    expect(testLink).toBeInTheDocument();
    expect(testLink).toHaveAttribute(
      'href',
      jobRoute({projectId, jobId, pipelineId: 'test'}, false),
    );

    click(testLink);

    await waitFor(() =>
      expect(
        queryByTestId('Description__InputsSkeleton'),
      ).not.toBeInTheDocument(),
    );

    expect(window.location.pathname).toBe(
      jobRoute({projectId, jobId, pipelineId: 'test'}, false),
    );
  });

  it('should default to the first job in the set if not specified', async () => {
    const {queryByTestId, getByTestId} = render(<TestBed />);

    await waitFor(() =>
      expect(queryByTestId('JobDetails__loading')).not.toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(
        queryByTestId('Description__InputsSkeleton'),
      ).not.toBeInTheDocument(),
    );

    expect(window.location.pathname).toBe(
      jobRoute({projectId, jobId, pipelineId: 'likelihoods'}, false),
    );

    expect(getByTestId('InfoPanel__id')).toHaveTextContent(jobId);
    expect(getByTestId('InfoPanel__pipeline')).toHaveTextContent('likelihoods');
    expect(getByTestId('InfoPanel__state')).toHaveTextContent('Failure');
    expect(getByTestId('InfoPanel__started')).toHaveTextContent(
      formatDistanceToNow(fromUnixTime(1614136189), {
        addSuffix: true,
      }),
    );
  });

  it('should display correct pipeline job based on url', async () => {
    window.history.replaceState(
      '',
      '',
      jobRoute({projectId, jobId, pipelineId: 'models'}),
    );

    const {getByTestId, queryByTestId} = render(<TestBed />);

    await waitFor(() =>
      expect(queryByTestId('JobDetails__loading')).not.toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(
        queryByTestId('Description__InputsSkeleton'),
      ).not.toBeInTheDocument(),
    );

    const pipelineName = getByTestId('InfoPanel__pipeline');
    expect(pipelineName).toHaveTextContent('models');
  });

  it('should fallback to first job if pipeline job cannot be found', async () => {
    window.history.replaceState(
      '',
      '',
      jobRoute({projectId, jobId, pipelineId: 'bogus'}),
    );

    const {queryByTestId} = render(<TestBed />);

    await waitFor(() =>
      expect(queryByTestId('JobDetails__loading')).not.toBeInTheDocument(),
    );
    await waitFor(() =>
      expect(
        queryByTestId('Description__InputsSkeleton'),
      ).not.toBeInTheDocument(),
    );

    expect(window.location.pathname).toBe(
      jobRoute({projectId, jobId, pipelineId: 'likelihoods'}, false),
    );
  });
});
