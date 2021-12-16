import {mockServer} from '@dash-backend/testHelpers';
import {render, waitFor} from '@testing-library/react';
import React from 'react';

import {
  withContextProviders,
  click,
  generateTutorialView,
} from '@dash-frontend/testHelpers';

import ProjectTutorial from '../../ProjectTutorial';

describe('Image Processing', () => {
  beforeEach(() => {
    window.history.replaceState(
      {},
      '',
      `/project/6?view=${generateTutorialView('image-processing')}`,
    );
  });
  const Tutorial = withContextProviders(() => {
    return <ProjectTutorial />;
  });

  it('should start the tutorial from the url param', async () => {
    const {findByText} = render(<Tutorial />);
    const tutorialTitle = await findByText('Create a pipeline');
    expect(tutorialTitle).toBeInTheDocument();
  });

  it('should allow the user to create the images repo', async () => {
    const {findByRole, findByText} = render(<Tutorial />);

    expect(mockServer.getState().repos['6']).toHaveLength(0);

    const repoCreationButton = await findByRole('button', {
      name: 'Create the images repo',
    });

    click(repoCreationButton);

    expect(await findByText('Task Completed!')).toBeInTheDocument();

    expect(mockServer.getState().repos['6']).toHaveLength(1);
  });

  it('should allow the user to create the edges Pipeline', async () => {
    const {findByRole, findByText, findAllByRole} = render(<Tutorial />);

    expect(mockServer.getState().repos['6']).toHaveLength(0);
    expect(mockServer.getState().pipelines['6']).toHaveLength(0);

    const repoCreationButton = await findByRole('button', {
      name: 'Create the images repo',
    });

    click(repoCreationButton);

    expect(await findByText('Task Completed!')).toBeInTheDocument();

    expect(mockServer.getState().repos['6']).toHaveLength(1);

    const pipelineCreationButton = (
      await findAllByRole('button', {
        name: 'Create the edges pipeline',
      })
    )[1];

    click(pipelineCreationButton);
    const nextStoryButton = (
      await findAllByRole('button', {name: 'Next Story'})
    )[1];

    await waitFor(() => expect(nextStoryButton).not.toBeDisabled());
    expect(mockServer.getState().repos['6']).toHaveLength(2);
    expect(mockServer.getState().pipelines['6']).toHaveLength(1);
  });
});
