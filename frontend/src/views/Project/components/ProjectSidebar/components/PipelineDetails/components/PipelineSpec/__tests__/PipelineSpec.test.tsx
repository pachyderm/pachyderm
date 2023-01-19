import {
  render,
  waitForElementToBeRemoved,
  screen,
} from '@testing-library/react';
import React from 'react';
import {Route} from 'react-router';

import {withContextProviders} from '@dash-frontend/testHelpers';
import {PROJECT_PIPELINE_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {pipelineRoute} from '@dash-frontend/views/Project/utils/routes';

import PipelineSpecComponent from '..';

describe('PipelineSpec', () => {
  const PipelineSpec = withContextProviders(() => (
    <Route path={PROJECT_PIPELINE_PATH} component={PipelineSpecComponent} />
  ));

  it('should correctly render JSON spec', async () => {
    window.history.replaceState(
      '',
      '',
      pipelineRoute({
        projectId: 'Solar-Panel-Data-Sorting',
        pipelineId: 'montage',
      }),
    );

    const {container} = render(<PipelineSpec />);

    await waitForElementToBeRemoved(
      screen.queryByTestId('PipelineSpec__loader'),
    );

    expect(container.firstChild).toMatchSnapshot();
  });
});
