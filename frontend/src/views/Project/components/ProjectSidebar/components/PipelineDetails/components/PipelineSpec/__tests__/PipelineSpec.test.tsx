import {
  render,
  waitForElementToBeRemoved,
  screen,
  waitFor,
} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';
import {Route} from 'react-router';

import {
  mockEmptyGetAuthorize,
  mockGetMontagePipeline,
} from '@dash-frontend/mocks';
import {withContextProviders} from '@dash-frontend/testHelpers';
import {LINEAGE_PIPELINE_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {pipelineRoute} from '@dash-frontend/views/Project/utils/routes';

import PipelineSpecComponent from '..';

describe('PipelineSpec', () => {
  const server = setupServer();

  const PipelineSpec = withContextProviders(() => (
    <Route path={LINEAGE_PIPELINE_PATH} component={PipelineSpecComponent} />
  ));

  beforeAll(() => {
    server.listen();
    server.use(mockEmptyGetAuthorize());
    server.use(mockGetMontagePipeline());
  });

  afterAll(() => server.close());

  it('should correctly render JSON spec', async () => {
    window.history.replaceState(
      '',
      '',
      pipelineRoute({
        projectId: 'default',
        pipelineId: 'montage',
      }),
    );

    render(<PipelineSpec />);

    await waitForElementToBeRemoved(() => screen.queryByRole('status'));

    const codeSpec = await screen.findByTestId(
      'ConfigFilePreview__codeElement',
    );

    await waitFor(() => {
      expect(document.querySelectorAll('.cm-cursor-primary')).toHaveLength(1);
    }); // wait for cursor to appear

    expect(codeSpec).toMatchSnapshot();
  });
});
