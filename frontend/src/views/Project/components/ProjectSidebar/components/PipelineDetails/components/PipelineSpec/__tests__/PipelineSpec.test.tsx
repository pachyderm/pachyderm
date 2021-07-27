import {render, waitForElementToBeRemoved} from '@testing-library/react';
import React from 'react';
import {Route} from 'react-router';

import {withContextProviders} from '@dash-frontend/testHelpers';
import {PIPELINE_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {pipelineRoute} from '@dash-frontend/views/Project/utils/routes';

import PipelineSpecComponent from '..';

describe('PipelineSpec', () => {
  const PipelineSpec = withContextProviders(() => (
    <Route path={PIPELINE_PATH} component={PipelineSpecComponent} />
  ));

  it('should correctly render JSON spec', async () => {
    window.history.replaceState(
      '',
      '',
      pipelineRoute({projectId: '1', pipelineId: 'montage'}),
    );

    const {container, queryByTestId} = render(<PipelineSpec />);

    await waitForElementToBeRemoved(queryByTestId('PipelineSpec__loader'));

    expect(container.firstChild).toMatchSnapshot();
  });
});
