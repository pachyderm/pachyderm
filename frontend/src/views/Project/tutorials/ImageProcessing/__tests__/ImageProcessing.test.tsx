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

  it('should allow the user to complete the tutorial', async () => {
    const {findByRole, findByText, findAllByRole, findByLabelText} = render(
      <Tutorial />,
    );

    const minimize = async () => {
      const minimizeButton = (
        await findAllByRole('button', {
          name: 'Minimize',
        })
      )[0];

      click(minimizeButton);
    };

    const maximize = async () => {
      const maximizeButton = (
        await findAllByRole('button', {
          name: 'Maximize',
        })
      )[0];
      click(maximizeButton);
    };

    const nextStory = async () => {
      const nextStoryButton = (
        await findAllByRole('button', {name: 'Next Story'})
      )[1];

      await waitFor(() => expect(nextStoryButton).not.toBeDisabled());
      click(nextStoryButton);
    };

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

    await nextStory();

    expect(
      await findByText('Add files to the images repo you created'),
    ).toBeInTheDocument();

    expect(mockServer.getState().files['6']).toBeUndefined();
    const checkbox1 = await findByLabelText('birthday-cake.jpg');
    click(checkbox1);

    const imageUploadButton1 = await findByRole('button', {
      name: 'Add these images',
    });

    click(imageUploadButton1);
    expect(await findByText('Task Completed!')).toBeInTheDocument();
    expect(mockServer.getState().files['6']['/']).toHaveLength(1);

    await minimize();
    await maximize();

    await nextStory();

    expect(
      await findByText(
        'Add a pipeline called "montage" that uses "edges" as its input',
      ),
    ).toBeInTheDocument();

    const montagePipelineCreationButton = (
      await findAllByRole('button', {
        name: 'Create the montage pipeline',
      })
    )[0];

    click(montagePipelineCreationButton);
    expect(await findByText('Task Completed!')).toBeInTheDocument();
    expect(mockServer.getState().pipelines['6']).toHaveLength(2);

    await minimize();
    await maximize();

    await nextStory();

    expect(
      await findByText('Global identifiers tie together commits and code'),
    ).toBeInTheDocument();

    await minimize();
    await maximize();

    await nextStory();

    expect(
      await findByText('About Incremental Scalability'),
    ).toBeInTheDocument();

    const checkbox2 = await findByLabelText('pooh.jpg');
    click(checkbox2);

    const imageUploadButton2 = await findByRole('button', {
      name: 'Add these images',
    });

    click(imageUploadButton2);
    expect(await findByText('Task Completed!')).toBeInTheDocument();

    expect(mockServer.getState().files['6']['/']).toHaveLength(2);

    await minimize();
    await maximize();

    await nextStory();
  });
});
