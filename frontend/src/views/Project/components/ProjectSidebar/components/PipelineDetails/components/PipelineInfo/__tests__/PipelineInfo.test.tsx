import {
  render,
  waitForElementToBeRemoved,
  within,
  screen,
} from '@testing-library/react';
import React from 'react';
import {Route} from 'react-router';

import {withContextProviders} from '@dash-frontend/testHelpers';

import PipelineInfoComponent from '../PipelineInfo';

describe('PipelineInfo', () => {
  const PipelineInfo = withContextProviders(() => {
    return (
      <Route path="/lineage/:projectId/pipelines/:pipelineId">
        <PipelineInfoComponent />
      </Route>
    );
  });

  it('should display information about the pipeline', async () => {
    const projectId = 'Solar-Panel-Data-Sorting';
    const pipelineId = 'montage';

    window.history.replaceState(
      '',
      '',
      `/lineage/${projectId}/pipelines/${pipelineId}`,
    );

    render(<PipelineInfo />);

    await waitForElementToBeRemoved(
      screen.queryByTestId('Description__Pipeline TypeSkeleton'),
    );

    expect(screen.getByTestId('PipelineState__state')).toHaveTextContent(
      'Failure',
    );
    expect(
      screen.getByText('Pipeline Type').parentElement?.nextElementSibling,
    ).toHaveTextContent('Standard');
    expect(
      screen.getByText('Failure Reason').parentElement?.nextElementSibling,
    ).toHaveTextContent('failed');
    expect(
      screen.getByText('Description').parentElement?.nextElementSibling,
    ).toHaveTextContent('Not my favorite pipeline');

    const outputRepo =
      screen.getByText('Output Repo').parentElement?.nextElementSibling;
    expect(outputRepo).toHaveTextContent(pipelineId);
    expect(within(outputRepo as HTMLElement).getByRole('link')).toHaveAttribute(
      'href',
      `/lineage/${projectId}/repos/${pipelineId}`,
    );

    expect(
      screen.getByText('Datum Timeout').parentElement?.nextElementSibling,
    ).toHaveTextContent('N/A');
    expect(
      screen.getByText('Datum Tries').parentElement?.nextElementSibling,
    ).toHaveTextContent('0');
    expect(
      screen.getByText('Job Timeout').parentElement?.nextElementSibling,
    ).toHaveTextContent('N/A');
    expect(
      screen.getByText('Output Branch').parentElement?.nextElementSibling,
    ).toHaveTextContent('master');
    expect(
      screen.getByText('Egress').parentElement?.nextElementSibling,
    ).toHaveTextContent('Yes');
    expect(
      screen.getByText('S3 Output Repo').parentElement?.nextElementSibling,
    ).toHaveTextContent(`s3//${pipelineId}`);
  });
});
