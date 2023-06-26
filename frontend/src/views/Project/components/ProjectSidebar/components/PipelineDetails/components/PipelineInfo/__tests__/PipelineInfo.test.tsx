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
    expect(screen.getByLabelText('Pipeline Type')).toHaveTextContent(
      'Standard',
    );
    expect(screen.getByLabelText('Failure Reason')).toHaveTextContent('failed');
    expect(screen.getByLabelText('Description')).toHaveTextContent(
      'Not my favorite pipeline',
    );

    const outputRepo = screen.getByLabelText('Output Repo');
    expect(outputRepo).toHaveTextContent(pipelineId);
    expect(within(outputRepo as HTMLElement).getByRole('link')).toHaveAttribute(
      'href',
      `/lineage/${projectId}/repos/${pipelineId}`,
    );

    expect(screen.getByLabelText('Datum Timeout')).toHaveTextContent('N/A');
    expect(screen.getByLabelText('Datum Tries')).toHaveTextContent('0');
    expect(screen.getByLabelText('Job Timeout')).toHaveTextContent('N/A');
    expect(screen.getByLabelText('Output Branch')).toHaveTextContent('master');
    expect(screen.getByLabelText('Egress')).toHaveTextContent('Yes');
    expect(screen.getByLabelText('S3 Output Repo')).toHaveTextContent(
      `s3//${pipelineId}`,
    );
  });
});
