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

    expect(getByText('Pipeline Status').nextElementSibling).toHaveTextContent(
      'Failure',
    );
    expect(getByText('Pipeline Type').nextElementSibling).toHaveTextContent(
      'Standard',
    );
    expect(getByText('Description').nextElementSibling).toHaveTextContent(
      'Not my favorite pipeline',
    );

    const outputRepo = getByText('Output Repo').nextElementSibling;
    expect(outputRepo).toHaveTextContent(pipelineId);
    expect(within(outputRepo as HTMLElement).getByRole('link')).toHaveAttribute(
      'href',
      `/project/${projectId}/repos/${pipelineId}/branch/master`,
    );

    expect(getByText('Datum Timeout').nextElementSibling).toHaveTextContent(
      'N/A',
    );
    expect(getByText('Datum Tries').nextElementSibling).toHaveTextContent('0');
    expect(getByText('Job Timeout').nextElementSibling).toHaveTextContent(
      'N/A',
    );
    expect(getByText('Output Branch').nextElementSibling).toHaveTextContent(
      'master',
    );
    expect(getByText('Egress').nextElementSibling).toHaveTextContent('Yes');
    expect(getByText('S3 Output Repo').nextElementSibling).toHaveTextContent(
      `s3//${pipelineId}`,
    );
  });
});
