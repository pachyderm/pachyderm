import {
  render,
  waitForElementToBeRemoved,
  screen,
} from '@testing-library/react';
import React from 'react';

import {withContextProviders, click} from '@dash-frontend/testHelpers';

import PipelineListComponent from '../PipelineList';

describe('Pipeline Steps', () => {
  const PipelineList = withContextProviders(() => {
    return <PipelineListComponent />;
  });

  beforeEach(() => {
    window.history.replaceState(
      '',
      '',
      '/project/Solar-Panel-Data-Sorting/pipelines',
    );
  });

  it('should display pipeline step details', async () => {
    render(<PipelineList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('PipelineStepsTable__loadingDots'),
    );

    const pipelines = screen.getAllByTestId('PipelineStepsList__row');
    expect(pipelines[0]).toHaveTextContent('montage');
    expect(pipelines[0]).toHaveTextContent('Error - Failure');
    expect(pipelines[0]).toHaveTextContent('Running - Created');
    expect(pipelines[0]).toHaveTextContent('v:0');
    expect(pipelines[0]).toHaveTextContent('Not my favorite pipeline');
  });

  it('should allow the user to select a subset of pipeline steps', async () => {
    render(<PipelineList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('PipelineStepsTable__loadingDots'),
    );

    expect(
      screen.getByText('Select rows to view detailed info'),
    ).toBeInTheDocument();
    expect(await screen.findAllByTestId('PipelineStepsList__row')).toHaveLength(
      2,
    );

    await click(screen.getAllByText('montage')[0]);
    await click(screen.getAllByText('edges')[0]);

    expect(
      screen.getByText('Detailed info for 2 pipeline steps'),
    ).toBeInTheDocument();
  });

  it('should allow users to view a pipeline step in the DAG', async () => {
    render(<PipelineList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('PipelineStepsTable__loadingDots'),
    );

    await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
    await click((await screen.findAllByText('View in DAG'))[0]);

    expect(window.location.pathname).toBe(
      '/lineage/Solar-Panel-Data-Sorting/pipelines/montage',
    );
  });

  it('should sort pipeline steps', async () => {
    render(<PipelineList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('PipelineStepsTable__loadingDots'),
    );

    let pipelines = screen.getAllByTestId('PipelineStepsList__row');
    expect(pipelines[0]).toHaveTextContent('montage');
    expect(pipelines[1]).toHaveTextContent('edges');

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByText('Alphabetical: A-Z'));
    pipelines = screen.getAllByTestId('PipelineStepsList__row');
    expect(pipelines[0]).toHaveTextContent('edges');
    expect(pipelines[1]).toHaveTextContent('montage');

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByText('Alphabetical: Z-A'));
    pipelines = screen.getAllByTestId('PipelineStepsList__row');
    expect(pipelines[0]).toHaveTextContent('montage');
    expect(pipelines[1]).toHaveTextContent('edges');
  });

  it('should filter by pipeline state', async () => {
    render(<PipelineList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('PipelineStepsTable__loadingDots'),
    );

    expect(await screen.findAllByTestId('PipelineStepsList__row')).toHaveLength(
      2,
    );
    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByText('Idle'));

    const pipelines = await screen.findAllByTestId('PipelineStepsList__row');
    expect(pipelines).toHaveLength(1);
    expect(pipelines[0]).toHaveTextContent('Idle - Running');
  });

  it('should filter by last job status', async () => {
    render(<PipelineList />);

    await waitForElementToBeRemoved(() =>
      screen.queryByTestId('PipelineStepsTable__loadingDots'),
    );

    expect(await screen.findAllByTestId('PipelineStepsList__row')).toHaveLength(
      2,
    );

    await click(screen.getByLabelText('expand filters'));
    await click(screen.getByRole('checkbox', {name: 'Success', hidden: false}));
    expect(screen.getByText('No matching results')).toBeInTheDocument();
  });
});
