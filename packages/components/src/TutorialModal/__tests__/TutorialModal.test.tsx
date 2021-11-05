import {render} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React, {useEffect} from 'react';

import {click} from 'testHelpers';
import {Step, TaskComponentProps} from 'TutorialModal/lib/types';

import {Link} from '../../Link';
import Mark from '../components/Mark';
import TaskCard from '../components/TaskCard';
import TutorialModal from '../TutorialModal';

const Task1Component: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  index,
  name,
}) => {
  return (
    <TaskCard
      index={index}
      name={name}
      action={onCompleted}
      currentTask={currentTask}
      actionText="Create pipeline spec"
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
        name="Minimize the overlay and inspect the pipeline and resulting output repo in the DAG"
      />
    </>
  );
};

const steps: Step[] = [
  {
    label: 'Introduction',
    name: 'Creating Pipelines',
    tasks: [
      {
        name: (
          <>
            Create a pipeline using the provided <Mark>edges.json</Mark>{' '}
            specification.
          </>
        ),
        info: {
          name: 'Quickly define pipelines',
          text: [
            'Pachyderm pipelines can be defined in just a few lines of JSON or YAML specification. Pachyderm defines pipelines using docker containers tags, a command to be executed, and an input repo.',
          ],
        },
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
        name:
          'Minimize the overlay and inspect the pipeline and resulting output repo in the DAG',
        Task: Task2Component,
        followUp:
          'The pipeline spec contains a few simple sections. The pipeline section contains a name, which is how you will identify your pipeline. Your pipeline will also automatically create an output repo with the same name. The transform section allows you to specify the docker image you want to use. In this case, pachyderm/opencv is the docker image (defaults to DockerHub as the registry), and the entry point is',
      },
    ],
    instructionsHeader: 'Creating pipelines with just a few lines of code',
    instructionsText: (
      <>
        <p>
          Pipelines are the core processing primitive in Pachyderm. Pipelines
          are defined with a simple JSON file called a pipeline specification or
          pipeline spec for short. For this example, we already created the
          pipeline spec for you. When you want to create your own pipeline specs
          later, you can refer to the full{' '}
          <Link externalLink inline to="foo">
            Pipeline Specification
          </Link>{' '}
          to use more advanced options. Options include building your own code
          into a container. In this tutorial, we are using a pre-built Docker
          image.
        </p>
        <p>
          For now, we are going to create a single pipeline spec that takes in
          images and does some simple edge detection.
        </p>
      </>
    ),
  },
];

describe('TutorialModal', () => {
  it('should disable moving to the next step if all tasks are not complete', async () => {
    const {findByRole} = render(<TutorialModal steps={steps} />);

    const nextButton = await findByRole('button', {
      name: 'Next Story',
    });
    expect(nextButton).toBeDisabled();
  });

  it('should disable moving to the next step if the user is on the last step', async () => {
    const {findByRole} = render(<TutorialModal steps={steps} />);

    const pipelineButton = await findByRole('button', {
      name: 'Create pipeline spec',
    });
    click(pipelineButton);

    const minimizeButton = await findByRole('button', {name: 'Minimize'});
    click(minimizeButton);

    const nextButton = await findByRole('button', {
      name: 'Next Story',
    });
    expect(nextButton).toBeDisabled();
  });

  it('should allow moving to the next step', async () => {
    const nextStep: Step = {
      name: 'The next step',
      tasks: [],
      instructionsHeader: '',
      instructionsText: '',
    };

    const {findByRole, findByText} = render(
      <TutorialModal steps={steps.concat([nextStep])} />,
    );

    expect(await findByText('Introduction')).toBeInTheDocument();

    const pipelineButton = await findByRole('button', {
      name: 'Create pipeline spec',
    });
    click(pipelineButton);

    const minimizeButton = await findByRole('button', {name: 'Minimize'});
    click(minimizeButton);

    const nextButton = await findByRole('button', {
      name: 'Next Story',
    });
    userEvent.click(nextButton);

    expect(await findByText('Step 1')).toBeInTheDocument();
  });

  it('should update the maximize/minimize button text', async () => {
    const {findByRole} = render(<TutorialModal steps={steps} />);

    const minimizeButton = await findByRole('button', {name: 'Minimize'});

    click(minimizeButton);

    expect(await findByRole('button', {name: 'Maximize'})).toBeInTheDocument();
  });

  it('should display info for the current task in the side bar', async () => {
    const {findByText} = render(<TutorialModal steps={steps} />);

    expect(await findByText('Quickly define pipelines')).toBeInTheDocument();
  });

  it('should not mark a task completed unless the previous tasks are completed', async () => {
    const {findByRole, findAllByLabelText} = render(
      <TutorialModal steps={steps} />,
    );

    const sizeButton = await findByRole('button', {name: 'Minimize'});

    click(sizeButton);
    click(sizeButton);

    const task2Checkboxes = await findAllByLabelText(
      'Minimize the overlay and inspect the pipeline and resulting output repo in the DAG',
    );
    task2Checkboxes.forEach((checkbox) => expect(checkbox).not.toBeChecked());

    const pipelineButton = await findByRole('button', {
      name: 'Create pipeline spec',
    });
    click(pipelineButton);

    const task1Checkboxes = await findAllByLabelText(
      'Create a pipeline using the provided edges.json specification.',
    );
    task1Checkboxes.forEach((checkbox) => expect(checkbox).toBeChecked());

    click(sizeButton);
    click(sizeButton);

    const updatedTask2Checkboxes = await findAllByLabelText(
      'Minimize the overlay and inspect the pipeline and resulting output repo in the DAG',
    );

    updatedTask2Checkboxes.forEach((checkbox) =>
      expect(checkbox).toBeChecked(),
    );
  });

  it('should display the current task when minimized', async () => {
    const {findByRole, findAllByLabelText} = render(
      <TutorialModal steps={steps} />,
    );

    const minimizeButton = await findByRole('button', {name: 'Minimize'});
    click(minimizeButton);

    const tasks = await findAllByLabelText(
      'Create a pipeline using the provided edges.json specification.',
    );
    expect(tasks.length).toBe(3);
  });

  it('should include the continue step when displaying the final task while minimized', async () => {
    const {findByRole, findAllByLabelText} = render(
      <TutorialModal steps={steps} />,
    );

    const pipelineButton = await findByRole('button', {
      name: 'Create pipeline spec',
    });
    click(pipelineButton);

    const minimizeButton = await findByRole('button', {name: 'Minimize'});
    click(minimizeButton);

    const initialTasks = await findAllByLabelText(
      'Create a pipeline using the provided edges.json specification.',
    );
    expect(initialTasks.length).toBe(2);
    const finalTasks = await findAllByLabelText(
      'Minimize the overlay and inspect the pipeline and resulting output repo in the DAG',
    );
    expect(finalTasks.length).toBe(3);
    const completeSteps = await findAllByLabelText('Continue to next step.');
    expect(completeSteps.length).toBe(2);
  });
});
