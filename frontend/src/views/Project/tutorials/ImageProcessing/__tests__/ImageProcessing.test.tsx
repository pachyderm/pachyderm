import {mockServer} from '@dash-backend/testHelpers';
import {render, waitFor, screen} from '@testing-library/react';
import React from 'react';

import {
  withContextProviders,
  click,
  generateTutorialView,
} from '@dash-frontend/testHelpers';
import {TutorialModalBodyProvider} from '@pachyderm/components';

import ProjectTutorialComponent from '../../ProjectTutorial';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const ProjectTutorial = (props: any): any => {
  return (
    <TutorialModalBodyProvider>
      <ProjectTutorialComponent {...props} />
    </TutorialModalBodyProvider>
  );
};

describe('Image Processing', () => {
  beforeEach(() => {
    window.history.replaceState(
      {},
      '',
      `/lineage/Empty-Project?view=${generateTutorialView('image-processing')}`,
    );
  });

  const Tutorial = withContextProviders(() => {
    return <ProjectTutorial />;
  });

  it('should start the tutorial from the url param', async () => {
    render(<Tutorial />);
    const tutorialTitle = await screen.findByText('Create a pipeline');
    expect(tutorialTitle).toBeInTheDocument();
  });

  it('should allow the user to pause the tutorial', async () => {
    const {container} = render(<Tutorial />);

    await click(await screen.findByTestId('TutorialModalBody__closeTutorial'));

    await waitFor(() => expect(container).toBeEmptyDOMElement());
  });

  it('should allow the user to complete the tutorial', async () => {
    const {container} = render(<Tutorial />);

    const minimize = async () => {
      const minimizeButton = await screen.findByTestId(
        'TutorialModalBody__minimize',
      );

      await click(minimizeButton);
    };

    const maximize = async () => {
      const maximizeButton = await screen.findByTestId(
        'TutorialModalBody__maximize',
      );
      await click(maximizeButton);
    };

    const nextStory = async () => {
      const nextStoryButton = await screen.findByRole('button', {
        name: 'Next Story',
      });

      await waitFor(() => expect(nextStoryButton).toBeEnabled());
      await click(nextStoryButton);
    };

    const addTheseImages = async () => {
      const imageUploadButton = await screen.findByRole('button', {
        name: 'Add these images',
      });

      await click(imageUploadButton);

      return waitFor(() => expect(imageUploadButton).not.toBeInTheDocument());
    };

    expect(mockServer.getState().repos['Empty-Project']).toHaveLength(0);
    expect(mockServer.getState().pipelines['Empty-Project']).toHaveLength(0);

    const repoCreationButton = await screen.findByRole('button', {
      name: 'Create the images repo',
    });

    await waitFor(() => expect(repoCreationButton).toBeEnabled());
    await click(repoCreationButton);

    expect(await screen.findByText('Task Completed!')).toBeInTheDocument();

    expect(mockServer.getState().repos['Empty-Project']).toHaveLength(1);

    const pipelineCreationButton = (
      await screen.findAllByRole('button', {
        name: 'Create the edges pipeline',
      })
    )[1];

    await click(pipelineCreationButton);
    const nextStoryButton = await screen.findByRole('button', {
      name: 'Next Story',
    });

    await waitFor(() => expect(nextStoryButton).toBeEnabled());

    expect(mockServer.getState().repos['Empty-Project']).toHaveLength(2);
    expect(mockServer.getState().pipelines['Empty-Project']).toHaveLength(1);

    await nextStory();

    expect(
      await screen.findByText('Add files to the images repo you created'),
    ).toBeInTheDocument();

    expect(mockServer.getState().files['Empty-Project']).toBeUndefined();
    const checkbox1 = await screen.findByLabelText('birthday-cake.jpg');
    await click(checkbox1);

    await addTheseImages();

    expect(await screen.findByText('Task Completed!')).toBeInTheDocument();
    expect(mockServer.getState().files['Empty-Project']['/']).toHaveLength(1);

    await minimize();
    await maximize();

    await nextStory();

    expect(
      await screen.findByText(
        'Add a pipeline called "montage" that uses "edges" as its input',
      ),
    ).toBeInTheDocument();

    const montagePipelineCreationButton = (
      await screen.findAllByRole('button', {
        name: 'Create the montage pipeline',
      })
    )[0];

    await click(montagePipelineCreationButton);
    expect(await screen.findByText('Task Completed!')).toBeInTheDocument();
    expect(mockServer.getState().pipelines['Empty-Project']).toHaveLength(2);

    await minimize();
    await maximize();

    await nextStory();

    expect(
      await screen.findByText(
        'Global identifiers tie together commits and code',
      ),
    ).toBeInTheDocument();

    await minimize();
    await maximize();

    await nextStory();

    // expect(
    //   await screen.findByText('Basic reproducibility concepts'),
    // ).toBeInTheDocument();

    // await minimize();
    // await maximize();

    // const kitten = await screen.findByLabelText('kitten.jpg');
    //await click(kitten);

    // await addTheseImages();

    // expect(mockServer.getState().files['Empty-Project']['/']).toHaveLength(2);

    // await minimize();
    // await maximize();

    // const moveBranchButton = await screen.findByRole('button', {
    //   name: 'Move images branch',
    // });

    //await click(moveBranchButton);

    // await waitFor(() => expect(moveBranchButton).not.toBeInTheDocument());

    // expect(
    //   await screen.findByText(
    //     "Confirm that the montage's original version is restored",
    //   ),
    // ).toBeInTheDocument();

    // await minimize();
    // await maximize();

    // await nextStory();

    expect(
      await screen.findByText('About Incremental Scalability'),
    ).toBeInTheDocument();

    const checkbox2 = await screen.findByLabelText('pooh.jpg');
    await click(checkbox2);

    await addTheseImages();

    expect(await screen.findByText('Task Completed!')).toBeInTheDocument();

    // expect(mockServer.getState().files['Empty-Project']['/']).toHaveLength(3);
    expect(mockServer.getState().files['Empty-Project']['/']).toHaveLength(2);

    await minimize();
    await maximize();

    const completeStoryButton = await screen.findByRole('button', {
      name: 'Close Tutorial',
    });

    await click(completeStoryButton);

    await waitFor(() => expect(container).toBeEmptyDOMElement());
  });
});
