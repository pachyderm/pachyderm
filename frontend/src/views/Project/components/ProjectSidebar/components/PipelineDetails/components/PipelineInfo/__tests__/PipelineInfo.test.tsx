import {
  render,
  waitForElementToBeRemoved,
  within,
} from '@testing-library/react';
import React from 'react';
import {Route} from 'react-router';

import {withContextProviders} from '@dash-frontend/testHelpers';

import PipelineInfoComponent from '../PipelineInfo';

describe('PipelineInfo', () => {
  const PipelineInfo = withContextProviders(() => {
    return (
      <Route path="/project/:projectId/pipelines/:pipelineId">
        <PipelineInfoComponent />
      </Route>
    );
  });

  it('should display information about the pipeline', async () => {
    const projectId = '1';
    const pipelineId = 'montage';

    window.history.replaceState(
      '',
      '',
      `/project/${projectId}/pipelines/${pipelineId}`,
    );

    const {queryByTestId, getByText} = render(<PipelineInfo />);

    await waitForElementToBeRemoved(
      queryByTestId('Description__Pipeline StatusSkeleton'),
    );

    expect(
      getByText('Pipeline Status').parentElement?.nextElementSibling,
    ).toHaveTextContent('Failure');
    expect(
      getByText('Pipeline Type').parentElement?.nextElementSibling,
    ).toHaveTextContent('Standard');
    expect(
      getByText('Failure Reason').parentElement?.nextElementSibling,
    ).toHaveTextContent('failed');
    expect(
      getByText('Description').parentElement?.nextElementSibling,
    ).toHaveTextContent('Not my favorite pipeline');

    const outputRepo =
      getByText('Output Repo').parentElement?.nextElementSibling;
    expect(outputRepo).toHaveTextContent(pipelineId);
    expect(within(outputRepo as HTMLElement).getByRole('link')).toHaveAttribute(
      'href',
      `/project/${projectId}/repos/${pipelineId}/branch/default`,
    );

    expect(
      getByText('Datum Timeout').parentElement?.nextElementSibling,
    ).toHaveTextContent('N/A');
    expect(
      getByText('Datum Tries').parentElement?.nextElementSibling,
    ).toHaveTextContent('0');
    expect(
      getByText('Job Timeout').parentElement?.nextElementSibling,
    ).toHaveTextContent('N/A');
    expect(
      getByText('Output Branch').parentElement?.nextElementSibling,
    ).toHaveTextContent('master');
    expect(
      getByText('Egress').parentElement?.nextElementSibling,
    ).toHaveTextContent('Yes');
    expect(
      getByText('S3 Output Repo').parentElement?.nextElementSibling,
    ).toHaveTextContent(`s3//${pipelineId}`);
  });
});
