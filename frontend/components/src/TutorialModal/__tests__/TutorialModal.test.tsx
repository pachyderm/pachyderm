import {render, screen} from '@testing-library/react';
import React, {useEffect} from 'react';

import {click} from '@dash-frontend/testHelpers';
import {
  Story,
  TaskComponentProps,
} from '@pachyderm/components/TutorialModal/lib/types';

import {Link} from '../../Link';
import Mark from '../components/Mark';
import TaskCard from '../components/TaskCard';
import TutorialModalBodyProvider from '../components/TutorialModalBody/TutorialModalBodyProvider';
import TutorialModalComponent from '../TutorialModal';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const TutorialModal = (props: any): any => {
  return (
    <TutorialModalBodyProvider>
      <TutorialModalComponent {...props} />
    </TutorialModalBodyProvider>
  );
};

const Task1Component: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  index,
  name,
}) => {
  return (
    <TaskCard
      index={index}
      task={name}
      action={onCompleted}
      currentTask={currentTask}
      actionText="Create pipeline spec"
      taskInfoTitle="Quickly define pipelines"
      taskInfo="Pachyderm pipelines can be defined in just a few lines of JSON or YAML specification. Pachyderm defines pipelines using docker containers tags, a command to be executed, and an input repo."
    />
  );
};

const Task2Component: React.FC<TaskComponentProps> = ({
  currentTask,
  onCompleted,
  minimized,
  index,
}) => {
  useEffect(() => {
    if (currentTask === index && minimized) {
      onCompleted();
    }
  }, [currentTask, minimized, onCompleted, index]);

  return (
    <>
      <TaskCard
        index={index}
        currentTask={currentTask}
        task="Minimize the overlay and inspect the pipeline and resulting output repo in the DAG"
        taskInfoTitle="Automatically track data lineage"
        taskInfo="Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book."
      />
    </>
  );
};

const stories: Story[] = [
  {
    name: 'Creating Pipelines',
    sections: [
      {
        header: 'Creating pipelines with just a few lines of code',
        info: (
          <>
            <p>
              Pipelines are the core processing primitive in Pachyderm.
              Pipelines are defined with a simple JSON file called a pipeline
              specification or pipeline spec for short. For this example, we
              already created the pipeline spec for you. When you want to create
              your own pipeline specs later, you can refer to the full{' '}
              <Link externalLink inline to="foo">
                Pipeline Specification
              </Link>{' '}
              to use more advanced options. Options include building your own
              code into a container. In this tutorial, we are using a pre-built
              Docker image.
            </p>
            <span>
              For now, we are going to create a single pipeline spec that takes
              in images and does some simple edge detection.
            </span>
          </>
        ),
        taskName: (
          <span>
            Create a pipeline using the provided <Mark>edges.json</Mark>{' '}
            specification.
          </span>
        ),
        Task: Task1Component,
        followUp: (
          <p>
            The pipeline spec contains a few simple sections. The pipeline
            section contains a name, which is how you will identify your
            pipeline. Your pipeline will also automatically create an output
            repo with the same name. The transform section allows you to specify
            the docker image you want to use. In this case, pachyderm/opencv is
            the docker image (defaults to DockerHub as the registry), and the
            entry point is <Mark>edges.py</Mark> The input section specifies
            repos visible to the running pipeline.
          </p>
        ),
      },
      {
        header: 'Section without Task',
        info: 'This is a section without a task.',
      },
      {
        isSubHeader: true,
        header: 'Viewing commits in Console',
        info: (
          <p>
            Console provides // Lorem Ipsum is simply dummy text of the printing
            and typesetting industry. Lorem Ipsum has been the industry&apos;s
            standard dummy text ever since the 1500s, when an unknown printer
            took a galley of type and scrambled it to make a type specimen book.
            It has survived not only five centuries, but also the leap into
            electronic typesetting, remaining essentially unchanged. It was
            popularised in the 1960s with the release of Letraset sheets
            containing Lorem Ipsum passages, and more recently with desktop
            publishing software like Aldus PageMaker including versions of Lorem
            Ipsum.
          </p>
        ),
        taskName:
          'Minimize the overlay and inspect the pipeline and resulting output repo in the DAG',
        Task: Task2Component,
        followUp:
          'The pipeline spec contains a few simple sections. The pipeline section contains a name, which is how you will identify your pipeline. Your pipeline will also automatically create an output repo with the same name. The transform section allows you to specify the docker image you want to use. In this case, pachyderm/opencv is the docker image (defaults to DockerHub as the registry), and the entry point is',
      },
    ],
  },
  {
    name: 'The next story',
    sections: [
      {
        taskName: 'The next task',
        Task: Task1Component,
      },
    ],
  },
];

describe('TutorialModal', () => {
  it('should disable moving to the next story if all tasks are not complete', async () => {
    render(<TutorialModal stories={stories} tutorialName="test" />);

    const nextButtons = await screen.findAllByRole('button', {
      name: 'Next Story',
    });
    nextButtons.map((button) => expect(button).toBeDisabled());
  });

  it('should allow moving to the next story', async () => {
    render(<TutorialModal stories={stories} tutorialName="test" />);

    expect(await screen.findByText('Story 1 of 2')).toBeInTheDocument();

    const pipelineButton = await screen.findByRole('button', {
      name: 'Create pipeline spec',
    });
    await click(pipelineButton);

    const minimizeButton = await screen.findByRole('button', {
      name: 'minimize',
    });
    await click(minimizeButton);

    const nextButtons = await screen.findAllByRole('button', {
      name: 'Next Story',
    });

    await click(nextButtons[0]);

    expect(await screen.findByText('Story 2 of 2')).toBeInTheDocument();
  });

  it('should update the maximize/minimize button text', async () => {
    render(<TutorialModal stories={stories} tutorialName="test" />);

    const minimizeButton = await screen.findByRole('button', {
      name: 'minimize',
    });

    await click(minimizeButton);

    expect(
      await screen.findByRole('button', {name: 'maximize'}),
    ).toBeInTheDocument();
  });

  it('should display info for the current task in the side bar', async () => {
    render(<TutorialModal stories={stories} tutorialName="test" />);

    expect(
      await screen.findByText('Quickly define pipelines'),
    ).toBeInTheDocument();
  });

  it('should not mark a task completed unless the previous tasks are completed', async () => {
    render(<TutorialModal stories={stories} tutorialName="test" />);

    const sizeButton = await screen.findByRole('button', {name: 'minimize'});

    await click(sizeButton);
    await click(sizeButton);

    const task2Checkmark = screen.queryByLabelText('Task 2 complete');
    expect(task2Checkmark).not.toBeInTheDocument();

    const pipelineButton = await screen.findByRole('button', {
      name: 'Create pipeline spec',
    });
    await click(pipelineButton);

    const task1Checkmark = screen.queryByLabelText('Task 1 complete');
    expect(task1Checkmark).toBeInTheDocument();

    await click(sizeButton);
    await click(sizeButton);

    const updatedTask2Checkmark = screen.queryByLabelText('Task 2 complete');
    expect(updatedTask2Checkmark).toBeInTheDocument();
  });

  it('should be able to switch between stories using the dropdown', async () => {
    render(<TutorialModal stories={stories} tutorialName="test" />);

    expect(screen.getByText('Story 1 of 2')).toBeInTheDocument();

    const dropdownButton = screen.getByRole('button', {
      name: 'Creating Pipelines',
    });

    await click(dropdownButton);

    await click(screen.getByText('The next story'));
    expect(screen.getByText('Story 2 of 2')).toBeInTheDocument();
  });

  it("should invoke onSkip callback when user clicks the 'Skip Tutorial' button", async () => {
    const handleLeave = jest.fn();
    render(
      <TutorialModal
        stories={stories}
        onSkip={handleLeave}
        tutorialName="test"
      />,
    );

    const skipButton = await screen.findByTestId(
      'TutorialModalBody__skipTutorial',
    );
    await click(skipButton);

    expect(handleLeave).toHaveBeenCalled();
  });

  it("should invoke onClose callback when user clicks the 'End Tutorial' button", async () => {
    const handleLeave = jest.fn();
    render(
      <TutorialModal
        stories={stories}
        onClose={handleLeave}
        tutorialName="test"
      />,
    );

    const closeButton = await screen.findByTestId(
      'TutorialModalBody__closeTutorial',
    );
    await click(closeButton);

    expect(handleLeave).toHaveBeenCalled();
  });
});
